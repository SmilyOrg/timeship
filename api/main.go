package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/smilyorg/timeship/api/internal/adapter"
	"github.com/smilyorg/timeship/api/internal/adapter/local"
	"github.com/smilyorg/timeship/api/internal/api"
	"github.com/smilyorg/timeship/api/internal/middleware"
)

func main() {
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

	log.Printf("VueFinder API serving files from: %s", rootDir)

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

	// Create API server
	server := api.NewServer(adapters)

	// Create HTTP server with generated handler
	mux := http.NewServeMux()
	handler := api.HandlerWithOptions(server, api.StdHTTPServerOptions{
		BaseURL: "/api",
	})

	// Apply CORS middleware
	corsHandler := middleware.CORS()
	mux.Handle("/", corsHandler(handler))

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

	// Start server in a goroutine
	go func() {
		log.Printf("Starting VueFinder API server on http://%s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
