package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/sakajunquality/cloud-pubsub-events/gcrevent"
	"github.com/ubie-oss/flow/v4/flow"

	"gopkg.in/yaml.v3"
)

// Response is a HTTP response
type Response struct {
	Status int `json:"status"`
}

// PubSubMessage is a Push message from Cloud Pub/Sub
type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

var (
	f *flow.Flow
)

func main() {
	// Configure slog with JSON handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := getConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read the file:%s.\n", err)
		os.Exit(1)
	}

	f, err = initFlow(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing the config %s.\n", err)
		os.Exit(1)
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Post("/", handlePubSubMessage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("Starting server", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func getConfig() ([]byte, error) {
	return os.ReadFile(os.Getenv("FLOW_CONFIG_PATH"))
}

func initFlow(config []byte) (*flow.Flow, error) {
	cfg := new(flow.Config)
	if err := yaml.Unmarshal(config, cfg); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal error: %w", err)
	}
	f, err := flow.New(cfg)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func handlePubSubMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var m PubSubMessage
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		slog.Error("Failed to unmarshal PubSub message", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	event, err := gcrevent.ParseMessage(m.Message.Data)
	if err != nil {
		slog.Error("Failed to parse GCR event", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = f.ProcessGCREvent(ctx, event)
	if err != nil {
		slog.Error("Failed to process GCR event", "error", err)
	}

	res := &Response{
		Status: http.StatusOK,
	}
	render.JSON(w, r, res)
}
