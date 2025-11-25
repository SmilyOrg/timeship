// Package local provides adapters for local filesystems including ZFS snapshot support.
//
// # ZFS Snapshot Date/Time Parsing
//
// The ZFS adapter automatically parses timestamps from snapshot names using configurable patterns.
// By default, it supports common formats like:
//   - auto-weekly-2025-11-09_00-00
//   - auto-hourly-2025-11-09_13-30
//   - auto-daily-2025-11-09_00-00
//   - backup-2025-11-09_14-30-45 (with seconds)
//   - snapshot_20251109_143045 (compact format)
//   - daily-2025-11-09 (date only)
//
// # Custom Patterns
//
// You can provide custom date/time patterns when creating a ZFS adapter:
//
//	config := local.ZFSConfig{
//		DateTimePatterns: []local.DateTimePattern{
//			{
//				Regex:  `snap_(\d{8})`,
//				Layout: "20060102",
//			},
//			{
//				Regex:  `backup-(\d{4}\.\d{2}\.\d{2})`,
//				Layout: "2006.01.02",
//			},
//		},
//	}
//	zfs := local.NewZFSWithConfig("/path/to/root", config)
//
// Patterns are tried in order, and the first matching pattern is used.
// If no pattern matches, the snapshot directory's modification time is used as a fallback.
package local

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"timeship/internal/adapter"
)

// ZFSConfig holds configuration for the ZFS snapshot provider
type ZFSConfig struct {
	// DateTimePatterns is a list of patterns to try when parsing dates from snapshot names.
	// Each pattern consists of a regex and a Go time layout.
	// The regex should capture the date/time portion of the snapshot name.
	// If empty, defaults to common patterns.
	DateTimePatterns []DateTimePattern
}

// DateTimePattern defines how to extract and parse dates from snapshot names
type DateTimePattern struct {
	// Regex is the regular expression to match and extract the date/time portion
	// It should have a capturing group for the date/time part
	Regex string

	// Layout is the Go time layout string to parse the extracted date/time
	// See https://golang.org/pkg/time/#Parse for format
	Layout string

	// compiled is the compiled regex (cached)
	compiled *regexp.Regexp
}

// DefaultDateTimePatterns returns the default patterns for parsing snapshot names
func DefaultDateTimePatterns() []DateTimePattern {
	return []DateTimePattern{
		{
			// Matches: 2025-11-09_00-00-00 (with seconds - most specific, check first)
			Regex:  `(\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2})`,
			Layout: "2006-01-02_15-04-05",
		},
		{
			// Matches: 20251109_000000 (compact with seconds)
			Regex:  `(\d{8}_\d{6})`,
			Layout: "20060102_150405",
		},
		{
			// Matches: auto-weekly-2025-11-09_00-00, auto-hourly-2025-11-09_00-00, etc.
			Regex:  `(\d{4}-\d{2}-\d{2}_\d{2}-\d{2})`,
			Layout: "2006-01-02_15-04",
		},
		{
			// Matches: 2025-11-09 (date only)
			Regex:  `(\d{4}-\d{2}-\d{2})`,
			Layout: "2006-01-02",
		},
	}
}

// ZFS implements the SnapshotProvider interface for ZFS filesystems
type ZFS struct {
	rootDir          string
	dateTimePatterns []DateTimePattern
}

// NewZFS creates a new ZFS snapshot provider with default configuration
func NewZFS(rootDir string) *ZFS {
	return NewZFSWithConfig(rootDir, ZFSConfig{
		DateTimePatterns: DefaultDateTimePatterns(),
	})
}

// NewZFSWithConfig creates a new ZFS snapshot provider with custom configuration
func NewZFSWithConfig(rootDir string, config ZFSConfig) *ZFS {
	patterns := config.DateTimePatterns
	if len(patterns) == 0 {
		patterns = DefaultDateTimePatterns()
	}

	// Compile all regex patterns
	for i := range patterns {
		if patterns[i].Regex != "" {
			patterns[i].compiled = regexp.MustCompile(patterns[i].Regex)
		}
	}

	return &ZFS{
		rootDir:          rootDir,
		dateTimePatterns: patterns,
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

// parseTimestampFromName attempts to parse a timestamp from a snapshot name
// using the configured date/time patterns. Returns the Unix timestamp and true if successful,
// or 0 and false if no pattern matched.
func (z *ZFS) parseTimestampFromName(name string) (int64, bool) {
	for _, pattern := range z.dateTimePatterns {
		if pattern.compiled == nil {
			continue
		}

		matches := pattern.compiled.FindStringSubmatch(name)
		if len(matches) < 2 {
			continue
		}

		// Extract the date/time string from the first capturing group
		dateTimeStr := matches[1]

		// Try to parse it with the specified layout
		t, err := time.Parse(pattern.Layout, dateTimeStr)
		if err == nil {
			return t.Unix(), true
		}
	}

	return 0, false
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

		// Try to parse timestamp from the snapshot name first
		timestamp, parsed := z.parseTimestampFromName(entry.Name())

		// If parsing failed, fall back to the directory's modification time
		if !parsed {
			info, err := entry.Info()
			if err != nil {
				continue // Skip entries we can't stat
			}
			timestamp = info.ModTime().Unix()
		}

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
