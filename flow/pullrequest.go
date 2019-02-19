package flow

type PullRequests []PullRequest

type PullRequest struct {
	env string
	url string
	err error
}
