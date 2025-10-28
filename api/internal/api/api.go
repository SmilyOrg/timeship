package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// Server implements the ServerInterface
type Server struct {
	adapters       map[string]adapter.Adapter
	defaultAdapter string
}

// NewServer creates a new API server
// defaultAdapter specifies which adapter to use when no adapter parameter is provided
// Returns an error if the defaultAdapter is not found in the adapters map
func NewServer(adapters map[string]adapter.Adapter, defaultAdapter string) (*Server, error) {
	if defaultAdapter != "" {
		if _, ok := adapters[defaultAdapter]; !ok {
			return nil, fmt.Errorf("default adapter %q not found in adapters map", defaultAdapter)
		}
	}

	return &Server{
		adapters:       adapters,
		defaultAdapter: defaultAdapter,
	}, nil
}

// getAdapter returns the adapter for the given name, or the default adapter if name is empty.
// Returns the adapter, its name, and an error if the adapter is not found or not configured.
func (s *Server) getAdapter(adapter *Adapter) (adapter.Adapter, string, error) {
	name := ""
	if adapter == nil {
		name = s.defaultAdapter
	} else {
		name = string(*adapter)
	}

	if name == "" {
		return nil, "", fmt.Errorf("no adapters configured")
	}

	adpt, ok := s.adapters[name]
	if !ok {
		return nil, name, fmt.Errorf("invalid or unconfigured adapter: %s", name)
	}

	return adpt, name, nil
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
