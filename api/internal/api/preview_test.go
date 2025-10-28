package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smilyorg/timeship/api/internal/adapter"
)

// mockReaderAdapter implements adapter.Reader for testing
type mockReaderAdapter struct {
	content  string
	mimeType string
	size     int64
	err      error
}

func (m *mockReaderAdapter) ReadStream(path string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(strings.NewReader(m.content)), nil
}

func (m *mockReaderAdapter) FileSize(path string) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.size, nil
}

func (m *mockReaderAdapter) MimeType(path string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.mimeType, nil
}

func TestGetPreview(t *testing.T) {
	t.Run("successful preview", func(t *testing.T) {
		content := "Hello, World!"
		mock := &mockReaderAdapter{
			content:  content,
			mimeType: "text/plain",
			size:     int64(len(content)),
		}

		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "local://test.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		// Check Content-Type header
		contentType := resp.Header.Get("Content-Type")
		if contentType != "text/plain" {
			t.Errorf("expected Content-Type 'text/plain', got '%s'", contentType)
		}

		// Check Content-Length header
		contentLength := resp.Header.Get("Content-Length")
		if contentLength == "" {
			t.Error("expected Content-Length header to be set")
		}

		// Read and verify content
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != content {
			t.Errorf("expected body '%s', got '%s'", content, string(body))
		}
	})

	t.Run("missing path parameter", func(t *testing.T) {
		mock := &mockReaderAdapter{}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview})

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("empty path parameter", func(t *testing.T) {
		mock := &mockReaderAdapter{}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		emptyPath := ""
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &emptyPath})

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid adapter", func(t *testing.T) {
		mock := &mockReaderAdapter{}
		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "invalid://test.txt"
		invalidAdapter := Adapter("invalid")
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&adapter=invalid&path=invalid://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Adapter: &invalidAdapter, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("adapter does not support reading", func(t *testing.T) {
		// Create an adapter that doesn't implement Reader
		type nonReaderAdapter struct{}

		adapters := map[string]adapter.Adapter{
			"local": &nonReaderAdapter{},
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "local://test.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Errorf("expected status 501, got %d", resp.StatusCode)
		}
	})

	t.Run("mime type error", func(t *testing.T) {
		mock := &mockReaderAdapter{
			err: fmt.Errorf("mime type detection failed"),
		}

		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "local://test.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})

	t.Run("file size error", func(t *testing.T) {
		// Mock that succeeds for MimeType but fails for FileSize
		customReader := &customMockReader{
			mimeTypeResult: "text/plain",
			fileSizeError:  fmt.Errorf("size detection failed"),
		}

		adapters := map[string]adapter.Adapter{
			"local": customReader,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "local://test.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})

	t.Run("read stream error", func(t *testing.T) {
		// Mock that succeeds for MimeType and FileSize but fails for ReadStream
		mock := &customMockReader{
			mimeTypeResult:  "text/plain",
			fileSizeResult:  100,
			readStreamError: fmt.Errorf("failed to open file"),
		}

		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "local://test.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})

	t.Run("binary content", func(t *testing.T) {
		// Test with binary content (e.g., image)
		binaryContent := "\x89PNG\r\n\x1a\n" // PNG header
		mock := &mockReaderAdapter{
			content:  binaryContent,
			mimeType: "image/png",
			size:     int64(len(binaryContent)),
		}

		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "local://image.png"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://image.png", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType != "image/png" {
			t.Errorf("expected Content-Type 'image/png', got '%s'", contentType)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != binaryContent {
			t.Errorf("binary content mismatch")
		}
	})

	t.Run("large file", func(t *testing.T) {
		// Test with a larger file
		largeContent := strings.Repeat("Lorem ipsum dolor sit amet. ", 1000)
		mock := &mockReaderAdapter{
			content:  largeContent,
			mimeType: "text/plain",
			size:     int64(len(largeContent)),
		}

		adapters := map[string]adapter.Adapter{
			"local": mock,
		}

		server, err := NewServer(adapters, "local")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "local://large.txt"
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&path=local://large.txt", nil)
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

		if string(body) != largeContent {
			t.Errorf("content length mismatch: expected %d, got %d", len(largeContent), len(body))
		}
	})

	t.Run("with custom adapter parameter", func(t *testing.T) {
		content := "Custom adapter content"
		mock := &mockReaderAdapter{
			content:  content,
			mimeType: "text/plain",
			size:     int64(len(content)),
		}

		adapters := map[string]adapter.Adapter{
			"custom": mock,
		}

		server, err := NewServer(adapters, "custom")
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		path := "custom://test.txt"
		adapterParam := Adapter("custom")
		req := httptest.NewRequest(http.MethodGet, "/api/?q=preview&adapter=custom&path=custom://test.txt", nil)
		w := httptest.NewRecorder()

		server.Get(w, req, GetParams{Q: GetParamsQPreview, Adapter: &adapterParam, Path: &path})

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != content {
			t.Errorf("expected body '%s', got '%s'", content, string(body))
		}
	})
}

// customMockReader allows selective errors for different methods
type customMockReader struct {
	mimeTypeResult  string
	mimeTypeError   error
	fileSizeResult  int64
	fileSizeError   error
	content         string
	readStreamError error
}

func (m *customMockReader) MimeType(path string) (string, error) {
	return m.mimeTypeResult, m.mimeTypeError
}

func (m *customMockReader) FileSize(path string) (int64, error) {
	return m.fileSizeResult, m.fileSizeError
}

func (m *customMockReader) ReadStream(path string) (io.ReadCloser, error) {
	if m.readStreamError != nil {
		return nil, m.readStreamError
	}
	return io.NopCloser(strings.NewReader(m.content)), nil
}
