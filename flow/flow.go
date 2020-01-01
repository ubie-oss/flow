package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/sakajunquality/cloud-pubsub-events/gcrevent"
)

const (
	pubsubTopicID = "cloud-builds"
	subName       = "cloudbuild-flow-sub"
)

var (
	subscription *pubsub.Subscription
	cfg          *Config
)

type Flow struct {
	Env           string
	projectID     string
	slackBotToken string
	githubToken   string
}

func New(c *Config) (*Flow, error) {
	cfg = c
	f := &Flow{
		projectID:     os.Getenv("FLOW_GCP_PROJECT_ID"),
		slackBotToken: os.Getenv("FLOW_SLACK_BOT_TOKEN"),
		githubToken:   os.Getenv("FLOW_GITHUB_TOKEN"),
	}

	if f.projectID == "" || f.slackBotToken == "" || f.githubToken == "" {
		return nil, errors.New("You need to specify a non-empty value for FLOW_GCP_PROJECT_ID, FLOW_SLACK_BOT_TOKEN and FLOW_GITHUB_TOKEN")
	}

	return f, nil
}

func (f *Flow) Start(ctx context.Context, errCh chan error) {
	pubsubClient, err := pubsub.NewClient(ctx, f.projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pubsub client: %v.\n", err)
	}

	// Create Cloud Pub/Sub topic if not exist
	topic := pubsubClient.Topic(pubsubTopicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for topic: %v.\n", err)

	}

	// Create topic subscription
	subscription = pubsubClient.Subscription(subName)
	exists, err = subscription.Exists(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for subscription: %v.\n", err)
	}
	if !exists {
		if _, err = pubsubClient.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create subscription: %v.\n", err)
		}
	}

	go f.subscribe(ctx, errCh)
}

func (f *Flow) ProcessGCREvent(ctx context.Context, e gcrevent.Event) error {
	if e.Action != gcrevent.ActionInsert {
		return nil
	}

	parts := strings.Split(*e.Tag, ":")
	image, version := parts[0], parts[1]

	if image == "" || version == "" {
		return fmt.Errorf("Image format invalid: %s", *e.Tag)
	}

	return f.processImage(ctx, image, version)
}
