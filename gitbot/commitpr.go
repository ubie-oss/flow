package gitbot

import (
	"regexp"
	"time"

	"github.com/google/go-github/v18/github"
)

func (r *Release) getRef() (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(r.ctx, r.sourceOwner, r.sourceRepo, "refs/heads/"+r.commitBranch); err == nil {
		return ref, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(r.ctx, r.sourceOwner, r.sourceRepo, "refs/heads/"+r.baseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + r.commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(r.ctx, r.sourceOwner, r.sourceRepo, newRef)
	return ref, err
}

func (r *Release) getTree(ref *github.Reference) (tree *github.Tree, err error) {
	entries := []github.TreeEntry{}

	// Load each file into the tree.
	for _, c := range r.Changes {
		content, err := r.getChangedContent(c, ref)
		if err != nil {
			return nil, err
		}

		entries = append(entries, github.TreeEntry{Path: github.String(c.filePath), Type: github.String("blob"), Content: github.String(content), Mode: github.String("100644")})
	}

	tree, _, err = client.Git.CreateTree(r.ctx, r.sourceOwner, r.sourceRepo, *ref.Object.SHA, entries)
	return tree, err
}

func (r *Release) pushCommit(ref *github.Reference, tree *github.Tree) (err error) {
	parent, _, err := client.Repositories.GetCommit(r.ctx, r.sourceOwner, r.sourceRepo, *ref.Object.SHA)
	if err != nil {
		return err
	}

	parent.Commit.SHA = parent.SHA

	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &r.authorName, Email: &r.authorEmail}
	commit := &github.Commit{Author: author, Message: &r.commitMessage, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(r.ctx, r.sourceOwner, r.sourceRepo, commit)
	if err != nil {
		return err
	}

	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(r.ctx, r.sourceOwner, r.sourceRepo, ref, false)
	return err
}

func (r *Release) createPR() (*string, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(r.prTitle),
		Head:                github.String(r.commitBranch),
		Base:                github.String(r.baseBranch),
		Body:                github.String(""), // maybe get diff from the app repo
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(r.ctx, r.sourceOwner, r.sourceRepo, newPR)
	if err != nil {
		return nil, err
	}

	return github.String(pr.GetHTMLURL()), nil
}

func (r *Release) getChangedContent(c Change, ref *github.Reference) (string, error) {
	opt := &github.RepositoryContentGetOptions{
		// Ref: *ref.URL,
		Ref: "master",
	}

	f, _, _, err := client.Repositories.GetContents(r.ctx, r.sourceOwner, r.sourceRepo, c.filePath, opt)

	if err != nil {
		return "", err
	}

	original, err := f.GetContent()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(c.regexText)
	return re.ReplaceAllString(original, c.changedText), nil
}
