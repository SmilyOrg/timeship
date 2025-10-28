# Go VueFinder Service Implementation Plan

## Overview
Implement a Go service in the `api/` directory using stdlib http libraries and [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to match the existing PHP VueFinder API.

## Project Structure

```
api/
├── api.yaml (existing)
├── VueFinder.php (existing)
├── go.mod
├── go.sum
├── tools.go (oapi-codegen dependency)
├── oapi-codegen.yaml (generator config)
├── main.go (server entry point)
└── internal/
    ├── api/
    │   ├── generate.go (go:generate directive)
    │   ├── api.gen.go (oapi-codegen output)
    │   └── api.go (ServerInterface implementation)
    ├── adapter/
    │   ├── adapter.go (interface)
    │   └── local/
    │       └── local.go (local filesystem implementation)
    └── middleware/
        └── cors.go (CORS middleware)
```

## Phase 1: Project Setup

### Task 1: Initialize Go Module
**Status:** Not Started

Create `api/go.mod`:
```bash
cd api
go mod init github.com/smilyorg/timeship/api
```

### Task 2: Create tools.go
**Status:** Not Started

Create `api/tools.go`:
```go
//go:build tools
// +build tools

package main

import (
    _ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
```

Then run:
```bash
go mod tidy
```

### Task 3: Create Directory Structure
**Status:** Not Started

```bash
mkdir -p internal/api
mkdir -p internal/adapter/local
mkdir -p internal/middleware
```

## Phase 2: Code Generation

### Task 4: Create oapi-codegen Configuration
**Status:** Not Started

Create `api/oapi-codegen.yaml`:
```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
output: internal/api/api.gen.go
generate:
  std-http-server: true
  models: true
```

### Task 5: Create generate.go
**Status:** Not Started

Create `api/internal/api/generate.go`:
```go
package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config ../../oapi-codegen.yaml ../../api.yaml
```

### Task 6: Generate Code
**Status:** Not Started

Run:
```bash
go generate ./...
```

This will create `api/internal/api/api.gen.go` containing:
- `ServerInterface` - interface you'll implement
- All request/response types matching the OpenAPI spec
- `HandlerWithOptions()` - function to wire up your implementation

## Phase 3: Filesystem Adapter (Local Only)

### Task 7: Define Adapter Interface
**Status:** Not Started

Create `api/internal/adapter/adapter.go`:
```go
package adapter

import (
    "io"
)

// FileNode represents a file or directory
type FileNode struct {
    Path         string
    Type         string // "file" or "dir"
    Basename     string
    Extension    string
    Size         int64
    LastModified int64
    MimeType     string
}

// Adapter defines filesystem operations
type Adapter interface {
    // ListContents returns files and directories at the given path
    ListContents(path string) ([]FileNode, error)
    
    // ReadStream opens a file for reading
    ReadStream(path string) (io.ReadCloser, error)
    
    // WriteStream writes data to a file
    WriteStream(path string, r io.Reader) error
    
    // Delete removes a file
    Delete(path string) error
    
    // DeleteDirectory removes a directory and all contents
    DeleteDirectory(path string) error
    
    // Move renames/moves a file or directory
    Move(from, to string) error
    
    // CreateDirectory creates a new directory
    CreateDirectory(path string) error
    
    // FileSize returns the size of a file in bytes
    FileSize(path string) (int64, error)
    
    // LastModified returns the Unix timestamp of last modification
    LastModified(path string) (int64, error)
    
    // MimeType detects and returns the MIME type of a file
    MimeType(path string) (string, error)
    
    // FileExists checks if a file exists
    FileExists(path string) (bool, error)
    
    // DirectoryExists checks if a directory exists
    DirectoryExists(path string) (bool, error)
}
```

### Task 8: Implement Local Filesystem Adapter
**Status:** Not Started

Create `api/internal/adapter/local/local.go`:
```go
package local

import (
    "io"
    "io/fs"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/smilyorg/timeship/api/internal/adapter"
)

// Adapter implements adapter.Adapter for local filesystem
type Adapter struct {
    rootPath string
}

// New creates a new local filesystem adapter
func New(rootPath string) (*Adapter, error) {
    // Ensure root path exists
    info, err := os.Stat(rootPath)
    if err != nil {
        return nil, err
    }
    if !info.IsDir() {
        return nil, fs.ErrInvalid
    }
    
    return &Adapter{
        rootPath: rootPath,
    }, nil
}

// resolvePath converts "local://path" to absolute filesystem path
func (a *Adapter) resolvePath(path string) string {
    // Remove "local://" prefix if present
    path = strings.TrimPrefix(path, "local://")
    return filepath.Join(a.rootPath, path)
}

// ListContents implements Adapter.ListContents
func (a *LocalAdapter) ListContents(path string) ([]FileNode, error) {
    absPath := a.resolvePath(path)
    
    entries, err := os.ReadDir(absPath)
    if err != nil {
        return nil, err
    }
    
    nodes := make([]FileNode, 0, len(entries))
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue // Skip files we can't stat
        }
        
        node := FileNode{
            Path:     filepath.Join(path, entry.Name()),
            Basename: entry.Name(),
        }
        
        if entry.IsDir() {
            node.Type = "dir"
        } else {
            node.Type = "file"
            node.Extension = strings.TrimPrefix(filepath.Ext(entry.Name()), ".")
            node.Size = info.Size()
            node.LastModified = info.ModTime().Unix()
            
            // Detect MIME type
            if node.Extension != "" {
                mimeType, _ := a.MimeType(node.Path)
                node.MimeType = mimeType
            }
        }
        
        nodes = append(nodes, node)
    }
    
    return nodes, nil
}

// MimeType implements Adapter.MimeType
func (a *LocalAdapter) MimeType(path string) (string, error) {
    absPath := a.resolvePath(path)
    
    file, err := os.Open(absPath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    // Read first 512 bytes for MIME detection
    buffer := make([]byte, 512)
    n, err := file.Read(buffer)
    if err != nil && err != io.EOF {
        return "", err
    }
    
    return http.DetectContentType(buffer[:n]), nil
}

// FileExists implements Adapter.FileExists
func (a *LocalAdapter) FileExists(path string) (bool, error) {
    absPath := a.resolvePath(path)
    info, err := os.Stat(absPath)
    if os.IsNotExist(err) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return !info.IsDir(), nil
}

// DirectoryExists implements Adapter.DirectoryExists
func (a *LocalAdapter) DirectoryExists(path string) (bool, error) {
    absPath := a.resolvePath(path)
    info, err := os.Stat(absPath)
    if os.IsNotExist(err) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return info.IsDir(), nil
}

// Implement remaining methods as stubs for now:
// - ReadStream
// - WriteStream
// - Delete
// - DeleteDirectory
// - Move
// - CreateDirectory
// - FileSize
// - LastModified
```

## Phase 4: Implement /index Endpoint

### Task 9: Create Handler Implementation
**Status:** Not Started

Create `api/internal/api/api.go`:
```go
package api

import (
    "encoding/json"
    "net/http"
    "sort"
    
    "github.com/smilyorg/timeship/api/internal/adapter"
)

// Handler implements ServerInterface
type Handler struct {
    adapters map[string]adapter.Adapter
}

// NewHandler creates a new handler with the given adapters
func NewHandler(adapters map[string]adapter.Adapter) *Handler {
    return &Handler{
        adapters: adapters,
    }
}

// GetIndex implements the /index endpoint
func (h *Handler) GetIndex(w http.ResponseWriter, r *http.Request) {
    // Get query parameters
    adapterKey := r.URL.Query().Get("adapter")
    if adapterKey == "" {
        adapterKey = "local" // default
    }
    
    path := r.URL.Query().Get("path")
    if path == "" {
        path = adapterKey + "://"
    }
    
    // Get adapter
    adptr, ok := h.adapters[adapterKey]
    if !ok {
        http.Error(w, "Invalid adapter", http.StatusBadRequest)
        return
    }
    
    // List contents
    nodes, err := adptr.ListContents(path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Sort: directories first, then files
    sort.Slice(nodes, func(i, j int) bool {
        if nodes[i].Type != nodes[j].Type {
            return nodes[i].Type == "dir"
        }
        return nodes[i].Basename < nodes[j].Basename
    })
    
    // Convert to API response format
    files := make([]FileNode, len(nodes))
    for i, node := range nodes {
        files[i] = FileNode{
            Path:     node.Path,
            Type:     FileType(node.Type),
            Basename: node.Basename,
            Storage:  adapterKey,
        }
        
        if node.Type == "file" {
            files[i].Extension = &node.Extension
            files[i].Size = &node.Size
            files[i].LastModified = &node.LastModified
            if node.MimeType != "" {
                files[i].MimeType = &node.MimeType
            }
        }
    }
    
    response := DirectoryListingResponse{
        Adapter:  adapterKey,
        Storages: []string{adapterKey}, // Just local for now
        Dirname:  path,
        Files:    files,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## Phase 5: Stub Remaining Endpoints

### Task 10: Add Stub Implementations
**Status:** Not Started

Add to `api/internal/api/api.go`:
```go
// Stub implementations returning 501 Not Implemented

func (h *Handler) GetSubfolders(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) GetSearch(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) GetPreview(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) GetDownload(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostNewfolder(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostNewfile(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostUpload(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostSave(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostRename(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostMove(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostDelete(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostArchive(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func (h *Handler) PostUnarchive(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not Implemented", http.StatusNotImplemented)
}
```

## Phase 6: CORS Middleware

### Task 11: Implement CORS Middleware
**Status:** Not Started

Create `api/internal/middleware/cors.go`:
```go
package middleware

import "net/http"

// CORS wraps an http.Handler with CORS headers
func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        
        // Handle preflight OPTIONS request
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        // Call next handler
        next.ServeHTTP(w, r)
    })
}
```

## Phase 7: Main Server

### Task 12: Create Main Server
**Status:** Not Started

Create `api/main.go`:
```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/smilyorg/timeship/api/internal/adapter/local"
    "github.com/smilyorg/timeship/api/internal/api"
    "github.com/smilyorg/timeship/api/internal/middleware"
)

func main() {
    // Create local adapter
    localAdapter, err := local.New("./storage")
    if err != nil {
        log.Fatalf("Failed to create local adapter: %v", err)
    }
    
    adapters := map[string]adapter.Adapter{
        "local": localAdapter,
    }
    
    // Create handler
    h := api.NewHandler(adapters)
    
    // Create HTTP handler with generated code
    mux := http.NewServeMux()
    apiHandler := api.HandlerWithOptions(h, api.StdHTTPServerOptions{
        BaseRouter: mux,
        BaseURL:    "/api",
    })
    
    // Wrap with CORS middleware
    finalHandler := middleware.CORS(apiHandler)
    
    // Create server
    server := &http.Server{
        Addr:         ":8080",
        Handler:      finalHandler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server in goroutine
    go func() {
        log.Printf("Server starting on %s", server.Addr)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Server shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server stopped")
}
```

## Phase 8: Testing

### Task 13: Manual Testing
**Status:** Not Started

1. Create test storage directory:
   ```bash
   mkdir -p api/storage
   echo "test" > api/storage/test.txt
   mkdir api/storage/subdir
   ```

2. Run the server:
   ```bash
   cd api
   go run main.go
   ```

3. Test the /index endpoint:
   ```bash
   curl http://localhost:8080/api/index?adapter=local&path=local://
   ```

4. Verify response structure matches OpenAPI spec:
   - Check `adapter` field
   - Check `storages` array
   - Check `dirname` field
   - Check `files` array with proper FileNode structure
   - Verify directories appear before files
   - Verify metadata (size, timestamp, mime type) for files

5. Test CORS headers:
   ```bash
   curl -i -X OPTIONS http://localhost:8080/api/index
   ```
   
   Should see:
   - `Access-Control-Allow-Origin: *`
   - `Access-Control-Allow-Headers: *`

## Future Enhancements

After the initial implementation is working:

1. Implement remaining endpoints one by one
2. Add proper error handling and validation
3. Add request logging middleware
4. Add unit tests for handlers and adapter
5. Add integration tests
6. Consider adding configuration file support
7. Add support for public URL generation
8. Implement HTTP range requests for streaming
9. Add ZIP archive support
10. Consider adding other adapters (S3, etc.) if needed

## Notes

- All code lives in `api/` directory (monorepo structure)
- Using Go 1.22+ features (pattern matching in http.ServeMux via generated code)
- No external routing frameworks - pure stdlib
- oapi-codegen handles request parameter extraction
- Focus on getting /index working first, then expand
- Generated code will be in `api/internal/api/api.gen.go` - don't edit it manually
