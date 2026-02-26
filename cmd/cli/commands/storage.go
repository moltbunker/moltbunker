package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

// NewStorageCmd creates the object storage command group.
func NewStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "storage",
		Aliases: []string{"s3", "store"},
		Short:   "Manage object storage buckets and objects",
		Long: `Manage S3-compatible encrypted object storage.

Objects are encrypted at rest with per-object AES-256-GCM keys and
distributed across the P2P network via IPFS.

Examples:
  moltbunker storage mb my-bucket              # Create a bucket
  moltbunker storage rb my-bucket              # Remove a bucket
  moltbunker storage ls                        # List buckets
  moltbunker storage ls my-bucket              # List objects
  moltbunker storage put my-bucket/file.txt local.txt   # Upload a file
  moltbunker storage get my-bucket/file.txt              # Download a file
  moltbunker storage rm my-bucket/file.txt               # Delete an object
  moltbunker storage usage                               # Show usage`,
	}

	cmd.AddCommand(
		newStorageMbCmd(),
		newStorageRbCmd(),
		newStorageLsCmd(),
		newStoragePutCmd(),
		newStorageGetCmd(),
		newStorageRmCmd(),
		newStorageUsageCmd(),
	)

	return cmd
}

func newStorageMbCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mb <bucket-name>",
		Short: "Create a new bucket",
		Args:  cobra.ExactArgs(1),
		RunE:  runStorageMb,
	}
}

func newStorageRbCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rb <bucket-name>",
		Short: "Remove a bucket (must be empty)",
		Args:  cobra.ExactArgs(1),
		RunE:  runStorageRb,
	}
}

func newStorageLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls [bucket[/prefix]]",
		Aliases: []string{"list"},
		Short:   "List buckets or objects",
		Args:    cobra.MaximumNArgs(1),
		RunE:    runStorageLs,
	}
}

func newStoragePutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "put <bucket/key> <local-file>",
		Short: "Upload a file to object storage",
		Args:  cobra.ExactArgs(2),
		RunE:  runStoragePut,
	}
}

func newStorageGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <bucket/key> [local-file]",
		Short: "Download an object to a local file",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runStorageGet,
	}
	return cmd
}

func newStorageRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <bucket/key>",
		Aliases: []string{"delete"},
		Short:   "Delete an object",
		Args:    cobra.ExactArgs(1),
		RunE:    runStorageRm,
	}
}

func newStorageUsageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Show storage usage for current wallet",
		Args:  cobra.NoArgs,
		RunE:  runStorageUsage,
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func runStorageMb(_ *cobra.Command, args []string) error {
	bucketName := args[0]

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	bucket, err := c.DaemonClient().StorageCreateBucket(bucketName)
	if err != nil {
		return fmt.Errorf("create bucket failed: %w", err)
	}

	Success(fmt.Sprintf("Created bucket: %s", bucket.Name))

	return nil
}

func runStorageRb(_ *cobra.Command, args []string) error {
	bucketName := args[0]

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	if err := c.DaemonClient().StorageDeleteBucket(bucketName); err != nil {
		return fmt.Errorf("delete bucket failed: %w", err)
	}

	Success(fmt.Sprintf("Deleted bucket: %s", bucketName))

	return nil
}

func runStorageLs(_ *cobra.Command, args []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	if len(args) == 0 {
		// List buckets
		buckets, err := c.DaemonClient().StorageListBuckets()
		if err != nil {
			return fmt.Errorf("list buckets failed: %w", err)
		}

		if len(buckets) == 0 {
			Info("No buckets")
			fmt.Println(Hint("Create one: moltbunker storage mb <name>"))
			return nil
		}

		headers := []string{"Name", "Created"}
		rows := make([][]string, 0, len(buckets))
		for _, b := range buckets {
			rows = append(rows, []string{
				b.Name,
				b.CreatedAt.Format("Jan 02 15:04"),
			})
		}
		fmt.Println(RenderTable(headers, rows))
		return nil
	}

	// List objects in bucket
	bucket, prefix := parseBucketPath(args[0])
	objects, err := c.DaemonClient().StorageListObjects(bucket, prefix)
	if err != nil {
		return fmt.Errorf("list objects failed: %w", err)
	}

	if len(objects) == 0 {
		Info(fmt.Sprintf("No objects in %s", args[0]))
		return nil
	}

	headers := []string{"Key", "Size", "Type", "Modified"}
	rows := make([][]string, 0, len(objects))
	for _, o := range objects {
		rows = append(rows, []string{
			o.Key,
			formatBytes(o.Size),
			o.ContentType,
			o.UpdatedAt.Format("Jan 02 15:04"),
		})
	}
	fmt.Println(RenderTable(headers, rows))

	return nil
}

func runStoragePut(_ *cobra.Command, args []string) error {
	bucket, key := parseBucketPath(args[0])
	localFile := args[1]

	if key == "" {
		// Use filename as key
		key = filepath.Base(localFile)
	}

	data, err := os.ReadFile(localFile)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	Info(fmt.Sprintf("Uploading %s (%s) to %s/%s", localFile, formatBytes(int64(len(data))), bucket, key))

	err = WithSpinner("Uploading", func() error {
		_, callErr := c.DaemonClient().StoragePutObject(&client.StoragePutObjectRequest{
			Bucket:      bucket,
			Key:         key,
			Data:        data,
			ContentType: detectContentType(localFile),
		})
		return callErr
	})
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	Success(fmt.Sprintf("Uploaded %s/%s", bucket, key))

	return nil
}

func runStorageGet(_ *cobra.Command, args []string) error {
	bucket, key := parseBucketPath(args[0])
	if key == "" {
		return fmt.Errorf("object key is required (use bucket/key format)")
	}

	// Determine output file
	outputFile := filepath.Base(key)
	if len(args) > 1 {
		outputFile = args[1]
	}

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	var resp *client.StorageGetObjectResponse
	err = WithSpinner("Downloading", func() error {
		var callErr error
		resp, callErr = c.DaemonClient().StorageGetObject(bucket, key)
		return callErr
	})
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if err := os.WriteFile(outputFile, resp.Data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	Success(fmt.Sprintf("Downloaded %s/%s → %s (%s)", bucket, key, outputFile, formatBytes(resp.Size)))

	return nil
}

func runStorageRm(_ *cobra.Command, args []string) error {
	bucket, key := parseBucketPath(args[0])
	if key == "" {
		return fmt.Errorf("object key is required (use bucket/key format)")
	}

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	if err := c.DaemonClient().StorageDeleteObject(bucket, key); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	Success(fmt.Sprintf("Deleted %s/%s", bucket, key))

	return nil
}

func runStorageUsage(_ *cobra.Command, _ []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	usage, err := c.DaemonClient().StorageUsage()
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	fields := [][2]string{
		{"Wallet", FormatAddress(usage.Wallet)},
		{"Buckets", fmt.Sprintf("%d", usage.BucketCount)},
		{"Objects", fmt.Sprintf("%d", usage.ObjectCount)},
		{"Total Size", formatBytes(usage.TotalBytes)},
	}

	fmt.Println(StatusBox("Storage Usage", fields))

	return nil
}

// parseBucketPath splits "bucket/key/path" into bucket and key components.
func parseBucketPath(path string) (bucket, key string) {
	bucket, key, _ = strings.Cut(path, "/")
	return bucket, key
}

// detectContentType guesses content type from file extension.
func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".txt", ".log":
		return "text/plain"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".gz", ".gzip":
		return "application/gzip"
	case ".tar":
		return "application/x-tar"
	case ".wasm":
		return "application/wasm"
	default:
		return "application/octet-stream"
	}
}
