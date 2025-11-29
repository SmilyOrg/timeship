package local

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"timeship/internal/storage"
)

const storageName = "local"

// Storage implements storage interfaces for local filesystem
type Storage struct {
	root     *os.Root
	rootPath string
	zfs      *ZFS
}

// New creates a new local filesystem storage
func New(rootPath string) (*Storage, error) {
	// Open the root directory with os.OpenRoot for traversal-resistant operations
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	return &Storage{
		root:     root,
		rootPath: rootPath,
		zfs:      NewZFS(rootPath),
	}, nil
}

// Close closes the root directory handle
func (s *Storage) Close() error {
	return s.root.Close()
}

// GetRootPath returns the root path of this storage
func (s *Storage) GetRootPath() string {
	return s.rootPath
}

func (s *Storage) urlToRelPath(vfPath url.URL) (string, error) {
	if vfPath.Scheme != storageName {
		return "", fmt.Errorf("unexpected storage scheme: %s", vfPath.Scheme)
	}
	path := vfPath.Path
	if path == "" {
		path = "."
	}
	// Strip leading slash - paths are always relative to storage root
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "."
	}
	if !filepath.IsLocal(path) {
		return "", fmt.Errorf("non-local paths are not supported: %s", path)
	}
	path = filepath.Clean(path)
	return path, nil
}

// open opens a file or directory, handling both normal paths and snapshots
// For snapshots: opens from the snapshot directory
// For normal paths: opens from the storage's root
// The caller is responsible for closing the returned file
func (s *Storage) open(vfPath url.URL) (*os.File, error) {
	relPath, err := s.urlToRelPath(vfPath)
	if err != nil {
		return nil, fmt.Errorf("unable to convert path: %w", err)
	}
	snapshotID := vfPath.Query().Get("snapshot")
	if snapshotID == "" {
		return s.root.Open(relPath)
	}
	root, snapshotRelPath, err := s.zfs.SnapshotRoot(relPath, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("unable to open: %w", err)
	}
	defer root.Close()
	return root.Open(snapshotRelPath)
}

// stat gets file info, handling both normal paths and snapshots
func (s *Storage) stat(vfPath url.URL) (os.FileInfo, error) {
	relPath, err := s.urlToRelPath(vfPath)
	if err != nil {
		return nil, fmt.Errorf("unable to convert path: %w", err)
	}
	snapshotID := vfPath.Query().Get("snapshot")
	if snapshotID == "" {
		return s.root.Stat(relPath)
	}
	root, snapshotRelPath, err := s.zfs.SnapshotRoot(relPath, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("unable to open: %w", err)
	}
	defer root.Close()
	return root.Stat(snapshotRelPath)
}

// ListContents implements storage.Lister
func (s *Storage) ListContents(vfPath url.URL) ([]storage.FileNode, error) {
	f, err := s.open(vfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// ReadDir to get directory entries
	entries, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	nodes := make([]storage.FileNode, 0, len(entries))
	for _, info := range entries {
		// Build the full path with storage prefix
		// Always remove leading slash to avoid local:///path issues
		filePath := vfPath
		joinedPath := path.Join(vfPath.Path, info.Name())
		filePath.Path = strings.TrimPrefix(joinedPath, "/")
		filePath.RawQuery = ""

		node := storage.FileNode{
			Path:         filePath,
			Basename:     info.Name(),
			LastModified: info.ModTime().Unix(),
		}

		if info.IsDir() {
			node.Type = "dir"
		} else {
			node.Type = "file"
			node.Extension = strings.TrimPrefix(path.Ext(info.Name()), ".")
			node.Size = info.Size()

			// Detect MIME type
			if node.Extension != "" {
				mimeType, _ := s.MimeType(node.Path)
				node.MimeType = mimeType
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// MimeType implements storage.Reader
func (s *Storage) MimeType(vfPath url.URL) (string, error) {
	file, err := s.open(vfPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read first 512 bytes for MIME detection
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	return http.DetectContentType(buffer[:n]), nil
}

// FileSize implements storage.Reader
func (s *Storage) FileSize(vfPath url.URL) (int64, error) {
	info, err := s.stat(vfPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// LastModified implements storage.Stater
func (s *Storage) LastModified(vfPath url.URL) (int64, error) {
	info, err := s.stat(vfPath)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}

// ReadStream implements storage.Reader
func (s *Storage) ReadStream(vfPath url.URL) (io.ReadCloser, error) {
	return s.open(vfPath)
}

// GetSnapshots implements storage.SnapshotProvider
func (s *Storage) ListSnapshots(vfPath url.URL) ([]storage.Snapshot, error) {
	relPath, err := s.urlToRelPath(vfPath)
	if err != nil {
		return nil, fmt.Errorf("unable to convert path: %w", err)
	}
	return s.zfs.Snapshots(relPath)
}
