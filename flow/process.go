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

	if e.isSuuccess() { // Cloud Build Success

		if e.isApplicationBuild() { // Build for Application
			app, err := getApplicationByEventRepoName(e.RepoName)
			if err != nil {
				return f.notifyOnlyBuildResult(e)
			}

			var prs PullRequests

			for _, manifest := range app.Manifests {
				prURL, err := f.createRelasePR(ctx, e, *app, manifest)
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
				f.notifyFalure(e, err.Error())
				return err
			}
			return f.notifyRelasePR(e, prs)
		}

		// Build for Deployment
		// return f.notifyDeploy(e)
		return nil
	}

	// Code Build Failure
	return f.notifyFalure(e, "")
}

func (f *Flow) createRelasePR(ctx context.Context, e Event, a Application, m Manifest) (string, error) {
	repo := gitbot.NewRepo(cfg.ManifestOwner, cfg.ManifestName, cfg.ManifestBaseBranch)
	version, err := getVersionFromImage(e.Images)
	if err != nil {
		return "", err
	}

	release := gitbot.NewRelease(*repo, a.Name, m.Env, version)

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

func (f *Flow) notifyOnlyBuildResult(e Event) error {
	d := slackbot.MessageDetail{
		IsSuccess:  true,
		IsPrNotify: false,
		LogURL:     e.LogURL,
		AppName:    e.getAppName(),
		Images:     e.Images,
	}

	return slackbot.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func (f *Flow) notifyRelasePR(e Event, prs PullRequests) error {
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
