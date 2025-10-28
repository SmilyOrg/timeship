package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/smilyorg/timeship/api/internal/adapter"
	"github.com/smilyorg/timeship/api/internal/adapter/local"
)

// TestPreviewIntegration tests the preview operation with the local adapter
func TestPreviewIntegration(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test file with known content
	testContent := "This is a test file for preview operation.\nLine 2.\nLine 3."
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create local adapter
	localAdapter, err := local.New(tempDir)
	if err != nil {
		t.Fatalf("failed to create local adapter: %v", err)
	}
	defer localAdapter.Close()

	// Create server with the local adapter
	adapters := map[string]adapter.Adapter{
		"local": localAdapter,
	}

	server, err := NewServer(adapters, "local")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	t.Run("preview existing file", func(t *testing.T) {
		path := "local://test.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Check Content-Type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "text/plain; charset=utf-8" {
			t.Errorf("expected Content-Type 'text/plain; charset=utf-8', got '%s'", contentType)
		}

		// Read and verify content
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != testContent {
			t.Errorf("content mismatch:\nexpected: %q\ngot: %q", testContent, string(body))
		}
	})

	t.Run("preview non-existent file", func(t *testing.T) {
		path := "local://nonexistent.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://nonexistent.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})

	t.Run("preview binary file", func(t *testing.T) {
		// Create a binary file
		binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
		binaryFile := filepath.Join(tempDir, "test.png")
		if err := os.WriteFile(binaryFile, binaryContent, 0644); err != nil {
			t.Fatalf("failed to create binary file: %v", err)
		}

		path := "local://test.png"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://test.png", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Check Content-Type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "image/png" {
			t.Errorf("expected Content-Type 'image/png', got '%s'", contentType)
		}

		// Read and verify content
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if len(body) != len(binaryContent) {
			t.Errorf("content length mismatch: expected %d, got %d", len(binaryContent), len(body))
		}

		for i := range binaryContent {
			if body[i] != binaryContent[i] {
				t.Errorf("byte mismatch at position %d: expected %02x, got %02x", i, binaryContent[i], body[i])
			}
		}
	})

	t.Run("preview file in subdirectory", func(t *testing.T) {
		// Create subdirectory
		subDir := filepath.Join(tempDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}

		// Create file in subdirectory
		subContent := "File in subdirectory"
		subFile := filepath.Join(subDir, "nested.txt")
		if err := os.WriteFile(subFile, []byte(subContent), 0644); err != nil {
			t.Fatalf("failed to create file in subdirectory: %v", err)
		}

		path := "local://subdir/nested.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://subdir/nested.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != subContent {
			t.Errorf("content mismatch: expected %q, got %q", subContent, string(body))
		}
	})

	t.Run("prevent path traversal in preview", func(t *testing.T) {
		// Try to access a file outside the root using path traversal
		path := "local://../etc/passwd"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://../etc/passwd", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		// Should fail because os.OpenRoot prevents path traversal
		if resp.StatusCode == http.StatusOK {
			t.Error("path traversal should have been prevented")
		}
	})
}
