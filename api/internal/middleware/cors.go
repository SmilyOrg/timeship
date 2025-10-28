package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

// CORS creates a CORS middleware handler with configuration from environment variables
func CORS() func(http.Handler) http.Handler {
	// Get allowed origins from environment variable
	allowedOrigins := os.Getenv("TIMESHIP_CORS_ALLOWED_ORIGINS")
	var origins []string

	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		// Trim whitespace from each origin
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	} else {
		// Default to localhost if not set
		origins = []string{"http://localhost:8080"}
	}

	// Create CORS handler with configuration
	c := cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		MaxAge: 300, // Maximum value not ignored by any of major browsers
	})

	return c.Handler
}
