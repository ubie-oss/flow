package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sakajunquality/flow/gitbot"
	"github.com/sakajunquality/flow/slackbot"
)

type PullRequests []PullRequest

type PullRequest struct {
	env string
	url string
	err error
}

func (f *Flow) process(ctx context.Context, e Event) error {
	if !e.isFinished() { // Notify only the finished
		fmt.Fprintf(os.Stdout, "Build hasn't finished\n")
		return nil
	}

	app, err := getApplicationByEventTriggerID(e.TriggerID)
	if err != nil {
		fmt.Fprintf(os.Stdout, "No app is configured for %s\n", e.TriggerID)
		return nil
	}

	if !e.isSuuccess() { // CloudBuild Failure
		return f.notifyFalure(e, "", nil)
	}

	var prs PullRequests

	version, err := getVersionFromImage(e.Images)
	if err != nil {
		return f.notifyFalure(e, fmt.Sprintf("Could not ditermine version from image: %s", err), nil)
	}

	for _, manifest := range app.Manifests {
		if !shouldCreatePR(manifest, version) {
			continue
		}

		prURL, err := f.createRelasePR(ctx, version, *app, manifest)

		if err != nil {
			prs = append(prs, PullRequest{
				env: manifest.Env,
				err: err,
			})
			continue
		}

		prs = append(prs, PullRequest{
			env: manifest.Env,
			url: prURL,
		})
	}

	if err != nil {
		f.notifyFalure(e, err.Error(), app)
		return err
	}
	return f.notifyRelasePR(e, prs, app)
}

func shouldCreatePR(m Manifest, version string) bool {
	for _, prefix := range m.Filters.ExcludePrefixes {
		if strings.HasPrefix(version, prefix) {
			return false
		}
	}

	if len(m.Filters.IncludePrefixes) == 0 {
		return true
	}

	for _, prefix := range m.Filters.IncludePrefixes {
		if strings.HasPrefix(version, prefix) {
			return true
		}
	}

	return false
}

// createRelasePR submits release PullRequest to manifest repository
func (f *Flow) createRelasePR(ctx context.Context, version string, a Application, m Manifest) (string, error) {
	repo := gitbot.NewRepo(a.ManifestOwner, a.ManifestName, a.ManifestBaseBranch)
	release := gitbot.NewRelease(*repo, a.Name, m.Env, version, m.PRBody)

	for _, filePath := range m.Files {
		release.AddChanges(filePath, fmt.Sprintf("%s:.*", a.ImageName), fmt.Sprintf("%s:%s", a.ImageName, version))
	}

	// Add Commit Author
	release.AddAuthor(cfg.GitAuthor.Name, cfg.GitAuthor.Email)

	fmt.Printf("%#v", release)

	// Create a release PullRequest
	prURL, err := release.Create(ctx, f.githubToken)
	if err != nil {
		return "", err
	}
	return *prURL, nil
}

func (f *Flow) notifyRelasePR(e Event, prs PullRequests, app *Application) error {
	var prURL string

	for _, pr := range prs {
		if pr.err != nil {
			prURL += fmt.Sprintf("`%s`\n```%s```\n", pr.env, pr.err)
			continue
		}

		prURL += fmt.Sprintf("`%s`\n```%s```\n", pr.env, pr.url)
	}

	d := slackbot.MessageDetail{
		IsSuccess:  true,
		IsPrNotify: true,
		LogURL:     e.LogURL,
		AppName:    app.Name,
		Images:     e.Images,
		TagName:    e.TagName,
		BranchName: e.BranchName,
		PrURL:      prURL,
	}

	return slackbot.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func (f *Flow) notifyDeploy(e Event) error {
	d := slackbot.MessageDetail{
		IsSuccess:  true,
		IsPrNotify: false,
		LogURL:     e.LogURL,
		AppName:    e.RepoName,
		TagName:    e.TagName,
		BranchName: e.BranchName,
	}

	return slackbot.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func (f *Flow) notifyFalure(e Event, errorMessage string, app *Application) error {
	d := slackbot.MessageDetail{
		IsSuccess:    false,
		LogURL:       e.LogURL,
		Images:       e.Images,
		ErrorMessage: errorMessage,
		TagName:      e.TagName,
		BranchName:   e.BranchName,
	}

	if app != nil {
		d.AppName = app.Name
	}

	return slackbot.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func getApplicationByEventRepoName(eventRepoName string) (*Application, error) {
	for _, app := range cfg.ApplicationList {
		// CloudBuild Repo Names
		if eventRepoName == fmt.Sprintf("github-%s-%s", app.SourceOwner, app.SourceName) {
			return &app, nil
		}
	}
	return nil, errors.New("No application found for " + eventRepoName)
}

func getApplicationByEventTriggerID(eventTriggerID string) (*Application, error) {
	for _, app := range cfg.ApplicationList {
		// CloudBuild Repo Names
		if eventTriggerID == app.TriggerID {
			return &app, nil
		}
	}
	return nil, errors.New("No application found for " + eventTriggerID)
}

// Retrieve Docker Image tag from the built image
func getVersionFromImage(images []string) (string, error) {
	if len(images) < 1 {
		return "", errors.New("no images found")
	}
	// does not support multiple images
	tags := strings.Split(images[0], ":")
	return tags[1], nil
}
