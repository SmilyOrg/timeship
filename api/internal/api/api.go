package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// Server implements the ServerInterface
type Server struct {
	adapters map[string]adapter.Adapter
}

// NewServer creates a new API server
func NewServer(adapters map[string]adapter.Adapter) *Server {
	return &Server{
		adapters: adapters,
	}
}

// GetIndex implements GET /api/index
func (s *Server) GetIndex(w http.ResponseWriter, r *http.Request, params GetIndexParams) {
	// Determine which adapter to use
	adapterKey := "local"
	if params.Adapter != nil {
		adapterKey = string(*params.Adapter)
	}

	// Get the adapter
	adapterInstance, ok := s.adapters[adapterKey]
	if !ok {
		s.sendError(w, "Invalid or unconfigured adapter", http.StatusBadRequest)
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
		path = string(*params.Path)
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

	// Build list of available storage adapters
	storages := make(AdapterList, 0, len(s.adapters))
	for key := range s.adapters {
		storages = append(storages, Adapter(key))
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

// sendError sends an error response
func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	response := ErrorResponse{
		Message: message,
		Status:  false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Stub implementations for remaining endpoints

func (s *Server) Options(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) PostArchive(w http.ResponseWriter, r *http.Request, params PostArchiveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostDelete(w http.ResponseWriter, r *http.Request, params PostDeleteParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetDownload(w http.ResponseWriter, r *http.Request, params GetDownloadParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostMove(w http.ResponseWriter, r *http.Request, params PostMoveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostNewfile(w http.ResponseWriter, r *http.Request, params PostNewfileParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostNewfolder(w http.ResponseWriter, r *http.Request, params PostNewfolderParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetPreview(w http.ResponseWriter, r *http.Request, params GetPreviewParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostRename(w http.ResponseWriter, r *http.Request, params PostRenameParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostSave(w http.ResponseWriter, r *http.Request, params PostSaveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetSearch(w http.ResponseWriter, r *http.Request, params GetSearchParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) GetSubfolders(w http.ResponseWriter, r *http.Request, params GetSubfoldersParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostUnarchive(w http.ResponseWriter, r *http.Request, params PostUnarchiveParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) PostUpload(w http.ResponseWriter, r *http.Request, params PostUploadParams) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}
