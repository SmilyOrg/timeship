package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// GetStoragesStorageSnapshots handles getting snapshots at storage root
func (s *Server) GetStoragesStorageSnapshots(w http.ResponseWriter, r *http.Request, storage Storage, params GetStoragesStorageSnapshotsParams) {
	// Delegate to the path-based handler with empty path
	pathParams := GetStoragesStorageSnapshotsPathParams{
		Type:   params.Type,
		Limit:  params.Limit,
		Offset: params.Offset,
		Sort:   (*GetStoragesStorageSnapshotsPathParamsSort)(params.Sort),
		Order:  (*GetStoragesStorageSnapshotsPathParamsOrder)(params.Order),
	}
	s.GetStoragesStorageSnapshotsPath(w, r, storage, "", pathParams)
}

// GetStoragesStorageSnapshotsPath handles getting snapshots for a specific node
func (s *Server) GetStoragesStorageSnapshotsPath(w http.ResponseWriter, r *http.Request, storage Storage, path string, params GetStoragesStorageSnapshotsPathParams) {
	// Get the storage adapter
	storageAdapter, err := s.getStorage(string(storage))
	if err != nil {
		s.sendError(w, "Storage Not Found", http.StatusNotFound, err.Error(), r.URL.Path)
		return
	}

	// Check if adapter supports snapshots
	snapshotLister, ok := storageAdapter.(adapter.SnapshotLister)
	if !ok {
		s.sendError(w, "Not Supported", http.StatusNotImplemented, "Storage adapter does not support snapshots", r.URL.Path)
		return
	}

	vfPath := url.URL{
		Scheme: string(storage),
		Path:   path,
	}

	// Get snapshots from the adapter
	snapshots, err := snapshotLister.ListSnapshots(vfPath)
	if err != nil {
		s.sendError(w, "Error", http.StatusInternalServerError, fmt.Sprintf("Failed to get snapshots: %v", err), r.URL.Path)
		return
	}

	// Apply pagination (limit and offset)
	limit := 1000
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
		Path:      path,
		Snapshots: apiSnapshots,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
