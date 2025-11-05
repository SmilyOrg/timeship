package local

import (
	"fmt"
	"net/url"
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
func (z *ZFS) findSnapshotRoot(nodePath string) (string, error) {
	// Convert to absolute path
	currentPath := nodePath
	if !filepath.IsAbs(currentPath) {
		currentPath = filepath.Join(z.rootDir, currentPath)
	}

	// Make sure we're working with clean paths
	currentPath = filepath.Clean(currentPath)

	// Start from the given path and traverse up
	for {
		zfsSnapshotDir := filepath.Join(currentPath, ".zfs", "snapshot")
		stat, err := os.Stat(zfsSnapshotDir)
		if err == nil && stat.IsDir() {
			// Found it!
			return currentPath, nil
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
func (z *ZFS) Snapshots(path url.URL) ([]adapter.Snapshot, error) {

	// Find the ZFS root
	zfsRoot, err := z.findSnapshotRoot(path.Path)
	if err != nil {
		return nil, err
	}

	if zfsRoot == "" {
		// No ZFS filesystem found
		return []adapter.Snapshot{}, nil
	}

	// List snapshots in .zfs/snapshot
	snapshotDir := filepath.Join(zfsRoot, ".zfs", "snapshot")
	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, err
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
				"zfs_root": zfsRoot,
			},
		}

		snapshots = append(snapshots, snapshot)
	}

	// Sort by timestamp in descending order (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp > snapshots[j].Timestamp
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
func (z *ZFS) OpenSnapshotRoot(path url.URL, snapshotID string) (*os.Root, string, error) {
	// Strip the adapter prefix from the path if present
	fsPath := adapter.StripPrefix(path, "local")
	if fsPath == "." {
		fsPath = ""
	}

	// Convert to absolute path relative to rootDir
	var absPath string
	if filepath.IsAbs(fsPath) {
		absPath = fsPath
	} else {
		absPath = filepath.Join(z.rootDir, fsPath)
	}
	absPath = filepath.Clean(absPath)

	// Find the ZFS root
	zfsRoot, err := z.findSnapshotRoot(absPath)
	if err != nil {
		return nil, "", err
	}

	if zfsRoot == "" {
		return nil, "", fmt.Errorf("ZFS root not found for path: %s", path.String())
	}

	// Get the snapshot name from the snapshot ID
	snapshotName, err := z.getSnapshotPath(snapshotID)
	if err != nil {
		return nil, "", err
	}

	// Calculate the relative path from the ZFS root to the requested node
	relPath, err := filepath.Rel(zfsRoot, absPath)
	if err != nil {
		return nil, "", err
	}

	// Construct the path to the snapshot root
	snapshotRootPath := filepath.Join(zfsRoot, ".zfs", "snapshot", snapshotName)

	// Open the snapshot root
	snapshotRoot, err := os.OpenRoot(snapshotRootPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open snapshot root: %w", err)
	}

	// Return the root and the relative path within it
	// Convert "." to empty string for consistency
	if relPath == "." {
		relPath = ""
	}

	return snapshotRoot, relPath, nil
}
