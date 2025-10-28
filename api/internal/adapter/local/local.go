package local

import (
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// Adapter implements adapter interfaces for local filesystem
type Adapter struct {
	root *os.Root
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
	}, nil
}

// Close closes the root directory handle
func (a *Adapter) Close() error {
	return a.root.Close()
}

// resolvePath strips the "local://" prefix from a VueFinder path
// and converts empty paths to "." for os.Root compatibility
func (a *Adapter) resolvePath(path string) string {
	// Remove "local://" prefix if present
	path = strings.TrimPrefix(path, "local://")

	// os.Root requires "." for the root directory, not ""
	if path == "" {
		return "."
	}

	return path
}

// ListContents implements adapter.Lister
func (a *Adapter) ListContents(vfPath string) ([]adapter.FileNode, error) {
	dirPath := a.resolvePath(vfPath)

	// Open the directory within the root
	f, err := a.root.Open(dirPath)
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
		node := adapter.FileNode{
			Path:     path.Join(vfPath, info.Name()),
			Basename: info.Name(),
		}

		if info.IsDir() {
			node.Type = "dir"
		} else {
			node.Type = "file"
			node.Extension = strings.TrimPrefix(path.Ext(info.Name()), ".")
			node.Size = info.Size()
			node.LastModified = info.ModTime().Unix()

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
func (a *Adapter) MimeType(vfPath string) (string, error) {
	filePath := a.resolvePath(vfPath)

	file, err := a.root.Open(filePath)
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
func (a *Adapter) FileSize(vfPath string) (int64, error) {
	filePath := a.resolvePath(vfPath)

	info, err := a.root.Stat(filePath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// ReadStream implements adapter.Reader
func (a *Adapter) ReadStream(vfPath string) (io.ReadCloser, error) {
	filePath := a.resolvePath(vfPath)
	return a.root.Open(filePath)
}

// FileExists implements adapter.Existence
func (a *Adapter) FileExists(vfPath string) (bool, error) {
	filePath := a.resolvePath(vfPath)

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
func (a *Adapter) DirectoryExists(vfPath string) (bool, error) {
	dirPath := a.resolvePath(vfPath)

	info, err := a.root.Stat(dirPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}
