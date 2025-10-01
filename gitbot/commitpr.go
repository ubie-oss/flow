package gitbot

import (
	"context"
	"log/slog"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/google/go-github/v75/github"
)

func (r *release) getRef(ctx context.Context, client *github.Client) (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(ctx, r.repo.SourceOwner, r.repo.SourceRepo, "refs/heads/"+r.repo.CommitBranch); err == nil {
		return ref, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(ctx, r.repo.SourceOwner, r.repo.SourceRepo, "refs/heads/"+r.repo.BaseBranch); err != nil {
		return nil, err
	}
	newRef := github.CreateRef{Ref: "refs/heads/" + r.repo.CommitBranch, SHA: *baseRef.Object.SHA}
	ref, _, err = client.Git.CreateRef(ctx, r.repo.SourceOwner, r.repo.SourceRepo, newRef)
	return ref, err
}

func (r *release) makeChange(ctx context.Context, client *github.Client, filePath, regexText string, evaluator regexp2.MatchEvaluator) {
	// rewrite if target is already changed
	content, ok := r.changedContentMap[filePath]
	if ok {
		r.changedContentMap[filePath] = getChangedText(content, regexText, evaluator)
		return
	}

	content, err := r.getOriginalContent(ctx, client, filePath, r.repo.BaseBranch)
	if err != nil {
		slog.Error("Error fetching content", "error", err)
		return
	}

	changed := getChangedText(content, regexText, evaluator)
	r.changedContentMap[filePath] = changed
}

func (r *release) getTree(ctx context.Context, client *github.Client, ref *github.Reference) (*github.Tree, error) {
	entries := []*github.TreeEntry{}
	for path, content := range r.changedContentMap {
		entries = append(entries, &github.TreeEntry{Path: github.Ptr(path), Type: github.Ptr("blob"), Content: github.Ptr(content), Mode: github.Ptr("100644")})
	}

	tree, _, err := client.Git.CreateTree(ctx, r.repo.SourceOwner, r.repo.SourceRepo, *ref.Object.SHA, entries)
	return tree, err
}

func (r *release) pushCommit(ctx context.Context, client *github.Client, ref *github.Reference, tree *github.Tree) error {
	parent, _, err := client.Repositories.GetCommit(ctx, r.repo.SourceOwner, r.repo.SourceRepo, *ref.Object.SHA, nil)
	if err != nil {
		return err
	}

	parent.Commit.SHA = parent.SHA

	date := time.Now()
	author := &github.CommitAuthor{Date: &github.Timestamp{Time: date}, Name: &r.author.Name, Email: &r.author.Email}
	commit := &github.Commit{Author: author, Message: &r.message, Tree: tree, Parents: []*github.Commit{parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(ctx, r.repo.SourceOwner, r.repo.SourceRepo, *commit, nil)
	if err != nil {
		return err
	}

	ref.Object.SHA = newCommit.SHA
	updateRef := github.UpdateRef{SHA: *newCommit.SHA, Force: github.Ptr(false)}
	_, _, err = client.Git.UpdateRef(ctx, r.repo.SourceOwner, r.repo.SourceRepo, *ref.Ref, updateRef)
	return err
}

func (r *release) createPR(ctx context.Context, client *github.Client) (*string, error) {
	newPR := &github.NewPullRequest{
		Title:               github.Ptr(r.message),
		Head:                github.Ptr(r.repo.CommitBranch),
		Base:                github.Ptr(r.repo.BaseBranch),
		Body:                github.Ptr(r.body),
		MaintainerCanModify: github.Ptr(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, r.repo.SourceOwner, r.repo.SourceRepo, newPR)
	if err != nil {
		return nil, err
	}

	err = r.addLabels(ctx, client, *pr.Number)
	if err != nil {
		slog.Error("Error adding labels", "error", err)
	}

	return github.Ptr(pr.GetHTMLURL()), nil
}

func (r *release) addLabels(ctx context.Context, client *github.Client, prNumber int) error {
	_, _, err := client.Issues.AddLabelsToIssue(ctx, r.repo.SourceOwner, r.repo.SourceRepo, prNumber, r.labels)
	return err
}

func (r *release) getOriginalContent(ctx context.Context, client *github.Client, filePath, baseBranch string) (string, error) {
	opt := &github.RepositoryContentGetOptions{
		Ref: baseBranch,
	}

	f, _, _, err := client.Repositories.GetContents(ctx, r.repo.SourceOwner, r.repo.SourceRepo, filePath, opt)

	if err != nil {
		return "", err
	}

	return f.GetContent()
}

func getChangedText(original, regex string, evaluator regexp2.MatchEvaluator) string {
	re := regexp2.MustCompile(regex, 0)
	result, err := re.ReplaceFunc(original, evaluator, 0, -1)

	if err != nil {
		return original
	}

	return result
}
