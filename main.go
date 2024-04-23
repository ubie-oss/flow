package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/sakajunquality/cloud-pubsub-events/gcrevent"
	"github.com/sakajunquality/flow/v4/flow"

	"gopkg.in/yaml.v3"

	_ "github.com/GoogleCloudPlatform/berglas/pkg/auto"
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

var (
	q *bool
)

func main() {
	q = flag.Bool("q", true, "if true, the version is quoted. Default is true.")
	flag.Parse()

	cfg, err := getConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cloud not read the file:%s.\n", err)
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
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getConfig() ([]byte, error) {
	return os.ReadFile(os.Getenv("FLOW_CONFIG_PATH"))
}

func initFlow(config []byte) (*flow.Flow, error) {
	cfg := new(flow.Config)
	if err := yaml.Unmarshal(config, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "yaml.Unmarshal error:%v.\n", err)
		os.Exit(1)
	}
	return flow.New(cfg)
}

func handlePubSubMessage(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	var m PubSubMessage
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("iotuil.ReadAll: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	event, err := gcrevent.ParseMessage(m.Message.Data)
	if err != nil {
		log.Printf("gcrevent.ParseMessage: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = f.ProcessGCREvent(ctx, event, *q)
	log.Printf("process: %s", err)

	res := &Response{
		Status: http.StatusOK,
	}
	render.JSON(w, r, res)
}
