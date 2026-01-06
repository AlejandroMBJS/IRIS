/*
Package main - IRIS Payroll System Alternative Server Entry Point

==============================================================================
FILE: cmd/server/main.go
==============================================================================

DESCRIPTION:
    This is an alternative/minimal entry point for the IRIS Payroll System.
    It provides a simplified server setup for testing or development purposes.
    The primary entry point is cmd/api/main.go - use that for production.

USER PERSPECTIVE:
    - This is a development/testing server
    - Only provides a basic health check endpoint
    - NOT recommended for production use

DEVELOPER GUIDELINES:
    ⚠️  DEPRECATED - Use cmd/api/main.go instead
    ✅  Can be used for: Quick testing, minimal deployments
    ❌  Missing: Authentication, full API routes, graceful shutdown

SYNTAX EXPLANATION:
    - http.NewServeMux(): Creates a new HTTP request multiplexer (router)
    - mux.HandleFunc(): Registers a handler function for a URL pattern
    - http.Server{}: Struct literal to configure server settings
    - server.ListenAndServe(): Starts HTTP server and blocks until error

==============================================================================
*/
package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/internal/config"
)

func main() {
	log.Println("IRIS Payroll Backend")
	log.Println("====================")

	// Load complete application configuration
	appConfig, err := config.LoadAppConfig("./configs")
	if err != nil {
		log.Fatal("Error loading configuration:", err)
	}

	log.Println("✅ Configuration loaded successfully")

	// Initialize and start the web server
	initServer(appConfig)
}

// initServer sets up the routes and starts the HTTP server.
func initServer(appConfig *config.AppConfig) {
	// Basic router setup
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/api/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
	})

	// Server configuration
	addr := ":" + fmt.Sprintf("%d", appConfig.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("🚀 Starting server on port %d...", appConfig.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", addr, err)
	}
}
