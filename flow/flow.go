package flow

import (
	"context"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
)

const (
	pubsubTopicID = "cloudbuild"
	subName       = "cloudbuild-flow-sub"
)

var (
	pubsubClient *pubsub.Client
	subscription *pubsub.Subscription
	cfg          *Config
)

type Flow struct {
	Env           string
	slackBotToken string
	githubToken   string
}

func New(ctx context.Context, c *Config) (*Flow, error) {
	cfg = c
	f := &Flow{
		Env:           os.Getenv("FLOW_ENV"),
		slackBotToken: os.Getenv("FLOW_SLACK_BOT_TOKEN"),
		githubToken:   os.Getenv("FLOW_GITHUB_TOKEN"),
	}

	if f.Env == "" || f.slackBotToken == "" || f.githubToken == "" {
		return nil, errors.New("You need to specify a non-empty value for FLOW_ENV, FLOW_SLACK_BOT_TOKEN and FLOW_GITHUB_TOKEN")
	}

	return f, nil
}

func (f *Flow) Start() {
	ctx := context.Background()

	// Create Cloud Pub/Sub topic if not exist
	topic := pubsubClient.Topic(pubsubTopicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for topic: %v.\n", err)

	}
	if !exists {
		if _, err := pubsubClient.CreateTopic(ctx, pubsubTopicID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create topic: %v.\n", err)
		}
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

	go f.subscribe()
}
