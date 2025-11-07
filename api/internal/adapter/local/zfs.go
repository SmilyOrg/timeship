package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// ZFS implements the SnapshotProvider interface for ZFS filesystems
type ZFS struct {
	rootDir string
}

// NewZFS creates a new ZFS snapshot provider
func NewZFS(rootDir string) *ZFS {
	return &ZFS{
		rootDir: rootDir,
	}
}

// findSnapshotRoot traverses up from the given path looking for a .zfs directory
// Returns the path to the ZFS root (where .zfs/snapshot exists) or empty string if not found
func (z *ZFS) findSnapshotRoot(relPath string) (string, error) {
	currentPath := filepath.Join(z.rootDir, relPath)

	// Start from the given path and traverse up
	for {
		dir := filepath.Join(currentPath, ".zfs", "snapshot")
		stat, err := os.Stat(dir)
		if err == nil && stat.IsDir() {
			// Found it!
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// We've reached the root and didn't find .zfs
			break
		}
		currentPath = parent
	}

	// Not found
	return "", nil
}

// Snapshots returns all ZFS snapshots available for a given path
func (z *ZFS) Snapshots(relPath string) ([]adapter.Snapshot, error) {

	rootPath, err := z.findSnapshotRoot(relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to find snapshot root: %w", err)
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot dir: %w", err)
	}

	snapshots := []adapter.Snapshot{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Get file info to retrieve modification time
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't stat
		}

		// Parse the snapshot directory info to get timestamp
		// Format is typically like: auto-daily-2025-10-28_00-00, auto-hourly-2025-10-30_01-00, etc.
		modTime := info.ModTime()
		timestamp := modTime.Unix()

		snapshot := adapter.Snapshot{
			ID:        fmt.Sprintf("zfs:%s", entry.Name()),
			Type:      "zfs",
			Timestamp: timestamp,
			Name:      entry.Name(),
			Size:      -1, // ZFS snapshot size is not easily determinable
			Metadata: adapter.SnapshotMetadata{
				"zfs_root": rootPath,
			},
		}

		snapshots = append(snapshots, snapshot)
	}

	// Sort by timestamp in descending order (newest first)
	// sort.Slice(snapshots, func(i, j int) bool {
	// 	return snapshots[i].Timestamp > snapshots[j].Timestamp
	// })

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Name > snapshots[j].Name
	})

	return snapshots, nil
}

// getSnapshotPath extracts the snapshot path from the snapshot ID
// Input format: "zfs:snapshot-name"
// Returns just the "snapshot-name" part
func (z *ZFS) getSnapshotPath(snapshotID string) (string, error) {
	parts := strings.SplitN(snapshotID, ":", 2)
	if len(parts) != 2 || parts[0] != "zfs" {
		return "", fmt.Errorf("invalid snapshot ID format: %s", snapshotID)
	}
	return parts[1], nil
}

// OpenSnapshotRoot opens an os.Root for a snapshot, allowing safe traversal within it
// Returns the os.Root and the relative path within the snapshot
func (z *ZFS) SnapshotRoot(relPath string, snapshotID string) (*os.Root, error) {
	// Find the ZFS root
	rootPath, err := z.findSnapshotRoot(relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to find snapshot root: %w", err)
	}

	if rootPath == "" {
		return nil, fmt.Errorf("root path empty: %s", relPath)
	}

	// Get the snapshot name from the snapshot ID
	snapshotName, err := z.getSnapshotPath(snapshotID)
	if err != nil {
		return nil, fmt.Errorf("unable to get snapshot path: %w", err)
	}

	// Calculate the relative snapshotPath from the ZFS root to the requested node
	snapshotPath := filepath.Join(rootPath, snapshotName)

	// Open the snapshot root
	root, err := os.OpenRoot(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open snapshot root: %w", err)
	}

	return root, nil
}
