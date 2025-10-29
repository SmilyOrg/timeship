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

// Archive operations - not implemented yet

func (s *Server) GetStoragesStorageArchives(w http.ResponseWriter, r *http.Request, storage Storage, params GetStoragesStorageArchivesParams) {
	s.sendNotImplemented(w, r)
}

func (s *Server) PostStoragesStorageArchives(w http.ResponseWriter, r *http.Request, storage Storage, params PostStoragesStorageArchivesParams) {
	s.sendNotImplemented(w, r)
}

func (s *Server) PostStoragesStorageArchivesPath(w http.ResponseWriter, r *http.Request, storage Storage, path string) {
	s.sendNotImplemented(w, r)
}

// Copy and Move operations - not implemented yet

func (s *Server) PostStoragesStorageCopies(w http.ResponseWriter, r *http.Request, storage Storage) {
	s.sendNotImplemented(w, r)
}

func (s *Server) PostStoragesStorageMoves(w http.ResponseWriter, r *http.Request, storage Storage) {
	s.sendNotImplemented(w, r)
}

// Node CRUD operations - only GET is implemented

// Pathless node endpoints (for storage root)

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

func (s *Server) PostStoragesStorageNodes(w http.ResponseWriter, r *http.Request, storage Storage) {
	// Delegate to the path-based handler with empty path
	s.PostStoragesStorageNodesPath(w, r, storage, "")
}

// Path-based node endpoints

func (s *Server) DeleteStoragesStorageNodesPath(w http.ResponseWriter, r *http.Request, storage Storage, path NodePath, params DeleteStoragesStorageNodesPathParams) {
	s.sendNotImplemented(w, r)
}

func (s *Server) PatchStoragesStorageNodesPath(w http.ResponseWriter, r *http.Request, storage Storage, path NodePath) {
	s.sendNotImplemented(w, r)
}

func (s *Server) PostStoragesStorageNodesPath(w http.ResponseWriter, r *http.Request, storage Storage, path NodePath) {
	s.sendNotImplemented(w, r)
}
