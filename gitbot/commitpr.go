package gitbot

import (
	"context"
	"log"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/google/go-github/v29/github"
)

func (r *release) getRef(ctx context.Context, client *github.Client) (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(ctx, r.SourceOwner, r.SourceRepo, "refs/heads/"+r.CommitBranch); err == nil {
		return ref, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(ctx, r.SourceOwner, r.SourceRepo, "refs/heads/"+r.BaseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + r.CommitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(ctx, r.SourceOwner, r.SourceRepo, newRef)
	return ref, err
}

func (r *release) makeChange(ctx context.Context, client *github.Client, filePath, regexText string, evaluator regexp2.MatchEvaluator) {
	// rewrite if target is already changed
	content, ok := r.changedContentMap[filePath]
	if ok {
		r.changedContentMap[filePath] = getChangedText(content, regexText, evaluator)
		return
	}

	content, err := r.getOriginalContent(ctx, client, filePath, r.Repo.BaseBranch)
	if err != nil {
		log.Printf("Error fetching content %s", err)
		return
	}

	changed := getChangedText(content, regexText, evaluator)
	r.changedContentMap[filePath] = changed
}

func (r *release) getTree(ctx context.Context, client *github.Client, ref *github.Reference) (*github.Tree, error) {
	entries := []github.TreeEntry{}
	for path, content := range r.changedContentMap {
		entries = append(entries, github.TreeEntry{Path: github.String(path), Type: github.String("blob"), Content: github.String(content), Mode: github.String("100644")})
	}

	tree, _, err := client.Git.CreateTree(ctx, r.SourceOwner, r.SourceRepo, *ref.Object.SHA, entries)
	return tree, err
}

func (r *release) pushCommit(ctx context.Context, client *github.Client, ref *github.Reference, tree *github.Tree) error {
	parent, _, err := client.Repositories.GetCommit(ctx, r.SourceOwner, r.SourceRepo, *ref.Object.SHA)
	if err != nil {
		return err
	}

	parent.Commit.SHA = parent.SHA

	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &r.Author.Name, Email: &r.Author.Email}
	commit := &github.Commit{Author: author, Message: &r.Message, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(ctx, r.SourceOwner, r.SourceRepo, commit)
	if err != nil {
		return err
	}

	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(ctx, r.SourceOwner, r.SourceRepo, ref, false)
	return err
}

func (r *release) createPR(ctx context.Context, client *github.Client) (*string, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(r.Message),
		Head:                github.String(r.CommitBranch),
		Base:                github.String(r.BaseBranch),
		Body:                github.String(r.Body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, r.SourceOwner, r.SourceRepo, newPR)
	if err != nil {
		return nil, err
	}

	err = r.addLabels(ctx, client, *pr.Number)
	if err != nil {
		log.Printf("Error Adding Lables: %s", err)
	}

	return github.String(pr.GetHTMLURL()), nil
}

func (r *release) addLabels(ctx context.Context, client *github.Client, prNumber int) error {
	_, _, err := client.Issues.AddLabelsToIssue(ctx, r.SourceOwner, r.SourceRepo, prNumber, r.Labels)
	return err
}

func (r *release) getOriginalContent(ctx context.Context, client *github.Client, filePath, baseBranch string) (string, error) {
	opt := &github.RepositoryContentGetOptions{
		Ref: baseBranch,
	}

	f, _, _, err := client.Repositories.GetContents(ctx, r.SourceOwner, r.SourceRepo, filePath, opt)

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
