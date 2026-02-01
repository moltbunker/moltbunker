package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/moltbunker/moltbunker/internal/util"
)

// LogManager manages container log files
type LogManager struct {
	logsDir     string
	logFiles    map[string]*ContainerLog
	mu          sync.RWMutex
	maxFileSize int64 // Max size per log file before rotation
	maxFiles    int   // Max number of rotated files to keep
}

// ContainerLog represents a container's log files
type ContainerLog struct {
	ContainerID string
	StdoutPath  string
	StderrPath  string
	StdoutFile  *os.File
	StderrFile  *os.File
	mu          sync.Mutex
}

// NewLogManager creates a new log manager
func NewLogManager(logsDir string) (*LogManager, error) {
	if err := os.MkdirAll(logsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	return &LogManager{
		logsDir:     logsDir,
		logFiles:    make(map[string]*ContainerLog),
		maxFileSize: 100 * 1024 * 1024, // 100MB
		maxFiles:    5,
	}, nil
}

// CreateLog creates log files for a container
func (lm *LogManager) CreateLog(containerID string) (*ContainerLog, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	containerDir := filepath.Join(lm.logsDir, containerID)
	if err := os.MkdirAll(containerDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create container log directory: %w", err)
	}

	stdoutPath := filepath.Join(containerDir, "stdout.log")
	stderrPath := filepath.Join(containerDir, "stderr.log")

	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout log: %w", err)
	}

	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		stdoutFile.Close()
		return nil, fmt.Errorf("failed to create stderr log: %w", err)
	}

	log := &ContainerLog{
		ContainerID: containerID,
		StdoutPath:  stdoutPath,
		StderrPath:  stderrPath,
		StdoutFile:  stdoutFile,
		StderrFile:  stderrFile,
	}

	lm.logFiles[containerID] = log
	return log, nil
}

// GetLog retrieves a container's log
func (lm *LogManager) GetLog(containerID string) (*ContainerLog, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	log, exists := lm.logFiles[containerID]
	return log, exists
}

// CloseLog closes a container's log files
func (lm *LogManager) CloseLog(containerID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	log, exists := lm.logFiles[containerID]
	if !exists {
		return nil
	}

	log.mu.Lock()
	defer log.mu.Unlock()

	if log.StdoutFile != nil {
		log.StdoutFile.Close()
	}
	if log.StderrFile != nil {
		log.StderrFile.Close()
	}

	delete(lm.logFiles, containerID)
	return nil
}

// DeleteLog deletes a container's log files
func (lm *LogManager) DeleteLog(containerID string) error {
	lm.CloseLog(containerID)

	containerDir := filepath.Join(lm.logsDir, containerID)
	return os.RemoveAll(containerDir)
}

// WriteStdout writes to a container's stdout log
func (cl *ContainerLog) WriteStdout(data []byte) (int, error) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.StdoutFile == nil {
		return 0, fmt.Errorf("stdout log not open")
	}

	// Add timestamp prefix
	timestamp := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %s", timestamp, string(data))

	return cl.StdoutFile.WriteString(line)
}

// WriteStderr writes to a container's stderr log
func (cl *ContainerLog) WriteStderr(data []byte) (int, error) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.StderrFile == nil {
		return 0, fmt.Errorf("stderr log not open")
	}

	// Add timestamp prefix
	timestamp := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %s", timestamp, string(data))

	return cl.StderrFile.WriteString(line)
}

// StdoutWriter returns a writer for stdout
func (cl *ContainerLog) StdoutWriter() io.Writer {
	return &logWriter{log: cl, stream: "stdout"}
}

// StderrWriter returns a writer for stderr
func (cl *ContainerLog) StderrWriter() io.Writer {
	return &logWriter{log: cl, stream: "stderr"}
}

// logWriter implements io.Writer for logging
type logWriter struct {
	log    *ContainerLog
	stream string
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	if lw.stream == "stdout" {
		return lw.log.WriteStdout(p)
	}
	return lw.log.WriteStderr(p)
}

// ReadLogs reads logs from a container
func (lm *LogManager) ReadLogs(ctx context.Context, containerID string, follow bool, tail int) (io.ReadCloser, error) {
	containerDir := filepath.Join(lm.logsDir, containerID)
	stdoutPath := filepath.Join(containerDir, "stdout.log")
	stderrPath := filepath.Join(containerDir, "stderr.log")

	// Check if logs exist
	if _, err := os.Stat(stdoutPath); os.IsNotExist(err) {
		// Return empty reader if no logs
		return io.NopCloser(emptyReader{}), nil
	}

	pr, pw := io.Pipe()

	util.SafeGoWithName("read-logs", func() {
		defer pw.Close()

		// Read existing logs
		if err := lm.readExistingLogs(pw, stdoutPath, stderrPath, tail); err != nil {
			pw.CloseWithError(err)
			return
		}

		// If following, continue reading new logs
		if follow {
			lm.followLogs(ctx, pw, stdoutPath, stderrPath)
		}
	})

	return pr, nil
}

