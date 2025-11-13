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

	"github.com/smilyorg/timeship/api/internal/adapter"
)

const adapterName = "local"

// Adapter implements adapter interfaces for local filesystem
type Adapter struct {
	root     *os.Root
	rootPath string
	zfs      *ZFS
}

// New creates a new local filesystem adapter
func New(rootPath string) (*Adapter, error) {
	// Open the root directory with os.OpenRoot for traversal-resistant operations
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		root:     root,
		rootPath: rootPath,
		zfs:      NewZFS(rootPath),
	}, nil
}

// Close closes the root directory handle
func (a *Adapter) Close() error {
	return a.root.Close()
}

// GetRootPath returns the root path of this adapter
func (a *Adapter) GetRootPath() string {
	return a.rootPath
}

func (a *Adapter) urlToRelPath(vfPath url.URL) (string, error) {
	if vfPath.Scheme != adapterName {
		return "", fmt.Errorf("unexpected adapter scheme: %s", vfPath.Scheme)
	}
	path := vfPath.Path
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
// For normal paths: opens from the adapter's root
// The caller is responsible for closing the returned file
func (a *Adapter) open(vfPath url.URL) (*os.File, error) {
	relPath, err := a.urlToRelPath(vfPath)
	if err != nil {
		return nil, fmt.Errorf("unable to convert path: %w", err)
	}
	snapshotID := vfPath.Query().Get("snapshot")
	if snapshotID == "" {
		return a.root.Open(relPath)
	}
	root, err := a.zfs.SnapshotRoot(relPath, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("unable to open: %w", err)
	}
	defer root.Close()
	return root.Open(relPath)
}

// stat gets file info, handling both normal paths and snapshots
func (a *Adapter) stat(vfPath url.URL) (os.FileInfo, error) {
	relPath, err := a.urlToRelPath(vfPath)
	if err != nil {
		return nil, fmt.Errorf("unable to convert path: %w", err)
	}
	snapshotID := vfPath.Query().Get("snapshot")
	if snapshotID == "" {
		return a.root.Stat(relPath)
	}
	root, err := a.zfs.SnapshotRoot(relPath, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("unable to open: %w", err)
	}
	defer root.Close()
	return root.Stat(relPath)
}

// ListContents implements adapter.Lister
func (a *Adapter) ListContents(vfPath url.URL) ([]adapter.FileNode, error) {
	f, err := a.open(vfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// ReadDir to get directory entries
	entries, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	nodes := make([]adapter.FileNode, 0, len(entries))
	for _, info := range entries {
		// Build the full path with adapter prefix
		// fullPath := adapter.JoinPath(vfPath, info.Name(), adapterName)
		// vfPath.JoinPath(f.Name(), )
		filePath := vfPath
		filePath.Path = path.Join(vfPath.Path, info.Name())
		filePath.RawQuery = ""

		node := adapter.FileNode{
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
				mimeType, _ := a.MimeType(node.Path)
				node.MimeType = mimeType
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// MimeType implements adapter.Reader
func (a *Adapter) MimeType(vfPath url.URL) (string, error) {
	file, err := a.open(vfPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read first 512 bytes for MIME detection
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	return http.DetectContentType(buffer[:n]), nil
}

// FileSize implements adapter.Reader
func (a *Adapter) FileSize(vfPath url.URL) (int64, error) {
	info, err := a.stat(vfPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ReadStream implements adapter.Reader
func (a *Adapter) ReadStream(vfPath url.URL) (io.ReadCloser, error) {
	return a.open(vfPath)
}

// GetSnapshots implements adapter.SnapshotProvider
func (a *Adapter) ListSnapshots(vfPath url.URL) ([]adapter.Snapshot, error) {
	relPath, err := a.urlToRelPath(vfPath)
	if err != nil {
		return nil, fmt.Errorf("unable to convert path: %w", err)
	}
	return a.zfs.Snapshots(relPath)
}
