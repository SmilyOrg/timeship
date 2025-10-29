package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// mockAdapter implements adapter.Lister for testing
type mockAdapter struct {
	nodes []adapter.FileNode
	err   error
}

func (m *mockAdapter) ListContents(path string) ([]adapter.FileNode, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.nodes, nil
}

func TestGetIndex(t *testing.T) {
	t.Run("successful listing", func(t *testing.T) {
		// Create mock adapter with test data
		mockNodes := []adapter.FileNode{
			{
				Path:     "local://subdir",
				Type:     "dir",
				Basename: "subdir",
			},
			{
				Path:         "local://file.txt",
				Type:         "file",
				Basename:     "file.txt",
				Extension:    "txt",
				Size:         1024,
				LastModified: 1234567890,
				MimeType:     "text/plain",
			},
		}

		mock := &mockAdapter{nodes: mockNodes}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/?q=index", nil)
		w := httptest.NewRecorder()

		// Call handler with new unified Get method
		server.Get(w, req, GetParams{Q: GetParamsQIndex})

		// Check response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response DirectoryListingResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Verify response
		if response.Adapter != "local" {
			t.Errorf("expected adapter 'local', got '%s'", response.Adapter)
		}

		if response.Dirname != "local://" {
			t.Errorf("expected dirname 'local://', got '%s'", response.Dirname)
		}

		if len(response.Files) != 2 {
			t.Errorf("expected 2 files, got %d", len(response.Files))
		}

		if len(response.Storages) != 1 {
			t.Errorf("expected 1 storage, got %d", len(response.Storages))
		}

		// Verify files are sorted (directories first)
		if response.Files[0].Type != "dir" {
			t.Error("expected first item to be directory")
		}

		if response.Files[1].Type != "file" {
			t.Error("expected second item to be file")
		}

		// Verify file has all fields
		file := response.Files[1]
		if file.Extension == nil || *file.Extension != "txt" {
			t.Error("expected file to have extension 'txt'")
		}
		if file.FileSize == nil || *file.FileSize != 1024 {
			t.Error("expected file size 1024")
		}
		if file.LastModified == nil || *file.LastModified != 1234567890 {
			t.Error("expected file last_modified 1234567890")
		}
		if file.MimeType == nil || *file.MimeType != "text/plain" {
			t.Error("expected file mime_type 'text/plain'")
		}
	})

	t.Run("with custom adapter parameter", func(t *testing.T) {
		mock := &mockAdapter{nodes: []adapter.FileNode{}}
		adapters := map[string]adapter.Adapter{
			"custom": mock,
		}

		server, err := NewServer(adapters, "custom")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index&adapter=custom", nil)
		w := httptest.NewRecorder()

		adapterParam := Adapter("custom")
		server.Get(w, req, GetParams{Q: GetParamsQIndex, Adapter: &adapterParam})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var response DirectoryListingResponse
		json.NewDecoder(resp.Body).Decode(&response)

		if response.Adapter != "custom" {
			t.Errorf("expected adapter 'custom', got '%s'", response.Adapter)
		}
	})

	t.Run("with custom path parameter", func(t *testing.T) {
		mock := &mockAdapter{nodes: []adapter.FileNode{}}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index&path=local://subdir", nil)
		w := httptest.NewRecorder()

		pathParam := "local://subdir"
		server.Get(w, req, GetParams{Q: GetParamsQIndex, Path: &pathParam})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var response DirectoryListingResponse
		json.NewDecoder(resp.Body).Decode(&response)

		if response.Dirname != "local://subdir" {
			t.Errorf("expected dirname 'local://subdir', got '%s'", response.Dirname)
		}
	})

	t.Run("invalid adapter", func(t *testing.T) {
		// Create server with adapters but request invalid adapter
		mock := &mockAdapter{nodes: []adapter.FileNode{}}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index&adapter=invalid", nil)
		w := httptest.NewRecorder()

		invalidAdapter := Adapter("invalid")
		server.Get(w, req, GetParams{Q: GetParamsQIndex, Adapter: &invalidAdapter})

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}

		var errorResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errorResp)

		if errorResp.Status != false {
			t.Error("expected error status to be false")
		}

		if errorResp.Message == "" {
			t.Error("expected error message")
		}
	})

	t.Run("no adapters configured", func(t *testing.T) {
		server, err := NewServer(map[string]adapter.Adapter{}, "")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex})

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}

		var errorResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errorResp)

		if errorResp.Status != false {
			t.Error("expected error status to be false")
		}

		if errorResp.Message == "" {
			t.Error("expected error message")
		}
	})

	t.Run("adapter does not support listing", func(t *testing.T) {
		// Create an adapter that doesn't implement Lister
		type nonListerAdapter struct{}

		adapters := map[string]adapter.Adapter{
			"local": &nonListerAdapter{},
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex})

		resp := w.Result()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Errorf("expected status 501, got %d", resp.StatusCode)
		}
	})

	t.Run("adapter returns error", func(t *testing.T) {
		mock := &mockAdapter{
			err: http.ErrNotSupported,
		}

		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex})

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}

		var errorResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errorResp)

		if errorResp.Status != false {
			t.Error("expected error status to be false")
		}
	})

	t.Run("sorting directories before files", func(t *testing.T) {
		// Create mixed list - files and directories interleaved
		mockNodes := []adapter.FileNode{
			{Path: "local://file1.txt", Type: "file", Basename: "file1.txt"},
			{Path: "local://dir2", Type: "dir", Basename: "dir2"},
			{Path: "local://file2.txt", Type: "file", Basename: "file2.txt"},
			{Path: "local://dir1", Type: "dir", Basename: "dir1"},
			{Path: "local://file3.txt", Type: "file", Basename: "file3.txt"},
		}

		mock := &mockAdapter{nodes: mockNodes}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex})

		var response DirectoryListingResponse
		json.NewDecoder(w.Result().Body).Decode(&response)

		// Check all directories come first
		dirCount := 0
		for i, file := range response.Files {
			if file.Type == "dir" {
				dirCount++
				if i >= 2 {
					t.Error("directories should be first in the list")
				}
			}
		}

		if dirCount != 2 {
			t.Errorf("expected 2 directories, got %d", dirCount)
		}

		// Check directories are sorted alphabetically
		if response.Files[0].Basename != "dir1" {
			t.Errorf("expected first dir to be 'dir1', got '%s'", response.Files[0].Basename)
		}
		if response.Files[1].Basename != "dir2" {
			t.Errorf("expected second dir to be 'dir2', got '%s'", response.Files[1].Basename)
		}

		// Check files are sorted alphabetically after directories
		if response.Files[2].Basename != "file1.txt" {
			t.Errorf("expected first file to be 'file1.txt', got '%s'", response.Files[2].Basename)
		}
	})

	t.Run("verify dirname format for empty path", func(t *testing.T) {
		mockNodes := []adapter.FileNode{
			{Path: "local://file.txt", Type: "file", Basename: "file.txt"},
		}

		mock := &mockAdapter{nodes: mockNodes}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Test with empty path (should default to root)
		req := httptest.NewRequest(http.MethodGet, "/api/?q=index", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex})

		var response DirectoryListingResponse
		json.NewDecoder(w.Result().Body).Decode(&response)

		if response.Dirname != "local://" {
			t.Errorf("expected dirname 'local://' for empty path, got '%s'", response.Dirname)
		}
	})

	t.Run("verify dirname format for path without prefix", func(t *testing.T) {
		mockNodes := []adapter.FileNode{
			{Path: "local://public/media/file.txt", Type: "file", Basename: "file.txt"},
		}

		mock := &mockAdapter{nodes: mockNodes}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Test with path without prefix
		pathParam := "public/media"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=index&path=public/media", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex, Path: &pathParam})

		var response DirectoryListingResponse
		json.NewDecoder(w.Result().Body).Decode(&response)

		if response.Dirname != "local://public/media" {
			t.Errorf("expected dirname 'local://public/media', got '%s'", response.Dirname)
		}
	})

	t.Run("verify dirname preserves existing prefix", func(t *testing.T) {
		mockNodes := []adapter.FileNode{
			{Path: "local://subdir/file.txt", Type: "file", Basename: "file.txt"},
		}

		mock := &mockAdapter{nodes: mockNodes}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		// Test with path that already has prefix
		pathParam := "local://subdir"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=index&path=local://subdir", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex, Path: &pathParam})

		var response DirectoryListingResponse
		json.NewDecoder(w.Result().Body).Decode(&response)

		if response.Dirname != "local://subdir" {
			t.Errorf("expected dirname 'local://subdir', got '%s'", response.Dirname)
		}
	})

	t.Run("verify all file paths have adapter prefix", func(t *testing.T) {
		mockNodes := []adapter.FileNode{
			{Path: "local://dir1", Type: "dir", Basename: "dir1"},
			{Path: "local://file1.txt", Type: "file", Basename: "file1.txt"},
			{Path: "local://public", Type: "dir", Basename: "public"},
		}

		mock := &mockAdapter{nodes: mockNodes}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=index", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQIndex})

		var response DirectoryListingResponse
		json.NewDecoder(w.Result().Body).Decode(&response)

		// Verify all files have the local:// prefix
		for _, file := range response.Files {
			path := string(file.Path)
			if !strings.HasPrefix(path, "local://") {
				t.Errorf("file path '%s' should have 'local://' prefix", path)
			}
		}
	})
}

func TestSendError(t *testing.T) {
	server, err := NewServer(map[string]adapter.Adapter{}, "")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	w := httptest.NewRecorder()

	server.sendError(w, "test error message", http.StatusTeapot)

	resp := w.Result()
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("expected status 418, got %d", resp.StatusCode)
	}

	var errorResp ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errorResp)

	if errorResp.Message != "test error message" {
		t.Errorf("expected message 'test error message', got '%s'", errorResp.Message)
	}

	if errorResp.Status != false {
		t.Error("expected status to be false")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}
