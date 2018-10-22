package gitbot

import (
	"context"
	"errors"

	"github.com/google/go-github/v18/github"
	"golang.org/x/oauth2"
)

type Release struct {
	ctx context.Context
	Repo
	Author
	PullRequest
	Changes []Change
}

type Repo struct {
	sourceOwner string
	sourceRepo  string
	baseBranch  string
}

type PullRequest struct {
	commitBranch  string
	commitMessage string
	prTitle       string
}

type Author struct {
	authorName  string
	authorEmail string
}

type Change struct {
	file          string
	beforeReplace string
	afterReplace  string
}

const (
	baseBranch = "master"
)

var client *github.Client

// NewRelease is ...
func NewRelease() *Release {
	return &Release{}
}

func (r *Release) AddChanges() {

}

func (r *Release) Create(ctx context.Context) (*string, error) {
	token := "hogehoge"

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)

	ref, err := r.getRef()
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, errors.New("git reference was nil ")
	}

	tree, err := r.getTree(ref)
	if err != nil {
		return nil, err
	}

	if err := r.pushCommit(ref, tree); err != nil {
		return nil, err
	}

	return r.createPR()
}
