package api

import (
	"encoding/json"
	"net/http"
	"sort"
)

// GetStorages lists all available storage backends
func (s *Server) GetStorages(w http.ResponseWriter, r *http.Request) {
	// Build list of available storages
	storages := make([]string, 0, len(s.storages))
	for name := range s.storages {
		storages = append(storages, name)
	}

	// Sort alphabetically
	sort.Strings(storages)

	response := struct {
		Storages []string `json:"storages"`
	}{
		Storages: storages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
