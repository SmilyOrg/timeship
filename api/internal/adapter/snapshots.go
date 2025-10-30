package adapter

import (
	"time"
)

// Snapshots lists snapshots for a specific path
// The path parameter MUST include the adapter prefix (e.g., "local://documents/file.txt")
// Returns snapshots sorted in descending order by timestamp
type Snapshots interface {
	// GetSnapshots returns all snapshots available for the given path
	// Returns empty slice if no snapshots found (not an error)
	GetSnapshots(path string) ([]Snapshot, error)

	// GetSnapshotsOfType returns snapshots filtered by type
	// Returns empty slice if no snapshots found (not an error)
	GetSnapshotsOfType(path string, snapshotType string) ([]Snapshot, error)

	// ReadSnapshotFile reads a file from a snapshot at the specified path and snapshot ID
	// path: The file path (relative to storage root, may include adapter prefix)
	// snapshotID: The snapshot identifier in format "type:backend-id"
	ReadSnapshotFile(path string, snapshotID string) ([]byte, error)

	// ListSnapshotContents lists the contents of a directory in a snapshot
	// path: The directory path (relative to storage root, may include adapter prefix)
	// snapshotID: The snapshot identifier in format "type:backend-id"
	// Returns FileNode entries for the directory contents at that snapshot point in time
	ListSnapshotContents(path string, snapshotID string) ([]FileNode, error)
}

// SnapshotProvider is an optional capability that adapters can implement
// to support point-in-time snapshot browsing
type SnapshotProvider interface {
	// GetAvailableSnapshotTypes returns the types of snapshots this adapter can provide
	// e.g., ["zfs"] or ["zfs", "git"]
	GetAvailableSnapshotTypes() []string

	// GetSnapshots returns all snapshots available for the given path
	// The path parameter MUST include the adapter prefix (e.g., "local://documents/file.txt")
	// Returns snapshots sorted in descending order by timestamp
	GetSnapshots(path string) ([]Snapshot, error)

	// GetSnapshotsOfType returns snapshots of a specific type for the given path
	// The path parameter MUST include the adapter prefix
	// Returns snapshots sorted in descending order by timestamp
	GetSnapshotsOfType(path string, snapshotType string) ([]Snapshot, error)

	// ReadSnapshotFile reads a file from a snapshot at the specified path and snapshot ID
	// path: The file path (must include adapter prefix)
	// snapshotID: The snapshot identifier in format "type:backend-id"
	ReadSnapshotFile(path string, snapshotID string) ([]byte, error)

	// ListSnapshotContents lists the contents of a directory in a snapshot
	// path: The directory path (must include adapter prefix)
	// snapshotID: The snapshot identifier in format "type:backend-id"
	// Returns FileNode entries for the directory contents at that snapshot point in time
	ListSnapshotContents(path string, snapshotID string) ([]FileNode, error)
}

// Helper function to convert time.Time to Unix timestamp
func TimeToTimestamp(t time.Time) int64 {
	return t.Unix()
}
