package gitbot

import (
	"context"
	"log/slog"
	"os"
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

func NewGitHubClientWithApp(ctx context.Context, appID, installationID int64, privateKey string) *github.Client {
	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, appID, installationID, []byte(privateKey))
	if err != nil {
		slog.Error("Failed to create GitHub installation transport", "error", err)
		os.Exit(1)
	}
	return github.NewClient(&http.Client{Transport: itr})
}
