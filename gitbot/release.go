package gitbot

import (
	"context"
	"errors"

	"github.com/google/go-github/v29/github"
)

type Release struct {
	Repo
	Author
	Changes []Change
	Message string
	Body    string
	Labels  []string
}

type Repo struct {
	SourceOwner  string
	SourceRepo   string
	BaseBranch   string
	CommitBranch string
}

type Author struct {
	Name  string
	Email string
}

type Change struct {
	filePath    string
	regexText   string
	changedText string
}

func (r *Release) AddChanges(filePath, regexText, changedText string) {
	r.Changes = append(r.Changes, Change{
		filePath:    filePath,
		regexText:   regexText,
		changedText: changedText,
	})
}

func (r *Release) Commit(ctx context.Context, client *github.Client) error {
	ref, err := r.getRef(ctx, client)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("git reference was nil ")
	}

	tree, err := r.getTree(ctx, client, ref)
	if err != nil {
		return err
	}

	return r.pushCommit(ctx, client, ref, tree)
}

func (r *Release) CreatePR(ctx context.Context, client *github.Client) (*string, error) {
	return r.createPR(ctx, client)
}
