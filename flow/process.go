package flow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sakajunquality/flow/gitbot"
	"github.com/sakajunquality/flow/slackbot"
)

func (f *Flow) process(ctx context.Context, e Event) error {
	if !e.isFinished() { // Notify only the finished
		return nil
	}

	if e.isSuuccess() { // Cloud Build Success

		if e.isApplicationBuild() { // Build for Application
			prURL, err := f.createRelasePR(ctx, e)
			if err != nil {
				f.notifyFalure(e, err.Error())
				return err
			}
			return f.notifyRelasePR(e, prURL)
		}

		// Build for Deployment
		return f.notifyDeploy(e)
	}

	// Code Build Failure
	return f.notifyFalure(e, "")
}

func (f *Flow) createRelasePR(ctx context.Context, e Event) (string, error) {
	app, err := getApplicationByEventRepoName(e.RepoName)
	if err != nil {
		return "", err
	}

	repo := gitbot.NewRepo(app.SourceOwner, app.SourceName, app.BaseBranch)
	version, err := getVersionFromImage(e.Images)
	if err != nil {
		return "", err
	}

	release := gitbot.NewRelease(*repo, app.Env, version)

	for _, filePath := range app.Manifests {
		release.AddChanges(filePath, fmt.Sprintf("%s:.*", app.ImageName), fmt.Sprintf("%s:%s", app.ImageName, version))
	}

	// Add Commit Author
	release.AddAuthor(cfg.GitAuthor.Name, cfg.GitAuthor.Email)

	// Create a release PullRequest
	prURL, err := release.Create(ctx, f.githubToken)
	if err != nil {
		return "", err
	}
	return *prURL, nil
}

func (f *Flow) notifyRelasePR(e Event, prURL string) error {
	d := slackbot.MessageDetail{
		IsSuccess:  true,
		IsPrNotify: true,
		LogURL:     e.LogURL,
		AppName:    e.getAppName(),
		Images:     e.Images,
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
	}

	return slackbot.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func (f *Flow) notifyFalure(e Event, errorMessage string) error {
	d := slackbot.MessageDetail{
		IsSuccess:    false,
		LogURL:       e.LogURL,
		AppName:      e.RepoName,
		Images:       e.Images,
		ErrorMessage: errorMessage,
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

func getVersionFromImage(images []string) (string, error) {
	if len(images) < 1 {
		return "", errors.New("no images found")
	}
	// does not support multiple images
	tags := strings.Split(images[0], ":")
	return tags[1], nil
}
