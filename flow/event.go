package flow

import (
	"strings"
	"time"
)

const (
	statusSuccess = "SUCCESS"
)

// Event is Cloud Build events published to Cloud Pub/Sub
type Event struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"projectId"`
	Status     string     `json:"status"`
	Timeout    string     `json:"timeout"`
	LogURL     string     `json:"logUrl"`
	StartTime  *time.Time `json:"startTime"`
	FinishTime *time.Time `json:"finishTime"`

	EventSource `json:"source"`
	Artifacts   `json:"artifacts"`
}

type EventSource struct {
	EventRepo `json:"repoSource"`
}

type EventRepo struct {
	RepoName string `json:"repoName"`
	TagName  string `json:"tagName"`
}

type Artifacts struct {
	Images []string `json:"images"`
}

func (e Event) isFinished() bool {
	return (e.FinishTime != nil)
}

func (e Event) isSuuccess() bool {
	return (e.Status == statusSuccess)
}

func (e Event) isApplicationBuild() bool {
	return (e.RepoName != cfg.ManifestName)
}

func (e Event) getAppName() string {
	// @todo trim organization
	return strings.Replace(e.RepoName, "github-com", "", 1)
}
