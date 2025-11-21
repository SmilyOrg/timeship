package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/charlievieth/fastwalk"
	"github.com/smilyorg/timeship/api/internal/adapter"
)

func (s *Server) GetStoragesStorageNodes(w http.ResponseWriter, r *http.Request, storage Storage, params GetStoragesStorageNodesParams) {
	// Delegate to the path-based handler with empty path
	pathParams := GetStoragesStorageNodesPathParams{
		Type:     params.Type,
		Filter:   params.Filter,
		Search:   params.Search,
		Children: params.Children,
		Download: params.Download,
		Sort:     (*GetStoragesStorageNodesPathParamsSort)(params.Sort),
		Order:    (*GetStoragesStorageNodesPathParamsOrder)(params.Order),
		Fields:   params.Fields,
		Snapshot: params.Snapshot,
	}
	s.GetStoragesStorageNodesPath(w, r, storage, "", pathParams)
}

// GetStoragesStorageNodesPath handles getting node information or content
// This combines both directory listing and file retrieval functionality
func (s *Server) GetStoragesStorageNodesPath(w http.ResponseWriter, r *http.Request, storage Storage, path NodePath, params GetStoragesStorageNodesPathParams) {
	// Get the storage adapter
	storageAdapter, err := s.getStorage(string(storage))
	if err != nil {
		s.sendError(w, "Storage Not Found", http.StatusNotFound, err.Error(), r.URL.Path)
		return
	}

	// Create url.URL with adapter prefix
	// vfPath := adapter.AddPrefix(nodePath, string(storage))
	vfPath := url.URL{
		Scheme: string(storage),
		Path:   path,
	}

	// Add snapshot query parameter if provided
	if params.Snapshot != nil && *params.Snapshot != "" {
		q := vfPath.Query()
		q.Set("snapshot", *params.Snapshot)
		vfPath.RawQuery = q.Encode()
	}

	// Determine if this is a directory listing or file retrieval based on Accept header
	acceptHeader := r.Header.Get("Accept")

	// If client accepts octet-stream, they want file content
	wantsFileContent := strings.Contains(acceptHeader, "application/octet-stream")

	// Check if the adapter supports listing (for directories) or reading (for files)
	lister, canList := storageAdapter.(adapter.Lister)
	reader, canRead := storageAdapter.(adapter.Reader)

	// First, try to determine if this is a file or directory
	// We'll attempt to list - if it fails, we'll try to read as a file
	if canList && !wantsFileContent {
		// Try to list as a directory
		nodes, err := lister.ListContents(vfPath)
		if err == nil {
			// It's a directory - return listing
			s.serveDirectoryListing(w, r, storage, path, nodes, params, storageAdapter)
			return
		}
	}

	// If listing failed or client wants file content, try to read as a file
	if canRead {
		s.serveFileContent(w, r, storage, path, vfPath, reader, params)
		return
	}

	// Neither listing nor reading worked
	s.sendError(w, "Not Found", http.StatusNotFound, "Node not found or storage does not support required operations", r.URL.Path)
}

