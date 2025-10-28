package adapter

import (
	"io"
)

// FileNode represents a file or directory
type FileNode struct {
	Path         string
	Type         string // "file" or "dir"
	Basename     string
	Extension    string
	Size         int64
	LastModified int64
	MimeType     string
}

// Optional capability interfaces that adapters can implement

// Lister lists directory contents (for /index endpoint)
type Lister interface {
	ListContents(path string) ([]FileNode, error)
}

// SubfolderLister lists subdirectories (for /subfolders endpoint)
type SubfolderLister interface {
	ListSubfolders(path string) ([]FileNode, error)
}

// Searcher searches for files (for /search endpoint)
type Searcher interface {
	Search(path, filter string) ([]FileNode, error)
}

// Reader reads file content (for /preview and /download endpoints)
type Reader interface {
	ReadStream(path string) (io.ReadCloser, error)
	FileSize(path string) (int64, error)
	MimeType(path string) (string, error)
}

// Writer writes file content (for /upload and /save endpoints)
type Writer interface {
	WriteStream(path string, r io.Reader) error
}

// Creator creates files and directories (for /newfile and /newfolder endpoints)
type Creator interface {
	CreateFile(path string) error
	CreateDirectory(path string) error
}

// Deleter deletes files and directories (for /delete endpoint)
type Deleter interface {
	Delete(path string) error
	DeleteDirectory(path string) error
}

// Mover moves/renames files and directories (for /move and /rename endpoints)
type Mover interface {
	Move(from, to string) error
}

// Archiver creates and extracts archives (for /archive and /unarchive endpoints)
type Archiver interface {
	Archive(items []string, archivePath string) error
	Unarchive(archivePath, targetPath string) error
}

// Existence checks if files/directories exist
type Existence interface {
	FileExists(path string) (bool, error)
	DirectoryExists(path string) (bool, error)
}
