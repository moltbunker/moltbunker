package doctor

import (
	"context"
	"fmt"
	"syscall"
)

// RecommendedFileDescriptors is the recommended minimum file descriptor soft limit for production
const RecommendedFileDescriptors uint64 = 65536

// FileDescriptorChecker checks the file descriptor soft limit
type FileDescriptorChecker struct{}

func NewFileDescriptorChecker() *FileDescriptorChecker {
	return &FileDescriptorChecker{}
}

func (c *FileDescriptorChecker) Name() string       { return "File descriptors" }
func (c *FileDescriptorChecker) Category() Category { return CategorySystem }
func (c *FileDescriptorChecker) CanFix() bool       { return false }

func (c *FileDescriptorChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Fixable:  false,
	}

	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		result.Status = StatusWarning
		result.Message = "File descriptors: Unable to check"
		result.Details = err.Error()
		return result
	}

	softLimit := rLimit.Cur

	if softLimit >= RecommendedFileDescriptors {
		result.Status = StatusOK
		result.Message = fmt.Sprintf("File descriptors: %d (>= %d recommended)", softLimit, RecommendedFileDescriptors)
	} else {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("File descriptors: %d (>= %d recommended for production)", softLimit, RecommendedFileDescriptors)
		result.Details = "Increase with 'ulimit -n 65536' or update /etc/security/limits.conf"
	}

	return result
}

func (c *FileDescriptorChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("file descriptor limits cannot be auto-fixed")
}
