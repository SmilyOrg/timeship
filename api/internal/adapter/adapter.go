package adapter

import (
	"io"
	"net/url"
	"path"
	"strings"
)

// Path Handling Convention:
//
// All paths in the adapter layer MUST use the following convention:
//   - Incoming paths: MUST include the adapter name prefix (e.g., "local://path/to/file")
//   - Outgoing paths: MUST include the adapter name prefix (e.g., "local://path/to/file")
//   - Root directory: Represented as "adapter://" (e.g., "local://")
//
// This ensures:
//   - Consistent path handling across all adapter implementations
//   - Clear identification of which adapter owns each path
//   - Proper integration with the VueFinder API specification
//
// Paths are represented as url.URL objects. Helper functions are provided below
// to assist with path manipulation.

// FileNode represents a file or directory
// All Path fields MUST include the adapter prefix (e.g., "local://path/to/file")
type FileNode struct {
	Path         url.URL // Full path with adapter prefix, e.g., "local://documents/file.txt"
	Type         string  // "file" or "dir"
	Basename     string  // Base name without path, e.g., "file.txt"
	Extension    string  // File extension without dot, e.g., "txt"
	Size         int64
	LastModified int64
	MimeType     string
}

// Snapshot represents a point-in-time snapshot of a node
type Snapshot struct {
	// ID is the unique identifier for this snapshot in format "type:backend-id"
	// e.g., "zfs:tank@daily-2024-10-28"
	ID string

	// Type is the snapshot backend type (e.g., "zfs", "git", "borg")
	Type string

	// Timestamp is the Unix timestamp when the snapshot was created
	Timestamp int64

	// Name is the human-readable name/label for the snapshot
	Name string

	// Size is the size of the node in this snapshot (file size or directory size)
	// May be -1 if unknown
	Size int64

	// Metadata contains backend-specific metadata
	Metadata SnapshotMetadata
}

// SnapshotMetadata represents backend-specific metadata for a snapshot
type SnapshotMetadata map[string]interface{}

// Adapter is a marker interface for storage adapters
// All methods are optional - adapters implement only the capabilities they support
type Adapter interface {
	// Adapter is a marker interface - no required methods
}

// Path Helper Functions
//
// These functions help adapters convert between VueFinder paths (with adapter prefix)
// and filesystem-specific paths (without prefix).

// StripPrefix removes the adapter prefix from a path and returns "." for empty paths.
// This handles the common case where filesystem operations require "." for the root directory.
// Examples:
//   - StripPrefix("local://documents/file.txt", "local") -> "documents/file.txt"
//   - StripPrefix("local://", "local") -> "."
//   - StripPrefix("", "local") -> "."
//   - StripPrefix("documents/file.txt", "local") -> "documents/file.txt" (no prefix to strip)
func StripPrefix(vfPath url.URL, adapterName string) string {
	fsPath := vfPath.Path
	if fsPath == "" || fsPath == "/" {
		return "."
	}
	return strings.TrimPrefix(fsPath, "/")
}

// AddPrefix adds the adapter prefix to a path and returns a url.URL.
// Examples:
//   - AddPrefix("documents/file.txt", "local") -> local://documents/file.txt
//   - AddPrefix("", "local") -> local://
func AddPrefix(fsPath, adapterName string) url.URL {
	u := url.URL{
		Scheme: adapterName,
		Host:   "",
		Path:   fsPath,
		// Path:   "/" + strings.TrimPrefix(fsPath, "/"),
	}
	// if fsPath == "" || fsPath == "." {
	// 	u.Path = "/"
	// }
	return u
}

// JoinPath joins path components and returns a url.URL with the adapter prefix.
// Examples:
//   - JoinPath("local://documents", "file.txt", "local") -> local://documents/file.txt
//   - JoinPath("local://", "file.txt", "local") -> local://file.txt
func JoinPath(basePath url.URL, component, adapterName string) url.URL {
	pathPart := path.Join(basePath.Path, component)
	return url.URL{
		Scheme: adapterName,
		Host:   "",
		Path:   pathPart,
	}
}

// Optional capability interfaces that adapters can implement

// Lister lists directory contents (for /index endpoint)
// The path parameter MUST include the adapter prefix (e.g., "local://documents")
// All returned FileNode.Path values MUST include the adapter prefix
type Lister interface {
	ListContents(path url.URL) ([]FileNode, error)
}

// SnapshotLister lists snapshots for a specific path (for /snapshots endpoint)
type SnapshotLister interface {
	ListSnapshots(path url.URL) ([]Snapshot, error)
}

// SubfolderLister lists subdirectories (for /subfolders endpoint)
// The path parameter MUST include the adapter prefix (e.g., "local://documents")
// All returned FileNode.Path values MUST include the adapter prefix
type SubfolderLister interface {
	ListSubfolders(path url.URL) ([]FileNode, error)
}

// Searcher searches for files (for /search endpoint)
// The path parameter MUST include the adapter prefix (e.g., "local://documents")
// All returned FileNode.Path values MUST include the adapter prefix
type Searcher interface {
	Search(path url.URL, filter string) ([]FileNode, error)
}

// Reader reads file content (for /preview and /download endpoints)
type Reader interface {
	ReadStream(path url.URL) (io.ReadCloser, error)
	FileSize(path url.URL) (int64, error)
	MimeType(path url.URL) (string, error)
}

// Writer writes file content (for /upload and /save endpoints)
type Writer interface {
	WriteStream(path url.URL, r io.Reader) error
}

// Creator creates files and directories (for /newfile and /newfolder endpoints)
type Creator interface {
	CreateFile(path url.URL) error
	CreateDirectory(path url.URL) error
}

// Deleter deletes files and directories (for /delete endpoint)
type Deleter interface {
	Delete(path url.URL) error
	DeleteDirectory(path url.URL) error
}

// Mover moves/renames files and directories (for /move and /rename endpoints)
type Mover interface {
	Move(from, to url.URL) error
}

// Archiver creates and extracts archives (for /archive and /unarchive endpoints)
type Archiver interface {
	Archive(items []url.URL, archivePath url.URL) error
	Unarchive(archivePath, targetPath url.URL) error
}

// Existence checks if files/directories exist
type Existence interface {
	FileExists(path url.URL) (bool, error)
	DirectoryExists(path url.URL) (bool, error)
}
