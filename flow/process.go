package flow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sakajunquality/flow/gitbot"
	"github.com/sakajunquality/flow/slacklib"
)

func (f *Flow) process(ctx context.Context, e Event) error {
	if !e.isFinished() { // Notify only the finished
		return nil
	}

	if e.isSuuccess() { // Cloud Build Success

		if e.isApplicationBuild() { // Build for Application
			prURL, err := f.createRelasePR(e)
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

	fmt.Println("notify failure\n%#v\n", e)
	return nil
}

func (f *Flow) createRelasePR(e Event) (string, error) {
	ctx := context.Background()

	app := getApplicationByRepoName(e.RepoName)
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
	d := slacklib.MessageDetail{
		IsSuccess:  true,
		IsPrNotify: true,
		LogURL:     e.LogURL,
		AppName:    e.getAppName(),
		Images:     e.Images,
		PrURL:      prURL,
	}

	return slacklib.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()

}

func (f *Flow) notifyDeploy(e Event) error {
	d := slacklib.MessageDetail{
		IsSuccess:  true,
		IsPrNotify: false,
		LogURL:     e.LogURL,
		AppName:    e.RepoName,
	}

	return slacklib.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func (f *Flow) notifyFalure(e Event, errorMessage string) error {
	d := slacklib.MessageDetail{
		IsSuccess:    false,
		LogURL:       e.LogURL,
		AppName:      e.RepoName,
		Images:       e.Images,
		ErrorMessage: errorMessage,
	}

	return slacklib.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func getApplicationByRepoName(repoName string) *Application {
	return nil
}

func getVersionFromImage(images []string) (string, error) {
	if len(images) < 1 {
		return "", errors.New("no images found")
	}
	// does not support multiple images
	tags := strings.Split(images[0], ":")
	return tags[1], nil
}
