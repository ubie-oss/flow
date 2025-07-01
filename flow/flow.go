package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sakajunquality/cloud-pubsub-events/gcrevent"
)

var (
	cfg *Config
)

type Flow struct {
	Env                   string
	useApp                bool
	githubToken           *string
	githubAppID           *int64
	githubAppInstlationID *int64
	githubAppPrivateKey   *string
	enableVersionQuote    bool
	enableAutoMerge       bool
	maxRetries            int
}

func New(c *Config) (*Flow, error) {
	cfg = c
	f := &Flow{}

	githubToken := os.Getenv("FLOW_GITHUB_TOKEN")
	githubAppID := os.Getenv("FLOW_GITHUB_APP_ID")
	githubAppInstlationID := os.Getenv("FLOW_GITHUB_APP_INSTALLATION_ID")
	githubAppPrivateKey := os.Getenv("FLOW_GITHUB_APP_PRIVATE_KEY")
	f.enableVersionQuote = os.Getenv("FLOW_ENABLE_VERSION_QUOTE") == "true"
	f.enableAutoMerge = os.Getenv("FLOW_ENABLE_AUTO_MERGE") == "true"
	f.githubToken = &githubToken

	// Set maxRetries: config file > environment variable > default (3)
	f.maxRetries = 3
	if c.MaxRetries > 0 {
		f.maxRetries = c.MaxRetries
	}
	if maxRetriesEnv := os.Getenv("FLOW_MAX_RETRIES"); maxRetriesEnv != "" {
		if maxRetriesInt, err := strconv.Atoi(maxRetriesEnv); err == nil && maxRetriesInt > 0 {
			f.maxRetries = maxRetriesInt
		}
	}

	if githubAppID != "" {
		f.useApp = true

		githubAppIDInt, err := strconv.ParseInt(githubAppID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid value for FLOW_GITHUB_APP_ID")
		}
		f.githubAppID = &githubAppIDInt

		githubAppInstlationIDInt, err := strconv.ParseInt(githubAppInstlationID, 10, 64)
		if err != nil {
			return nil, errors.New("invalid value for FLOW_GITHUB_APP_INSTALLATION_ID")
		}
		f.githubAppInstlationID = &githubAppInstlationIDInt

		f.githubAppPrivateKey = &githubAppPrivateKey
	}

	if !f.useApp && f.githubToken == nil {
		return nil, errors.New("you need to specify a non-empty value for FLOW_GITHUB_TOKEN if you don't specify FLOW_GITHUB_APP_ID")
	}

	return f, nil
}

func (f *Flow) ProcessGCREvent(ctx context.Context, e gcrevent.Event) error {
	if e.Action != gcrevent.ActionInsert {
		return nil
	}

	if e.Tag == nil {
		return nil
	}

	parts := strings.Split(*e.Tag, ":")
	if len(parts) < 2 {
		return errors.New("invalid image tag or missing version")
	}
	image, version := parts[0], parts[1]

	if image == "" || version == "" {
		return fmt.Errorf("image format invalid: %s", *e.Tag)
	}

	return f.processImage(ctx, image, version)
}
