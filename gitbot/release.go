package gitbot

import (
	"context"
	"errors"
	"fmt"

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
	filePath    string
	regexText   string
	changedText string
}

var client *github.Client

func NewRepo(sourceOwner, sourceRepo, baseBranch string) *Repo {
	return &Repo{
		sourceOwner: sourceOwner,
		sourceRepo:  sourceRepo,
		baseBranch:  baseBranch,
	}
}

// NewRelease is ...
func NewRelease(repo Repo, appEnv, appVersion string) *Release {
	branch := fmt.Sprintf("release/%s-%s-%s", repo.sourceRepo, appEnv, appVersion)
	subject := fmt.Sprintf("Release %s %s %s", repo.sourceRepo, appEnv, appVersion)

	return &Release{
		Repo: repo,
		PullRequest: PullRequest{
			commitBranch:  branch,
			commitMessage: subject,
			prTitle:       subject,
		},
	}
}

func (r *Release) AddAuthor(authorName, authorEmail string) {
	r.Author.authorName = authorName
	r.Author.authorEmail = authorEmail
}

func (r *Release) AddChanges(filePath, regexText, changedText string) {
	r.Changes = append(r.Changes, Change{
		filePath:    filePath,
		regexText:   regexText,
		changedText: changedText,
	})
}

func (r *Release) Create(ctx context.Context, token string) (*string, error) {
	r.ctx = ctx
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
