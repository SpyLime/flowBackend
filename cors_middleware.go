package main

import (
	"fmt"
	"net/http"
)

// buildCORSMiddleware creates a middleware that adds CORS headers to all responses
func buildCORSMiddleware() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log the request for debugging
			fmt.Printf("CORS middleware: %s %s\n", r.Method, r.URL.String())
			fmt.Printf("Request headers: %v\n", r.Header)

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-XSRF-TOKEN")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler
			handler.ServeHTTP(w, r)
		})
	}
}
