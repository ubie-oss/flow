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
	// autoApproveWaitForE2EMarker is the wait-for-e2e variant (gh-release -a -w).
	// Exact match only — do not loosen this string.
	autoApproveWaitForE2EMarker = "<!-- ubie:auto-approve-wait-for-e2e -->"
	// cloudDeployAutoApproveLabel is applied to the releases PR when the normal marker is present.
	// Downstream (github-pr-finder / cloudbuild-utils) consumes this label.
	cloudDeployAutoApproveLabel = "clouddeploy-auto-approve"
	// cloudDeployAutoApproveWaitForE2ELabel is applied when the wait-for-e2e marker is present.
	cloudDeployAutoApproveWaitForE2ELabel = "clouddeploy-auto-approve-wait-for-e2e"
)

type autoApproveKind int

const (
	autoApproveNone autoApproveKind = iota
	autoApproveNormal
	autoApproveWaitForE2E
)

// detectAutoApproveMarker reports which auto-approve marker is present.
// Markers are mutually exclusive; if both appear, prefer wait-for-e2e (safer).
func detectAutoApproveMarker(body string) autoApproveKind {
	hasWait := strings.Contains(body, autoApproveWaitForE2EMarker)
	hasNormal := strings.Contains(body, autoApproveMarker)
	if hasWait {
		return autoApproveWaitForE2E
	}
	if hasNormal {
		return autoApproveNormal
	}
	return autoApproveNone
}

// containsAutoApproveMarker reports whether body contains the exact normal auto-approve marker.
// Matching the full marker string avoids false positives from partial substrings.
func containsAutoApproveMarker(body string) bool {
	return detectAutoApproveMarker(body) == autoApproveNormal
}

// maybePropagateAutoApproveLabel fetches the source GitHub Release for tag and, when its
// body contains an auto-approve marker, appends the corresponding label.
// Wait marker → clouddeploy-auto-approve-wait-for-e2e only.
// Normal marker → clouddeploy-auto-approve only.
// Any failure (API error, missing release, empty body, no marker) is logged and ignored
// so image rewrite / commit / PR creation continues without the label.
func maybePropagateAutoApproveLabel(ctx context.Context, client *github.Client, owner, repo, tag string, labels []string) []string {
	kind := detectAutoApproveMarkerInRelease(ctx, client, owner, repo, tag)
	switch kind {
	case autoApproveWaitForE2E:
		return append(labels, cloudDeployAutoApproveWaitForE2ELabel)
	case autoApproveNormal:
		return append(labels, cloudDeployAutoApproveLabel)
	default:
		return labels
	}
}

func detectAutoApproveMarkerInRelease(ctx context.Context, client *github.Client, owner, repo, tag string) autoApproveKind {
	rel, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		slog.Warn("Failed to get GitHub Release for auto-approve marker; continuing without label",
			"owner", owner, "repo", repo, "tag", tag, "error", err)
		return autoApproveNone
	}
	if rel == nil {
		slog.Info("GitHub Release is nil for auto-approve marker; continuing without label",
			"owner", owner, "repo", repo, "tag", tag)
		return autoApproveNone
	}

	// GetBody returns "" when Body is nil.
	kind := detectAutoApproveMarker(rel.GetBody())
	switch kind {
	case autoApproveWaitForE2E:
		slog.Info("Found auto-approve-wait-for-e2e marker in Release body; adding label",
			"owner", owner, "repo", repo, "tag", tag, "label", cloudDeployAutoApproveWaitForE2ELabel)
	case autoApproveNormal:
		slog.Info("Found auto-approve marker in Release body; adding label",
			"owner", owner, "repo", repo, "tag", tag, "label", cloudDeployAutoApproveLabel)
	default:
		slog.Info("auto-approve marker not present in Release body; continuing without label",
			"owner", owner, "repo", repo, "tag", tag)
	}
	return kind
}
