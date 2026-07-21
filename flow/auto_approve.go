package flow

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/go-github/v75/github"
)

const (
	// autoApproveMarker is an opt-in marker embedded in a source GitHub Release body
	// (e.g. by gh-release --auto-approve). Exact match only — do not loosen this string.
	autoApproveMarker = "<!-- ubie:auto-approve -->"
	// cloudDeployAutoApproveLabel is applied to the releases PR when the marker is present.
	// Downstream (github-pr-finder / cloudbuild-utils) consumes this label.
	cloudDeployAutoApproveLabel = "clouddeploy-auto-approve"
)

// containsAutoApproveMarker reports whether body contains the exact auto-approve marker.
// Matching the full marker string avoids false positives from partial substrings.
func containsAutoApproveMarker(body string) bool {
	return strings.Contains(body, autoApproveMarker)
}

// maybePropagateAutoApproveLabel fetches the source GitHub Release for tag and, when its
// body contains the auto-approve marker, appends clouddeploy-auto-approve to labels.
// Any failure (API error, missing release, empty body, no marker) is logged and ignored
// so image rewrite / commit / PR creation continues without the label.
func maybePropagateAutoApproveLabel(ctx context.Context, client *github.Client, owner, repo, tag string, labels []string) []string {
	if !hasAutoApproveMarkerInRelease(ctx, client, owner, repo, tag) {
		return labels
	}
	return append(labels, cloudDeployAutoApproveLabel)
}

func hasAutoApproveMarkerInRelease(ctx context.Context, client *github.Client, owner, repo, tag string) bool {
	rel, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		slog.Warn("Failed to get GitHub Release for auto-approve marker; continuing without label",
			"owner", owner, "repo", repo, "tag", tag, "error", err)
		return false
	}
	if rel == nil {
		slog.Info("GitHub Release is nil for auto-approve marker; continuing without label",
			"owner", owner, "repo", repo, "tag", tag)
		return false
	}

	// GetBody returns "" when Body is nil.
	if !containsAutoApproveMarker(rel.GetBody()) {
		slog.Info("auto-approve marker not present in Release body; continuing without label",
			"owner", owner, "repo", repo, "tag", tag)
		return false
	}

	slog.Info("Found auto-approve marker in Release body; adding label",
		"owner", owner, "repo", repo, "tag", tag, "label", cloudDeployAutoApproveLabel)
	return true
}
