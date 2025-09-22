package gitbot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v61/github"

	"golang.org/x/oauth2"
)

func NewGitHubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func NewGitHubClientWithApp(ctx context.Context, appID, installationID int64, privateKey string) (*github.Client, error) {
	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, appID, installationID, []byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub installation transport: %w", err)
	}
	return github.NewClient(&http.Client{Transport: itr}), nil
}