// serveDirectoryListing returns directory listing as JSON
func (s *Server) serveDirectoryListing(w http.ResponseWriter, r *http.Request, storage Storage, path string, nodes []adapter.FileNode, params GetStoragesStorageNodesPathParams, storageAdapter adapter.Adapter) {
	// Sort nodes: directories first, then by name
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type == "dir"
		}
		return nodes[i].Basename < nodes[j].Basename
	})

	// Apply type filter if specified
	if params.Type != nil {
		filtered := []adapter.FileNode{}
		for _, node := range nodes {
			if string(*params.Type) == node.Type {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	// Apply filename filter if specified (glob pattern)
	if params.Filter != nil && *params.Filter != "" {
		// TODO: Implement glob pattern matching
		// For now, we'll do simple substring matching
		pattern := *params.Filter
		filtered := []adapter.FileNode{}
		for _, node := range nodes {
			if strings.Contains(node.Basename, strings.Trim(pattern, "*")) {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	// Apply search if specified
	if params.Search != nil && *params.Search != "" {
		// TODO: Implement recursive search
		// For now, we'll do simple name matching on current level
		query := strings.ToLower(*params.Search)
		filtered := []adapter.FileNode{}
		for _, node := range nodes {
			if strings.Contains(strings.ToLower(node.Basename), query) {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	// Convert adapter.FileNode to api.Node
	files := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		apiNode := Node{
			Path:         node.Path.String(),
			Type:         NodeType(node.Type),
			Basename:     node.Basename,
			Extension:    node.Extension,
			FileSize:     node.Size,
			LastModified: node.LastModified,
		}

		// Add optional fields
		if node.MimeType != "" {
			apiNode.MimeType = &node.MimeType
		}

		files = append(files, apiNode)
	}

	// Build list of available storages
	storages := make([]string, 0, len(s.storages))
	for name := range s.storages {
		storages = append(storages, name)
	}
	sort.Strings(storages)

	// Build dirname with storage prefix
	dirname := string(storage) + "://"
	if path != "" {
		dirname += path
	}

	// Create response - Files contains the direct children, not wrapped in a directory node
	response := NodeList{
		Files:    files,
		Dirname:  dirname,
		ReadOnly: false, // TODO: Determine read-only status from adapter capabilities
		Storages: storages,
	}

	// Handle optional fields
	if params.Fields != nil && *params.Fields != "" {
		fields := *params.Fields
		// Parse fields parameter - looking for (total_size)
		if strings.Contains(fields, "(total_size)") {
			// Compute total size if requested
			totalSize, err := s.computeTotalSize(storageAdapter, storage, path)
			if err != nil {
				log.Printf("Failed to compute total_size for %s://%s: %v", storage, path, err)
			} else {
				response.TotalSize = &totalSize
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// serveFileContent streams file content
func (s *Server) serveFileContent(w http.ResponseWriter, r *http.Request, storage Storage, path string, vfPath url.URL, reader adapter.Reader, params GetStoragesStorageNodesPathParams) {
	// Get MIME type
	mimeType, err := reader.MimeType(vfPath)
	if err != nil {
		s.sendError(w, "Not Found", http.StatusNotFound, "Failed to get file MIME type: "+err.Error(), r.URL.Path)
		return
	}

	// Get file size
	fileSize, err := reader.FileSize(vfPath)
	if err != nil {
		s.sendError(w, "Not Found", http.StatusNotFound, "Failed to get file size: "+err.Error(), r.URL.Path)
		return
	}

	// Open file stream
	stream, err := reader.ReadStream(vfPath)
	if err != nil {
		s.sendError(w, "Not Found", http.StatusNotFound, "Failed to open file: "+err.Error(), r.URL.Path)
		return
	}
	defer stream.Close()

	// Set headers
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))

	// Set Content-Disposition if download is requested
	if params.Download != nil && *params.Download {
		basename := getBasename(path)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", basename))
	}

	w.WriteHeader(http.StatusOK)

	// Stream the file content
	_, err = io.Copy(w, stream)
	if err != nil {
		// At this point we've already written headers, so we can't send an error response
		return
	}
}

// getBasename returns the last component of a path
func getBasename(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// computeTotalSize computes the total size of all files in a directory tree
// using fastwalk for parallel traversal
func (s *Server) computeTotalSize(storageAdapter adapter.Adapter, storage Storage, path string) (int64, error) {
	// We need a concrete type that has a root path
	// For now, we'll check if it's a local adapter
	type localAdapter interface {
		GetRootPath() string
	}

	la, ok := storageAdapter.(localAdapter)
	if !ok {
		return 0, fmt.Errorf("storage adapter does not support total size computation")
	}

	rootPath := la.GetRootPath()
	targetPath := rootPath
	if path != "" {
		targetPath = rootPath + "/" + path
	}

	var totalSize atomic.Int64

	conf := fastwalk.Config{
		Follow: false, // Don't follow symlinks to avoid cycles
	}

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log but don't stop on individual errors
			log.Printf("Error walking %s: %v", path, err)
			return nil
		}

		// Only count regular files
		if d.Type().IsRegular() {
			if info, err := d.Info(); err == nil {
				totalSize.Add(info.Size())
			}
		}

		return nil
	}

	err := fastwalk.Walk(&conf, targetPath, walkFn)
	if err != nil {
		return 0, fmt.Errorf("failed to walk directory: %w", err)
	}

	return totalSize.Load(), nil
}
