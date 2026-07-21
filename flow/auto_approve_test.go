package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v75/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectAutoApproveMarker(t *testing.T) {
	tests := []struct {
		name string
		body string
		want autoApproveKind
	}{
		{name: "exact normal marker only", body: autoApproveMarker, want: autoApproveNormal},
		{name: "exact wait marker only", body: autoApproveWaitForE2EMarker, want: autoApproveWaitForE2E},
		{name: "normal marker at end of body", body: "Release notes\n\n" + autoApproveMarker + "\n", want: autoApproveNormal},
		{name: "wait marker at end of body", body: "Release notes\n\n" + autoApproveWaitForE2EMarker + "\n", want: autoApproveWaitForE2E},
		{name: "both markers prefers wait", body: autoApproveMarker + "\n" + autoApproveWaitForE2EMarker, want: autoApproveWaitForE2E},
		{name: "both markers reverse order prefers wait", body: autoApproveWaitForE2EMarker + "\n" + autoApproveMarker, want: autoApproveWaitForE2E},
		{name: "empty body", body: "", want: autoApproveNone},
		{name: "partial without html comment", body: "ubie:auto-approve", want: autoApproveNone},
		{name: "similar wrong marker", body: "<!-- ubie:auto-approve-all -->", want: autoApproveNone},
		{name: "missing closing", body: "<!-- ubie:auto-approve", want: autoApproveNone},
		{name: "different marker", body: "<!-- auto-approve -->", want: autoApproveNone},
		{name: "unrelated body", body: "normal release notes", want: autoApproveNone},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, detectAutoApproveMarker(tc.body))
		})
	}
}

func TestContainsAutoApproveMarker(t *testing.T) {
	assert.True(t, containsAutoApproveMarker(autoApproveMarker))
	assert.False(t, containsAutoApproveMarker(autoApproveWaitForE2EMarker))
	assert.False(t, containsAutoApproveMarker(""))
}

func TestMaybePropagateAutoApproveLabel(t *testing.T) {
	const (
		owner = "wonderland"
		repo  = "alice"
		tag   = "v1.2.3"
	)
	baseLabels := []string{"alice", "production"}

	tests := []struct {
		name       string
		status     int
		body       *string
		wantLabels []string
	}{
		{
			name:       "normal marker present adds normal label",
			status:     http.StatusOK,
			body:       github.Ptr("notes\n" + autoApproveMarker + "\n"),
			wantLabels: []string{"alice", "production", cloudDeployAutoApproveLabel},
		},
		{
			name:       "wait marker present adds wait label only",
			status:     http.StatusOK,
			body:       github.Ptr("notes\n" + autoApproveWaitForE2EMarker + "\n"),
			wantLabels: []string{"alice", "production", cloudDeployAutoApproveWaitForE2ELabel},
		},
		{
			name:       "both markers prefers wait label only",
			status:     http.StatusOK,
			body:       github.Ptr(autoApproveMarker + "\n" + autoApproveWaitForE2EMarker),
			wantLabels: []string{"alice", "production", cloudDeployAutoApproveWaitForE2ELabel},
		},
		{
			name:       "marker absent keeps labels",
			status:     http.StatusOK,
			body:       github.Ptr("normal release notes"),
			wantLabels: baseLabels,
		},
		{
			name:       "nil body keeps labels",
			status:     http.StatusOK,
			body:       nil,
			wantLabels: baseLabels,
		},
		{
			name:       "empty body keeps labels",
			status:     http.StatusOK,
			body:       github.Ptr(""),
			wantLabels: baseLabels,
		},
		{
			name:       "release not found keeps labels",
			status:     http.StatusNotFound,
			wantLabels: baseLabels,
		},
		{
			name:       "api error keeps labels",
			status:     http.StatusInternalServerError,
			wantLabels: baseLabels,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, tag), func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tc.status)
				if tc.status != http.StatusOK {
					_, _ = w.Write([]byte(`{"message":"error"}`))
					return
				}
				rel := &github.RepositoryRelease{TagName: github.Ptr(tag), Body: tc.body}
				require.NoError(t, json.NewEncoder(w).Encode(rel))
			})
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			client := github.NewClient(server.Client())
			baseURL, err := url.Parse(server.URL + "/")
			require.NoError(t, err)
			client.BaseURL = baseURL

			got := maybePropagateAutoApproveLabel(context.Background(), client, owner, repo, tag, append([]string{}, baseLabels...))
			assert.Equal(t, tc.wantLabels, got)
		})
	}
}
