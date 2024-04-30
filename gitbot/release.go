package gitbot

import (
	"context"
	"errors"

	"github.com/dlclark/regexp2"
	"github.com/google/go-github/v61/github"
)

type release struct {
	repo              Repo
	author            Author
	message           string
	body              string
	labels            []string
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
		repo:              repo,
		author:            author,
		message:           message,
		body:              body,
		labels:            labels,
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

func (r *release) GetRepo() *Repo            { return &r.repo }
func (r *release) SetRepo(repo Repo)         { r.repo = repo }
func (r *release) GetAuthor() *Author        { return &r.author }
func (r *release) SetAuthor(author Author)   { r.author = author }
func (r *release) GetMessage() string        { return r.message }
func (r *release) SetMessage(s string)       { r.message = s }
func (r *release) GetBody() string           { return r.body }
func (r *release) SetBody(s string)          { r.body = s }
func (r *release) GetLabels() []string       { return r.labels }
func (r *release) SetLabels(labels []string) { r.labels = labels }
