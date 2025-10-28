package api

import (
	"encoding/json"
	"net/http"

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
