package local

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

const adapterName = "local"

// Adapter implements adapter interfaces for local filesystem
type Adapter struct {
	root *os.Root
	zfs  *ZFS
}

// New creates a new local filesystem adapter
func New(rootPath string) (*Adapter, error) {
	// Open the root directory with os.OpenRoot for traversal-resistant operations
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		root: root,
		zfs:  NewZFS(rootPath),
	}, nil
}

// Close closes the root directory handle
func (a *Adapter) Close() error {
	return a.root.Close()
}

// open opens a file or directory, handling both normal paths and snapshots
// For snapshots: opens from the snapshot directory
// For normal paths: opens from the adapter's root
// The caller is responsible for closing the returned file
func (a *Adapter) open(vfPath url.URL) (*os.File, error) {
	snapshotID := vfPath.Query().Get("snapshot")
	if snapshotID != "" {
		root, relPath, err := a.zfs.OpenSnapshotRoot(vfPath, snapshotID)
		if err != nil {
			return nil, err
		}
		defer root.Close()
		return root.Open(relPath)
	}

	// Normal path
	filePath := adapter.StripPrefix(vfPath, adapterName)
	return a.root.Open(filePath)
}

// stat gets file info, handling both normal paths and snapshots
func (a *Adapter) stat(vfPath url.URL) (os.FileInfo, error) {
	snapshotID := vfPath.Query().Get("snapshot")
	if snapshotID != "" {
		root, relPath, err := a.zfs.OpenSnapshotRoot(vfPath, snapshotID)
		if err != nil {
			return nil, err
		}
		defer root.Close()
		return root.Stat(relPath)
	}

	// Normal path
	filePath := adapter.StripPrefix(vfPath, adapterName)
	return a.root.Stat(filePath)
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
		fullPath := adapter.JoinPath(vfPath, info.Name(), adapterName)

		node := adapter.FileNode{
			Path:         fullPath,
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

// FileExists implements adapter.Existence
func (a *Adapter) FileExists(vfPath url.URL) (bool, error) {
	filePath := adapter.StripPrefix(vfPath, adapterName)

	info, err := a.root.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

// DirectoryExists implements adapter.Existence
func (a *Adapter) DirectoryExists(vfPath url.URL) (bool, error) {
	dirPath := adapter.StripPrefix(vfPath, adapterName)

	info, err := a.root.Stat(dirPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

// GetSnapshots implements adapter.SnapshotProvider
func (a *Adapter) ListSnapshots(path url.URL) ([]adapter.Snapshot, error) {
	return a.zfs.Snapshots(path)
}
