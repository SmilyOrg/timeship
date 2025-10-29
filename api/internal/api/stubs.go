package api

import (
	"net/http"
)

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
