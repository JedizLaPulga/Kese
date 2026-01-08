package kese

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RunWithShutdown starts the HTTP server with graceful shutdown support.
// It listens for interrupt signals (SIGINT, SIGTERM) and gracefully shuts down the server,
// allowing ongoing requests to complete within the specified timeout.
//
// address: Server address in format ":8080" or "localhost:8080"
// timeout: Maximum time to wait for ongoing requests to complete
//
// Example:
//
//	app.RunWithShutdown(":8080", 10*time.Second)
func (a *App) RunWithShutdown(address string, timeout time.Duration) error {
	server := &http.Server{
		Addr:    address,
		Handler: a,
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		fmt.Printf("🚀 Kese server starting on %s (with graceful shutdown)\n", address)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		fmt.Printf("\n🛑 Received signal %v, starting graceful shutdown...\n", sig)

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			// Force shutdown if graceful shutdown fails
			server.Close()
			return fmt.Errorf("failed to gracefully shutdown server: %w", err)
		}

		fmt.Println("✅ Server stopped gracefully")
		return nil
	}
}
