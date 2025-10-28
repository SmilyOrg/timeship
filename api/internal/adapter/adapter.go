package adapter

import (
	"io"
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
// Helper functions are provided below to assist with path manipulation.

// FileNode represents a file or directory
// All Path fields MUST include the adapter prefix (e.g., "local://path/to/file")
type FileNode struct {
	Path         string // Full path with adapter prefix, e.g., "local://documents/file.txt"
	Type         string // "file" or "dir"
	Basename     string // Base name without path, e.g., "file.txt"
	Extension    string // File extension without dot, e.g., "txt"
	Size         int64
	LastModified int64
	MimeType     string
}

// Adapter is a marker interface for storage adapters
// All methods are optional - adapters implement only the capabilities they support
//
// PATH CONVENTION: All path parameters and return values MUST include the adapter
// name prefix (e.g., "local://", "s3://", "ftp://"). Use the helper functions
// StripPrefix() and AddPrefix() to convert between prefixed and unprefixed paths.
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
func StripPrefix(vfPath, adapterName string) string {
	prefix := adapterName + "://"
	fsPath := strings.TrimPrefix(vfPath, prefix)
	if fsPath == "" {
		return "."
	}
	return fsPath
}

// AddPrefix adds the adapter prefix to a path if not already present.
// Examples:
//   - AddPrefix("documents/file.txt", "local") -> "local://documents/file.txt"
//   - AddPrefix("local://documents/file.txt", "local") -> "local://documents/file.txt" (already has prefix)
//   - AddPrefix("", "local") -> "local://"
func AddPrefix(fsPath, adapterName string) string {
	prefix := adapterName + "://"
	if strings.HasPrefix(fsPath, prefix) {
		return fsPath
	}
	return prefix + fsPath
}

// JoinPath joins path components and ensures the result has the adapter prefix.
// It properly handles the "://" separator to avoid path.Join removing slashes.
// Examples:
//   - JoinPath("local://documents", "file.txt", "local") -> "local://documents/file.txt"
//   - JoinPath("local://", "file.txt", "local") -> "local://file.txt"
//   - JoinPath("documents", "file.txt", "local") -> "local://documents/file.txt"
func JoinPath(basePath, component, adapterName string) string {
	prefix := adapterName + "://"

	// Ensure basePath has the prefix
	if !strings.HasPrefix(basePath, prefix) {
		basePath = prefix + basePath
	}

	// Handle root case
	if basePath == prefix {
		return prefix + component
	}

	// Strip prefix, join, then add it back to avoid path.Join mangling the ://
	baseWithoutPrefix := strings.TrimPrefix(basePath, prefix)
	joined := path.Join(baseWithoutPrefix, component)
	return prefix + joined
}

// Optional capability interfaces that adapters can implement
//
// PATH CONVENTION: All path parameters and FileNode.Path values in these interfaces
// MUST include the adapter prefix (e.g., "local://path/to/file").

// Lister lists directory contents (for /index endpoint)
// The path parameter MUST include the adapter prefix (e.g., "local://documents")
// All returned FileNode.Path values MUST include the adapter prefix
type Lister interface {
	ListContents(path string) ([]FileNode, error)
}

// SubfolderLister lists subdirectories (for /subfolders endpoint)
// The path parameter MUST include the adapter prefix (e.g., "local://documents")
// All returned FileNode.Path values MUST include the adapter prefix
type SubfolderLister interface {
	ListSubfolders(path string) ([]FileNode, error)
}

// Searcher searches for files (for /search endpoint)
// The path parameter MUST include the adapter prefix (e.g., "local://documents")
// All returned FileNode.Path values MUST include the adapter prefix
type Searcher interface {
	Search(path, filter string) ([]FileNode, error)
}

// Reader reads file content (for /preview and /download endpoints)
// All path parameters MUST include the adapter prefix (e.g., "local://documents/file.txt")
type Reader interface {
	ReadStream(path string) (io.ReadCloser, error)
	FileSize(path string) (int64, error)
	MimeType(path string) (string, error)
}

// Writer writes file content (for /upload and /save endpoints)
// All path parameters MUST include the adapter prefix (e.g., "local://documents/file.txt")
type Writer interface {
	WriteStream(path string, r io.Reader) error
}

// Creator creates files and directories (for /newfile and /newfolder endpoints)
// All path parameters MUST include the adapter prefix (e.g., "local://documents/newfile.txt")
type Creator interface {
	CreateFile(path string) error
	CreateDirectory(path string) error
}

// Deleter deletes files and directories (for /delete endpoint)
// All path parameters MUST include the adapter prefix (e.g., "local://documents/file.txt")
type Deleter interface {
	Delete(path string) error
	DeleteDirectory(path string) error
}

// Mover moves/renames files and directories (for /move and /rename endpoints)
// All path parameters MUST include the adapter prefix (e.g., "local://old/path", "local://new/path")
type Mover interface {
	Move(from, to string) error
}

// Archiver creates and extracts archives (for /archive and /unarchive endpoints)
// All path parameters MUST include the adapter prefix (e.g., "local://documents/archive.zip")
type Archiver interface {
	Archive(items []string, archivePath string) error
	Unarchive(archivePath, targetPath string) error
}

// Existence checks if files/directories exist
// All path parameters MUST include the adapter prefix (e.g., "local://documents/file.txt")
type Existence interface {
	FileExists(path string) (bool, error)
	DirectoryExists(path string) (bool, error)
}
