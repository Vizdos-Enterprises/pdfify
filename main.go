package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	endpoints "github.com/vizdos-enterprises/pdfify/internal"
	"github.com/vizdos-enterprises/pdfify/internal/generate"
	internal_s3 "github.com/vizdos-enterprises/pdfify/internal/s3"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	internal_s3.Connect()

	// Initialize the shared Chrome context
	generate.InitChrome()

	mux := http.NewServeMux()
	endpoints.Route(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", 9999),
		Handler: mux,
	}

	// Channel to listen for interrupt or terminate signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Channel to signal that the server has shut down
	shutdownChan := make(chan struct{})

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
		}
		close(shutdownChan)
	}()

	// Block until we receive a signal
	sig := <-sigChan
	log.Printf("Received signal %s, initiating shutdown...\n", sig)

	// Create a context with a timeout to allow the server to shut down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v\n", err)
	} else {
		log.Println("Server stopped gracefully")
	}

	// Shutdown Chrome context
	generate.ShutdownChrome()

	// Wait for the server to shut down
	<-shutdownChan
	log.Println("Server exiting")
}
