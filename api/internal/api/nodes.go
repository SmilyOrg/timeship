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

	"timeship/internal/storage"

	"github.com/charlievieth/fastwalk"
)

// extractPath returns just the path component from a url.URL without the scheme and host
func extractPath(u url.URL) string {
	// Return just the path, stripping leading slash if present
	return strings.TrimPrefix(u.Path, "/")
}

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
func (s *Server) GetStoragesStorageNodesPath(w http.ResponseWriter, r *http.Request, storageName Storage, path NodePath, params GetStoragesStorageNodesPathParams) {
	// Get the storage
	store, err := s.getStorage(string(storageName))
	if err != nil {
		s.sendError(w, "Storage Not Found", http.StatusNotFound, err.Error(), r.URL.Path)
		return
	}

	vfPath := url.URL{
		Scheme: string(storageName),
		Path:   path,
	}

	// Add snapshot query parameter if provided
	if params.Snapshot != nil && *params.Snapshot != "" {
		q := vfPath.Query()
		q.Set("snapshot", *params.Snapshot)
		vfPath.RawQuery = q.Encode()
	}

	// Determine if client wants JSON metadata or file content based on Accept header
	acceptHeader := r.Header.Get("Accept")
	wantsJSON := strings.Contains(acceptHeader, "application/json")

	// Check if the storage supports listing (for directories) or reading (for files)
	lister, canList := store.(storage.Lister)
	reader, canRead := store.(storage.Reader)

	// First, try to list as a directory
	if canList {
		nodes, err := lister.ListContents(vfPath)
		if err == nil {
			// It's a directory - return listing as JSON
			s.serveDirectoryListing(w, r, storageName, path, nodes, params, store)
			return
		}
	}

	// Not a directory, try to handle as a file
	if canRead {
		// If client wants JSON, return file metadata
		if wantsJSON {
			s.serveFileMetadata(w, r, storageName, path, vfPath, reader, params)
			return
		}
		// Otherwise, return file content
		s.serveFileContent(w, r, storageName, path, vfPath, reader, params)
		return
	}

	// Neither listing nor reading worked
	s.sendError(w, "Not Found", http.StatusNotFound, "Node not found or storage does not support required operations", r.URL.Path)
}

// serveDirectoryListing returns directory listing as JSON
func (s *Server) serveDirectoryListing(w http.ResponseWriter, r *http.Request, storageName Storage, path string, nodes []storage.FileNode, params GetStoragesStorageNodesPathParams, store storage.Storage) {
	// Sort nodes: directories first, then by name
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type == "dir"
		}
		return nodes[i].Basename < nodes[j].Basename
	})

	// Apply type filter if specified
	if params.Type != nil {
		filtered := []storage.FileNode{}
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
		filtered := []storage.FileNode{}
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
		filtered := []storage.FileNode{}
		for _, node := range nodes {
			if strings.Contains(strings.ToLower(node.Basename), query) {
				filtered = append(filtered, node)
			}
		}
		nodes = filtered
	}

	// Convert storage.FileNode to api.Node
	files := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		apiNode := Node{
			Path:         extractPath(node.Path),
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

	// dirname is just the path without storage prefix
	dirname := path

	// Create response - Files contains the direct children, not wrapped in a directory node
	response := NodeList{
		Files:    files,
		Dirname:  dirname,
		ReadOnly: false, // TODO: Determine read-only status from storage capabilities
		Storages: storages,
	}

	// Handle optional fields
	if params.Fields != nil && *params.Fields != "" {
		fields := *params.Fields
		// Parse fields parameter - looking for (total_size)
		if strings.Contains(fields, "(total_size)") {
			// Compute total size if requested
			totalSize, err := s.computeTotalSize(store, storageName, path)
			if err != nil {
				log.Printf("Failed to compute total_size for %s://%s: %v", storageName, path, err)
			} else {
				response.TotalSize = &totalSize
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// serveFileMetadata returns file metadata as JSON
func (s *Server) serveFileMetadata(w http.ResponseWriter, r *http.Request, storageName Storage, path string, vfPath url.URL, reader storage.Reader, params GetStoragesStorageNodesPathParams) {
	// Get file size
	fileSize, err := reader.FileSize(vfPath)
	if err != nil {
		s.sendError(w, "Not Found", http.StatusNotFound, "Failed to get file size: "+err.Error(), r.URL.Path)
		return
	}

	// Get MIME type
	mimeType, err := reader.MimeType(vfPath)
	if err != nil {
		log.Printf("Failed to get MIME type for %s: %v", vfPath.String(), err)
		mimeType = "application/octet-stream"
	}

	// Get last modified time if storage supports it
	var lastModified int64
	if stater, ok := reader.(storage.Stater); ok {
		lastModified, err = stater.LastModified(vfPath)
		if err != nil {
			log.Printf("Failed to get last modified time for %s: %v", vfPath.String(), err)
			lastModified = 0
		}
	}

	// Get basename and extension
	basename := getBasename(path)
	extension := ""
	if idx := strings.LastIndex(basename, "."); idx > 0 {
		extension = basename[idx:]
	}

	// Create node response with path relative to storage root
	node := Node{
		Path:         path,
		Type:         NodeType("file"),
		Basename:     basename,
		Extension:    extension,
		FileSize:     fileSize,
		LastModified: lastModified,
	}

	if mimeType != "" {
		node.MimeType = &mimeType
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(node)
}

// serveFileContent streams file content
func (s *Server) serveFileContent(w http.ResponseWriter, r *http.Request, storageName Storage, path string, vfPath url.URL, reader storage.Reader, params GetStoragesStorageNodesPathParams) {
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
func (s *Server) computeTotalSize(store storage.Storage, storage Storage, path string) (int64, error) {
	// We need a concrete type that has a root path
	// For now, we'll check if it's a local storage
	type localStorage interface {
		GetRootPath() string
	}

	la, ok := store.(localStorage)
	if !ok {
		return 0, fmt.Errorf("storage does not support total size computation")
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
