package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"timeship/internal/adapter"
)

// Server implements the ServerInterface
type Server struct {
	storages       map[string]adapter.Adapter
	defaultStorage string
}

// NewServer creates a new API server
// defaultStorage specifies which storage to use as default
// Returns an error if the defaultStorage is not found in the storages map
func NewServer(storages map[string]adapter.Adapter, defaultStorage string) (*Server, error) {
	if defaultStorage != "" {
		if _, ok := storages[defaultStorage]; !ok {
			return nil, fmt.Errorf("default storage %q not found in storages map", defaultStorage)
		}
	}

	return &Server{
		storages:       storages,
		defaultStorage: defaultStorage,
	}, nil
}

// getStorage returns the storage adapter for the given name.
// Returns the adapter and an error if the storage is not found.
func (s *Server) getStorage(name string) (adapter.Adapter, error) {
	if name == "" {
		return nil, fmt.Errorf("storage name is required")
	}

	adpt, ok := s.storages[name]
	if !ok {
		return nil, fmt.Errorf("storage not found: %s", name)
	}

	return adpt, nil
}

// sendError sends a RFC 9457 Problem Details error response
func (s *Server) sendError(w http.ResponseWriter, title string, status int, detail string, instance string) {
	response := ErrorResponse{
		Message: fmt.Sprintf("%s: %s", title, detail),
		Status:  false,
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// sendNotImplemented sends a 501 Not Implemented response
func (s *Server) sendNotImplemented(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not Implemented", http.StatusNotImplemented, "This operation is not yet implemented", r.URL.Path)
}
