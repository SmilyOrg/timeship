package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// getPreview implements the preview operation
func (s *Server) getPreview(w http.ResponseWriter, params GetParams) {
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

	// Check if adapter supports reading
	reader, ok := adapterInstance.(adapter.Reader)
	if !ok {
		s.sendError(w, "Adapter does not support file reading", http.StatusNotImplemented)
		return
	}

	// Path is required for preview
	if params.Path == nil || *params.Path == "" {
		s.sendError(w, "Path parameter is required for preview operation", http.StatusBadRequest)
		return
	}

	path := string(*params.Path)

	// Get MIME type
	mimeType, err := reader.MimeType(path)
	if err != nil {
		s.sendError(w, "Failed to get file MIME type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get file size
	fileSize, err := reader.FileSize(path)
	if err != nil {
		s.sendError(w, "Failed to get file size: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Open file stream
	stream, err := reader.ReadStream(path)
	if err != nil {
		s.sendError(w, "Failed to open file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// Set headers
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	w.WriteHeader(http.StatusOK)

	// Stream the file content
	_, err = io.Copy(w, stream)
	if err != nil {
		// At this point we've already written headers, so we can't send an error response
		// Just log and return
		return
	}
}
