package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/sakajunquality/cloud-pubsub-events/gcrevent"
	"github.com/ubie-oss/flow/v4/flow"

	"gopkg.in/yaml.v3"
)

// Response represents a HTTP response
type Response struct {
	Status int `json:"status"`
}

// PubSubMessage represents a Push message from Cloud Pub/Sub
type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

var (
	// flowInstance is the global Flow instance
	flowInstance *flow.Flow
	logger       = log.New(os.Stderr, "flow: ", log.LstdFlags|log.Lshortfile)
)

func main() {
	// Create a context that will be canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize configuration
	cfg, err := getConfig()
	if err != nil {
		logger.Fatalf("Failed to read configuration: %v", err)
	}

	// Initialize Flow instance
	flowInstance, err = initFlow(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize Flow: %v", err)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup HTTP server
	server := setupServer(port)

	// Start server in a goroutine
	go func() {
		logger.Printf("Starting server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Error during server shutdown: %v", err)
	}
}

// getConfig reads the configuration file specified by FLOW_CONFIG_PATH
func getConfig() ([]byte, error) {
	configPath := os.Getenv("FLOW_CONFIG_PATH")
	if configPath == "" {
		return nil, fmt.Errorf("FLOW_CONFIG_PATH environment variable not set")
	}
	return os.ReadFile(configPath)
}

// initFlow initializes a new Flow instance with the given configuration
func initFlow(config []byte) (*flow.Flow, error) {
	cfg := new(flow.Config)
	if err := yaml.Unmarshal(config, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}
	return flow.New(cfg)
}

// setupServer configures and returns a new HTTP server
func setupServer(port string) *http.Server {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Add routes
	r.Post("/", handlePubSubMessage)

	return &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
}

// handlePubSubMessage handles incoming Pub/Sub messages
func handlePubSubMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var m PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	event, err := gcrevent.ParseMessage(m.Message.Data)
	if err != nil {
		logger.Printf("Failed to parse Pub/Sub message: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := flowInstance.ProcessGCREvent(ctx, event); err != nil {
		logger.Printf("Failed to process GCR event: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, &Response{
		Status: http.StatusOK,
	})
}
