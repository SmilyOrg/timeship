package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// getIndex implements the index operation
func (s *Server) getIndex(w http.ResponseWriter, params GetParams) {
	// Get the adapter
	adapterInstance, adapterKey, err := s.getAdapter(params.Adapter)
	if err != nil {
		if adapterKey == "" {
			s.sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if adapter supports listing
	lister, ok := adapterInstance.(adapter.Lister)
	if !ok {
		s.sendError(w, "Adapter does not support directory listing", http.StatusNotImplemented)
		return
	}

	// Determine the path to list
	path := adapterKey + "://"
	if params.Path != nil {
		path = *params.Path
	}

	// List the directory contents
	nodes, err := lister.ListContents(path)
	if err != nil {
		s.sendError(w, "Failed to list directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Sort: directories first, then files
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type == "dir" // directories come first
		}
		return nodes[i].Basename < nodes[j].Basename
	})

	// Convert adapter.FileNode to api.FileNode
	apiFiles := make([]FileNode, len(nodes))
	for i, node := range nodes {
		apiFiles[i] = FileNode{
			Path:     Path(node.Path),
			Basename: Basename(node.Basename),
			Type:     FileType(node.Type),
			Storage:  Adapter(adapterKey),
		}

		// Add optional fields for files
		if node.Type == "file" {
			if node.Extension != "" {
				ext := Extension(node.Extension)
				apiFiles[i].Extension = &ext
			}
			if node.Size > 0 {
				size := FileSize(node.Size)
				apiFiles[i].Size = &size
			}
			if node.LastModified > 0 {
				ts := Timestamp(node.LastModified)
				apiFiles[i].LastModified = &ts
			}
			if node.MimeType != "" {
				mime := MimeType(node.MimeType)
				apiFiles[i].MimeType = &mime
			}
		}
	}

	// Build list of available storage adapters (in order)
	storages := make(AdapterList, 0, len(s.adapters))
	for name := range s.adapters {
		storages = append(storages, Adapter(name))
	}

	// Create response
	response := DirectoryListingResponse{
		Adapter:  Adapter(adapterKey),
		Dirname:  DirectoryPath(path),
		Files:    apiFiles,
		Storages: storages,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
