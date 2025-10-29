package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

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

	// Clean the path - empty path means storage root
	nodePath := string(path)
	if nodePath == "/" {
		nodePath = ""
	}

	// Determine if this is a directory listing or file retrieval based on Accept header
	acceptHeader := r.Header.Get("Accept")

	// If client accepts octet-stream, they want file content
	wantsFileContent := strings.Contains(acceptHeader, "application/octet-stream")

	// Check if the adapter supports listing (for directories) or reading (for files)
	lister, canList := storageAdapter.(adapter.Lister)
	reader, canRead := storageAdapter.(adapter.Reader)

	log.Printf("GetStoragesStorageNodesPath: storage=%s, path=%s, wantsFileContent=%v, canList=%v, canRead=%v", storage, nodePath, wantsFileContent, canList, canRead)

	// First, try to determine if this is a file or directory
	// We'll attempt to list - if it fails, we'll try to read as a file
	if canList && !wantsFileContent {
		// Try to list as a directory
		nodes, err := lister.ListContents(nodePath)
		if err == nil {
			// It's a directory - return listing
			s.serveDirectoryListing(w, r, storage, nodePath, nodes, params)
			return
		}
	}

	// If listing failed or client wants file content, try to read as a file
	if canRead {
		s.serveFileContent(w, r, storage, nodePath, reader, params)
		return
	}

	// Neither listing nor reading worked
	s.sendError(w, "Not Found", http.StatusNotFound, "Node not found or storage does not support required operations", r.URL.Path)
}

// serveDirectoryListing returns directory listing as JSON
func (s *Server) serveDirectoryListing(w http.ResponseWriter, r *http.Request, storage Storage, path string, nodes []adapter.FileNode, params GetStoragesStorageNodesPathParams) {
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
			Path:     node.Path,
			Type:     NodeType(node.Type),
			Basename: node.Basename,
			Storage:  string(storage),
		}

		// Add optional fields
		if node.Extension != "" {
			apiNode.Extension = &node.Extension
		}
		if node.MimeType != "" {
			apiNode.MimeType = &node.MimeType
		}
		if node.Size > 0 {
			apiNode.FileSize = &node.Size
		}
		if node.LastModified > 0 {
			apiNode.LastModified = &node.LastModified
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
		Adapter:  string(storage),
		Dirname:  dirname,
		Storage:  string(storage),
		Storages: storages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// serveFileContent streams file content
func (s *Server) serveFileContent(w http.ResponseWriter, r *http.Request, storage Storage, path string, reader adapter.Reader, params GetStoragesStorageNodesPathParams) {
	// Get MIME type
	mimeType, err := reader.MimeType(path)
	if err != nil {
		s.sendError(w, "Not Found", http.StatusNotFound, "Failed to get file MIME type: "+err.Error(), r.URL.Path)
		return
	}

	// Get file size
	fileSize, err := reader.FileSize(path)
	if err != nil {
		s.sendError(w, "Not Found", http.StatusNotFound, "Failed to get file size: "+err.Error(), r.URL.Path)
		return
	}

	// Open file stream
	stream, err := reader.ReadStream(path)
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
