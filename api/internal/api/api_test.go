package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// mockAdapterV2 implements adapter.Lister and adapter.Reader for testing v2 API
type mockAdapterV2 struct {
	nodes       []adapter.FileNode
	listErr     error
	content     string
	mimeType    string
	size        int64
	readErr     error
	mimeTypeErr error
	sizeErr     error
}

func (m *mockAdapterV2) ListContents(path string) ([]adapter.FileNode, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.nodes, nil
}

func (m *mockAdapterV2) MimeType(path string) (string, error) {
	if m.mimeTypeErr != nil {
		return "", m.mimeTypeErr
	}
	return m.mimeType, nil
}

func (m *mockAdapterV2) FileSize(path string) (int64, error) {
	if m.sizeErr != nil {
		return 0, m.sizeErr
	}
	return m.size, nil
}

func (m *mockAdapterV2) ReadStream(path string) (io.ReadCloser, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return io.NopCloser(strings.NewReader(m.content)), nil
}

func TestGetStorages(t *testing.T) {
	t.Run("list storages", func(t *testing.T) {
		mock := &mockAdapterV2{}
		storages := map[string]adapter.Adapter{
			"local": mock,
			"s3":    mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages", nil)
		w := httptest.NewRecorder()

		server.GetStorages(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var response struct {
			Storages []string `json:"storages"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(response.Storages) != 2 {
			t.Errorf("expected 2 storages, got %d", len(response.Storages))
		}
	})
}

func TestGetStoragesStorageNodesPath_DirectoryListing(t *testing.T) {
	t.Run("list root directory", func(t *testing.T) {
		mockNodes := []adapter.FileNode{
			{
				Path:     "subdir",
				Type:     "dir",
				Basename: "subdir",
			},
			{
				Path:         "file.txt",
				Type:         "file",
				Basename:     "file.txt",
				Extension:    "txt",
				Size:         1024,
				LastModified: 1234567890,
				MimeType:     "text/plain",
			},
		}

		mock := &mockAdapterV2{nodes: mockNodes}
		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var response NodeList
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Storage != "local" {
			t.Errorf("expected storage 'local', got '%s'", response.Storage)
		}

		if len(response.Files) != 2 {
			t.Errorf("expected 2 files (direct children), got %d", len(response.Files))
		}

		// Check dirname has storage prefix
		expectedDirname := "local://"
		if response.Dirname != expectedDirname {
			t.Errorf("expected dirname '%s', got '%s'", expectedDirname, response.Dirname)
		}
	})
}

func TestNotImplementedOperations(t *testing.T) {
	mock := &mockAdapterV2{}
	storages := map[string]adapter.Adapter{
		"local": mock,
	}

	server, err := NewServer(storages, "local")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "DeleteStoragesStorageNodesPath",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.DeleteStoragesStorageNodesPath(w, r, "local", "test", DeleteStoragesStorageNodesPathParams{})
			},
		},
		{
			name: "PatchStoragesStorageNodesPath",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.PatchStoragesStorageNodesPath(w, r, "local", "test")
			},
		},
		{
			name: "PostStoragesStorageNodesPath",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.PostStoragesStorageNodesPath(w, r, "local", "test")
			},
		},
		{
			name: "PostStoragesStorageCopies",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.PostStoragesStorageCopies(w, r, "local")
			},
		},
		{
			name: "PostStoragesStorageMoves",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.PostStoragesStorageMoves(w, r, "local")
			},
		},
		{
			name: "GetStoragesStorageArchives",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.GetStoragesStorageArchives(w, r, "local", GetStoragesStorageArchivesParams{})
			},
		},
		{
			name: "PostStoragesStorageArchives",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.PostStoragesStorageArchives(w, r, "local", PostStoragesStorageArchivesParams{})
			},
		},
		{
			name: "PostStoragesStorageArchivesPath",
			handler: func(w http.ResponseWriter, r *http.Request) {
				server.PostStoragesStorageArchivesPath(w, r, "local", "test.zip")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusNotImplemented {
				t.Errorf("expected status 501, got %d", resp.StatusCode)
			}

			var errorResp ErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}

			if errorResp.Status != false {
				t.Errorf("expected error status false, got %v", errorResp.Status)
			}

			if !strings.Contains(errorResp.Message, "Not Implemented") {
				t.Errorf("expected message containing 'Not Implemented', got '%s'", errorResp.Message)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	t.Run("valid server creation", func(t *testing.T) {
		mock := &mockAdapterV2{}
		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if server == nil {
			t.Fatal("expected server to be non-nil")
		}
	})

	t.Run("invalid default storage", func(t *testing.T) {
		mock := &mockAdapterV2{}
		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		_, err := NewServer(storages, "nonexistent")
		if err == nil {
			t.Fatal("expected error for invalid default storage")
		}
	})

	t.Run("empty default storage is allowed", func(t *testing.T) {
		mock := &mockAdapterV2{}
		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if server == nil {
			t.Fatal("expected server to be non-nil")
		}
	})
}

// TestGetStoragesStorageNodesPath_FileContent tests serving file content with the GetStoragesStorageNodesPath handler
func TestGetStoragesStorageNodesPath_FileContent(t *testing.T) {
	t.Run("serve text file", func(t *testing.T) {
		content := "Hello, World!"
		mock := &mockAdapterV2{
			content:  content,
			mimeType: "text/plain",
			size:     int64(len(content)),
		}

		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/test.txt", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "test.txt", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// For file content, headers should reflect the file type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "text/plain" {
			t.Errorf("expected Content-Type 'text/plain', got '%s'", contentType)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != content {
			t.Errorf("expected body '%s', got '%s'", content, string(body))
		}
	})

	t.Run("serve binary file", func(t *testing.T) {
		binaryContent := "\x89PNG\r\n\x1a\n" // PNG header
		mock := &mockAdapterV2{
			content:  binaryContent,
			mimeType: "image/png",
			size:     int64(len(binaryContent)),
		}

		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/image.png", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "image.png", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != binaryContent {
			t.Errorf("binary content mismatch")
		}
	})

	t.Run("serve large file", func(t *testing.T) {
		largeContent := strings.Repeat("Lorem ipsum dolor sit amet. ", 1000)
		mock := &mockAdapterV2{
			content:  largeContent,
			mimeType: "text/plain",
			size:     int64(len(largeContent)),
		}

		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/large.txt", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "large.txt", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != largeContent {
			t.Errorf("content length mismatch: expected %d, got %d", len(largeContent), len(body))
		}
	})

	t.Run("mime type detection error", func(t *testing.T) {
		mock := &mockAdapterV2{
			mimeTypeErr: http.ErrNotSupported,
		}

		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/test.txt", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "test.txt", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("file size error", func(t *testing.T) {
		mock := &mockAdapterV2{
			mimeType: "text/plain",
			sizeErr:  http.ErrNotSupported,
		}

		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/test.txt", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "test.txt", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("read stream error", func(t *testing.T) {
		mock := &mockAdapterV2{
			mimeType: "text/plain",
			size:     100,
			readErr:  http.ErrNotSupported,
		}

		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/test.txt", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "test.txt", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("adapter does not support reading", func(t *testing.T) {
		// Create an adapter that doesn't implement Reader
		type nonReaderAdapter struct{}

		storages := map[string]adapter.Adapter{
			"local": &nonReaderAdapter{},
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/test.txt", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		server.GetStoragesStorageNodesPath(w, req, "local", "test.txt", GetStoragesStorageNodesPathParams{})

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("with download parameter", func(t *testing.T) {
		content := "File content for download"
		mock := &mockAdapterV2{
			content:  content,
			mimeType: "text/plain",
			size:     int64(len(content)),
		}

		storages := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(storages, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/storages/local/nodes/document.txt", nil)
		req.Header.Set("Accept", "application/octet-stream")
		w := httptest.NewRecorder()

		downloadTrue := true
		server.GetStoragesStorageNodesPath(w, req, "local", "document.txt", GetStoragesStorageNodesPathParams{Download: &downloadTrue})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		contentDisposition := resp.Header.Get("Content-Disposition")
		if contentDisposition == "" {
			t.Error("expected Content-Disposition header to be set")
		}
		if !strings.Contains(contentDisposition, "attachment") {
			t.Errorf("expected attachment disposition, got '%s'", contentDisposition)
		}
	})
}
