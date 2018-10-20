package flow

import "time"

// Event is Cloud Build events published to Cloud Pub/Sub
type Event struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"projectId"`
	Status     string    `json:"status"`
	Timeout    string    `json:"timeout"`
	LogURL     string    `json:"logUrl"`
	StartTime  time.Time `json:"startTime"`
	FinishTime time.Time `json:"finishTime"`

	EventSource `json:"source"`
}

type EventSource struct {
	EventRepo `json:"repoSource"`
}

type EventRepo struct {
	RepoName string `json:"repoName"`
	TagName  string `json:"tagName"`
}
