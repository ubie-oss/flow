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

func TestContainsAutoApproveMarker(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "exact marker only", body: autoApproveMarker, want: true},
		{name: "marker at end of body", body: "Release notes\n\n" + autoApproveMarker + "\n", want: true},
		{name: "marker in middle", body: "before\n" + autoApproveMarker + "\nafter", want: true},
		{name: "empty body", body: "", want: false},
		{name: "partial without html comment", body: "ubie:auto-approve", want: false},
		{name: "similar wrong marker", body: "<!-- ubie:auto-approve-all -->", want: false},
		{name: "missing closing", body: "<!-- ubie:auto-approve", want: false},
		{name: "different marker", body: "<!-- auto-approve -->", want: false},
		{name: "unrelated body", body: "normal release notes", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, containsAutoApproveMarker(tc.body))
		})
	}
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
			name:       "marker present adds label",
			status:     http.StatusOK,
			body:       github.Ptr("notes\n" + autoApproveMarker + "\n"),
			wantLabels: []string{"alice", "production", cloudDeployAutoApproveLabel},
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