// readExistingLogs reads existing log content
func (lm *LogManager) readExistingLogs(w io.Writer, stdoutPath, stderrPath string, tail int) error {
	// Read stdout
	if err := lm.readLogFile(w, stdoutPath, tail, "stdout"); err != nil {
		return err
	}

	// Read stderr
	if err := lm.readLogFile(w, stderrPath, tail, "stderr"); err != nil {
		return err
	}

	return nil
}

// readLogFile reads a log file with optional tail
func (lm *LogManager) readLogFile(w io.Writer, path string, tail int, prefix string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	if tail > 0 {
		// Read last N lines
		lines, err := readLastLines(file, tail)
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Fprintf(w, "[%s] %s\n", prefix, line)
		}
	} else {
		// Read entire file
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Fprintf(w, "[%s] %s\n", prefix, scanner.Text())
		}
		return scanner.Err()
	}

	return nil
}

// readLastLines reads the last N lines from a file
func readLastLines(file *os.File, n int) ([]string, error) {
	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	if size == 0 {
		return nil, nil
	}

	// Read from end of file
	var lines []string
	bufSize := int64(4096)
	offset := size

	for len(lines) < n && offset > 0 {
		readSize := bufSize
		if offset < bufSize {
			readSize = offset
		}
		offset -= readSize

		buf := make([]byte, readSize)
		_, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return nil, err
		}

		// Process buffer backwards
		for i := len(buf) - 1; i >= 0; i-- {
			if buf[i] == '\n' && i < len(buf)-1 {
				line := string(buf[i+1:])
				if line != "" {
					lines = append([]string{line}, lines...)
				}
				buf = buf[:i]
			}
		}

		// Handle remaining content
		if offset == 0 && len(buf) > 0 {
			lines = append([]string{string(buf)}, lines...)
		}
	}

	// Trim to N lines
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}

// followLogs follows log files for new content using fsnotify
func (lm *LogManager) followLogs(ctx context.Context, w io.Writer, stdoutPath, stderrPath string) {
	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// Fall back to polling if fsnotify fails
		lm.followLogsPoll(ctx, w, stdoutPath, stderrPath)
		return
	}
	defer watcher.Close()

	// Open files for following
	stdoutFile, _ := os.Open(stdoutPath)
	stderrFile, _ := os.Open(stderrPath)

	if stdoutFile != nil {
		stdoutFile.Seek(0, io.SeekEnd)
		defer stdoutFile.Close()
	}
	if stderrFile != nil {
		stderrFile.Seek(0, io.SeekEnd)
		defer stderrFile.Close()
	}

	// Watch the log directory for changes
	dir := filepath.Dir(stdoutPath)
	if err := watcher.Add(dir); err != nil {
		// Fall back to polling if watch fails
		lm.followLogsPoll(ctx, w, stdoutPath, stderrPath)
		return
	}

	buf := make([]byte, 4096)

	// Helper to read and output new content
	readNew := func(file *os.File, prefix string) {
		if file == nil {
			return
		}
		for {
			n, err := file.Read(buf)
			if err != nil || n == 0 {
				break
			}
			fmt.Fprintf(w, "[%s] %s", prefix, string(buf[:n]))
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Check if our log files were modified
			if event.Has(fsnotify.Write) {
				if event.Name == stdoutPath {
					readNew(stdoutFile, "stdout")
				} else if event.Name == stderrPath {
					readNew(stderrFile, "stderr")
				}
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
			// Ignore errors, continue watching
		}
	}
}

// followLogsPoll is the fallback polling implementation
func (lm *LogManager) followLogsPoll(ctx context.Context, w io.Writer, stdoutPath, stderrPath string) {
	// Open files for following
	stdoutFile, _ := os.Open(stdoutPath)
	stderrFile, _ := os.Open(stderrPath)

	if stdoutFile != nil {
		stdoutFile.Seek(0, io.SeekEnd)
		defer stdoutFile.Close()
	}
	if stderrFile != nil {
		stderrFile.Seek(0, io.SeekEnd)
		defer stderrFile.Close()
	}

	ticker := time.NewTicker(500 * time.Millisecond) // Increased from 100ms to reduce CPU
	defer ticker.Stop()

	buf := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Read new content from stdout
			if stdoutFile != nil {
				n, err := stdoutFile.Read(buf)
				if err == nil && n > 0 {
					fmt.Fprintf(w, "[stdout] %s", string(buf[:n]))
				}
			}

			// Read new content from stderr
			if stderrFile != nil {
				n, err := stderrFile.Read(buf)
				if err == nil && n > 0 {
					fmt.Fprintf(w, "[stderr] %s", string(buf[:n]))
				}
			}
		}
	}
}

// emptyReader is an empty io.Reader
type emptyReader struct{}

func (emptyReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// GetLogPaths returns the log file paths for a container
func (lm *LogManager) GetLogPaths(containerID string) (stdout, stderr string) {
	containerDir := filepath.Join(lm.logsDir, containerID)
	return filepath.Join(containerDir, "stdout.log"), filepath.Join(containerDir, "stderr.log")
}
