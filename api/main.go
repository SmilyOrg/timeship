package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lpar/gzipped"
	"github.com/smilyorg/timeship/api/internal/adapter"
	"github.com/smilyorg/timeship/api/internal/adapter/local"
	"github.com/smilyorg/timeship/api/internal/api"
	"github.com/smilyorg/timeship/api/internal/middleware"
	"github.com/smilyorg/timeship/api/internal/network"
)

//go:generate go tool oapi-codegen -config oapi-codegen.yaml api.yaml

func main() {
	godotenv.Load()

	// Get the root directory for the local adapter from environment or use current directory
	rootDir := os.Getenv("TIMESHIP_ROOT")
	if rootDir == "" {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
	}

	// Create local adapter
	localAdapter, err := local.New(rootDir)
	if err != nil {
		log.Fatalf("Failed to create local adapter: %v", err)
	}
	defer localAdapter.Close()

	log.Printf("Timeship API serving files from: %s", rootDir)

	// Create adapters map
	adapters := map[string]adapter.Adapter{
		"local": localAdapter,
	}

	// Ensure adapters are closed on exit
	defer func() {
		for name, adapterInstance := range adapters {
			if closer, ok := adapterInstance.(io.Closer); ok {
				if err := closer.Close(); err != nil {
					log.Printf("Error closing adapter %s: %v", name, err)
				}
			}
		}
	}()

	// Create API server (local is the default adapter)
	server, err := api.NewServer(adapters, "local")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Get API prefix from environment or use default
	apiPrefix := os.Getenv("TIMESHIP_API_PREFIX")
	if apiPrefix == "" {
		apiPrefix = "/api"
	}

	// Create HTTP server with routing
	mux := http.NewServeMux()

	// API routes with CORS
	handler := api.HandlerWithOptions(server, api.StdHTTPServerOptions{})
	corsHandler := middleware.CORS()(handler)

	// Mount API, stripping prefix if not at root
	if apiPrefix == "/" {
		mux.Handle("/", corsHandler)
	} else {
		mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, corsHandler))
	}

	// Serve embedded UI if available (when built with -tags embedui)
	if apiPrefix != "/" {
		// Try to read from embedded FS to check if UI is available
		_, err := StaticFs.Open("ui/dist")
		if err != nil {
			log.Printf("UI not embedded, serving API only (build with -tags embedui to embed UI)")
		} else {
			// Hardcode well-known mime types, see https://github.com/golang/go/issues/32350
			mime.AddExtensionType(".js", "text/javascript")
			mime.AddExtensionType(".css", "text/css")
			mime.AddExtensionType(".html", "text/html")
			mime.AddExtensionType(".woff", "font/woff")
			mime.AddExtensionType(".woff2", "font/woff2")
			mime.AddExtensionType(".png", "image/png")
			mime.AddExtensionType(".jpg", "image/jpg")
			mime.AddExtensionType(".jpeg", "image/jpeg")
			mime.AddExtensionType(".ico", "image/vnd.microsoft.icon")
			mime.AddExtensionType(".svg", "image/svg+xml")
			mime.AddExtensionType(".webmanifest", "application/manifest+json")

			uifs, err := fs.Sub(StaticFs, "ui/dist")
			if err != nil {
				panic(err)
			}
			uihandler := gzipped.FileServer(
				middleware.SpaFs{
					Root: http.FS(uifs),
				},
			)

			// Create UI mux with middleware
			uiMux := http.NewServeMux()
			uiMux.Handle("/", uihandler)

			// Wrap with cache control and index.html middleware
			uiHandler := middleware.CacheControl()(middleware.IndexHTML()(uiMux))
			mux.Handle("/", uiHandler)

			log.Printf("Serving embedded UI at /")
		}
	}

	// Get server address from environment or use default
	addr := os.Getenv("TIMESHIP_ADDRESS")
	if addr == "" {
		addr = ":8080"
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create listener to get actual address
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	// Start server in a goroutine
	go func() {
		log.Printf("")
		log.Println("Starting Timeship server")
		if err := network.PrintListenURLs(listener.Addr(), apiPrefix); err != nil {
			log.Printf("Warning: couldn't list all network addresses: %v", err)
			log.Printf("  API: http://%s%s", addr, apiPrefix)
		}
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}
