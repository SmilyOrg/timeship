package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// GetStoragesStorageNodeSnapshots handles getting snapshots at storage root
func (s *Server) GetStoragesStorageNodeSnapshots(w http.ResponseWriter, r *http.Request, storage Storage, params GetStoragesStorageNodeSnapshotsParams) {
	// Delegate to the path-based handler with empty path
	pathParams := GetStoragesStorageNodeSnapshotsPathParams{
		Type:   params.Type,
		Limit:  params.Limit,
		Offset: params.Offset,
		Sort:   (*GetStoragesStorageNodeSnapshotsPathParamsSort)(params.Sort),
		Order:  (*GetStoragesStorageNodeSnapshotsPathParamsOrder)(params.Order),
	}
	s.GetStoragesStorageNodeSnapshotsPath(w, r, storage, "", pathParams)
}

// GetStoragesStorageNodeSnapshotsPath handles getting snapshots for a specific node
func (s *Server) GetStoragesStorageNodeSnapshotsPath(w http.ResponseWriter, r *http.Request, storage Storage, path string, params GetStoragesStorageNodeSnapshotsPathParams) {
	// Get the storage adapter
	storageAdapter, err := s.getStorage(string(storage))
	if err != nil {
		s.sendError(w, "Storage Not Found", http.StatusNotFound, err.Error(), r.URL.Path)
		return
	}

	// Clean the path - empty path means storage root
	nodePath := path
	if nodePath == "/" {
		nodePath = ""
	}

	// Check if adapter supports snapshots
	snapshotLister, ok := storageAdapter.(adapter.SnapshotLister)
	if !ok {
		s.sendError(w, "Not Supported", http.StatusNotImplemented, "Storage adapter does not support snapshots", r.URL.Path)
		return
	}

	// Create url.URL with adapter prefix
	vfPath := adapter.AddPrefix(nodePath, string(storage))

	log.Printf("GetStoragesStorageNodeSnapshotsPath: storage=%s, path=%s, vfPath=%s", storage, nodePath, vfPath.String())

	// Get snapshots from the adapter
	snapshots, err := snapshotLister.ListSnapshots(vfPath)
	if err != nil {
		s.sendError(w, "Error", http.StatusInternalServerError, fmt.Sprintf("Failed to get snapshots: %v", err), r.URL.Path)
		return
	}

	// Apply pagination (limit and offset)
	limit := 50
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	// Apply offset
	if offset >= len(snapshots) {
		snapshots = []adapter.Snapshot{}
	} else {
		snapshots = snapshots[offset:]
	}

	// Apply limit
	if len(snapshots) > limit {
		snapshots = snapshots[:limit]
	}

	// Convert to API response
	apiSnapshots := make([]Snapshot, len(snapshots))
	for i, snap := range snapshots {
		apiSnapshots[i] = Snapshot{
			Id:        snap.ID,
			Type:      SnapshotType(snap.Type),
			Timestamp: snap.Timestamp,
			Name:      &snap.Name,
		}
		if snap.Size >= 0 {
			apiSnapshots[i].Size = &snap.Size
		}
		if snap.Metadata != nil {
			apiSnapshots[i].Metadata = (*map[string]interface{})(&snap.Metadata)
		}
	}

	response := NodeSnapshotsList{
		Storage:   string(storage),
		Path:      nodePath,
		Snapshots: apiSnapshots,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetStoragesStorageSnapshotNodesSnapshot handles getting storage root as it was in a snapshot
func (s *Server) GetStoragesStorageSnapshotNodesSnapshot(w http.ResponseWriter, r *http.Request, storage Storage, snapshot string, params GetStoragesStorageSnapshotNodesSnapshotParams) {
	// Delegate to the path-based handler with empty path
	pathParams := GetStoragesStorageSnapshotNodesSnapshotPathParams{
		Type:     params.Type,
		Filter:   params.Filter,
		Children: params.Children,
		Sort:     (*GetStoragesStorageSnapshotNodesSnapshotPathParamsSort)(params.Sort),
		Order:    (*GetStoragesStorageSnapshotNodesSnapshotPathParamsOrder)(params.Order),
	}
	s.GetStoragesStorageSnapshotNodesSnapshotPath(w, r, storage, snapshot, "", pathParams)
}

// GetStoragesStorageSnapshotNodesSnapshotPath handles getting node content as it was in a snapshot
func (s *Server) GetStoragesStorageSnapshotNodesSnapshotPath(w http.ResponseWriter, r *http.Request, storage Storage, snapshot string, path string, params GetStoragesStorageSnapshotNodesSnapshotPathParams) {
	// Get the storage adapter
	storageAdapter, err := s.getStorage(string(storage))
	if err != nil {
		s.sendError(w, "Storage Not Found", http.StatusNotFound, err.Error(), r.URL.Path)
		return
	}

	// Check if adapter supports snapshots
	_, ok := storageAdapter.(adapter.SnapshotLister)
	if !ok {
		s.sendError(w, "Not Supported", http.StatusNotImplemented, "Storage adapter does not support snapshots", r.URL.Path)
		return
	}

	// Clean the path - empty path means storage root
	nodePath := path
	if nodePath == "/" {
		nodePath = ""
	}

	// Create url.URL with adapter prefix
	_ = adapter.AddPrefix(nodePath, string(storage))

	log.Printf("GetStoragesStorageSnapshotNodesSnapshotPath: storage=%s, snapshot=%s, path=%s", storage, snapshot, nodePath)

	// For now, we'll return "not yet implemented" for browsing snapshots
	// This requires reading directory/file contents from the snapshot
	s.sendNotImplemented(w, r)
}
