package flow

import (
	"time"
)

const (
	statusQueued  = "QUEUED"
	statusWorking = "WORKING"
	statusSuccess = "SUCCESS"
	statusFailure = "FAILURE"
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

	TriggerID string `json:"buildTriggerId"`

	EventSource `json:"source"`
	Artifacts   `json:"artifacts"`
}

type EventSource struct {
	EventRepo `json:"repoSource"`
}

type EventRepo struct {
	RepoName   string  `json:"repoName"`
	TagName    *string `json:"tagName"`
	BranchName *string `json:"branchName"`
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

func (e Event) isTriggerdBuld() bool {
	return (e.TriggerID != "")
}
