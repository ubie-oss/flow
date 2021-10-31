package gitbot

import (
	"context"
	"errors"

	"github.com/dlclark/regexp2"
	"github.com/google/go-github/v29/github"
)

type release struct {
	Repo
	Author
	Message           string
	Body              string
	Labels            []string
	changedContentMap map[string]string
}

type Release interface {
	MakeChange(ctx context.Context, client *github.Client, filePath, regexText, changedText string)
	MakeChangeFunc(ctx context.Context, client *github.Client, filePath, regexText string, evaluator regexp2.MatchEvaluator)
	Commit(ctx context.Context, client *github.Client) error
	CreatePR(ctx context.Context, client *github.Client) (*string, error)

	GetRepo() *Repo
	SetRepo(repo Repo)
	GetAuthor() *Author
	SetAuthor(author Author)
	GetMessage() string
	SetMessage(string)
	GetBody() string
	SetBody(string)
	GetLabels() []string
	SetLabels([]string)
}

type Repo struct {
	SourceOwner  string
	SourceRepo   string
	BaseBranch   string
	CommitBranch string
}

var _ Release = &release{}

type Author struct {
	Name  string
	Email string
}

func NewRelease(repo Repo, author Author, message string, body string, labels []string) Release {
	return &release{
		Repo:              repo,
		Author:            author,
		Message:           message,
		Body:              body,
		Labels:            labels,
		changedContentMap: make(map[string]string),
	}
}

func (r *release) MakeChange(ctx context.Context, client *github.Client, filePath, regexText, changedText string) {
	r.makeChange(ctx, client, filePath, regexText, func(regexp2.Match) string { return changedText })
}

func (r *release) MakeChangeFunc(ctx context.Context, client *github.Client, filePath, regexText string, evaluator regexp2.MatchEvaluator) {
	r.makeChange(ctx, client, filePath, regexText, evaluator)
}

func (r *release) Commit(ctx context.Context, client *github.Client) error {
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

func (r *release) CreatePR(ctx context.Context, client *github.Client) (*string, error) {
	return r.createPR(ctx, client)
}

func (r *release) GetRepo() *Repo            { return &r.Repo }
func (r *release) SetRepo(repo Repo)         { r.Repo = repo }
func (r *release) GetAuthor() *Author        { return &r.Author }
func (r *release) SetAuthor(author Author)   { r.Author = author }
func (r *release) GetMessage() string        { return r.Message }
func (r *release) SetMessage(s string)       { r.Message = s }
func (r *release) GetBody() string           { return r.Body }
func (r *release) SetBody(s string)          { r.Body = s }
func (r *release) GetLabels() []string       { return r.Labels }
func (r *release) SetLabels(labels []string) { r.Labels = labels }
