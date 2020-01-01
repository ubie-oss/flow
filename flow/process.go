package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sakajunquality/cloud-pubsub-events/cloudbuildevent"
	"github.com/sakajunquality/flow/gitbot"
	"github.com/sakajunquality/flow/slackbot"
)

type PullRequests []PullRequest

type PullRequest struct {
	env string
	url string
	err error
}

func (f *Flow) process(ctx context.Context, e cloudbuildevent.Event) error {
	if !e.IsFinished() { // Notify only the finished
		fmt.Fprintf(os.Stdout, "Build hasn't finished\n")
		return nil
	}

	if e.TriggerID == nil {
		return errors.New("Only the triggered build is supported")
	}

	app, err := getApplicationByEventTriggerID(*e.TriggerID)
	if err != nil {
		return fmt.Errorf("No app is configured for %s", *e.TriggerID)
	}

	if !e.IsSuuccess() { // CloudBuild Failure
		return f.notifyFalure(e, "", nil)
	}

	version, err := getVersionFromImage(e.Images)
	if err != nil {
		return f.notifyFalure(e, fmt.Sprintf("Could not ditermine version from image: %s", err), nil)
	}

	prs := f.generatePRs(ctx, app, version)
	return f.notifyReleasePR(e.Images[0], version, prs, app)
}

func (f *Flow) processImage(ctx context.Context, image, version string) error {
	app, err := getApplicationByImage(image)
	if err != nil {
		return err
	}

	prs := f.generatePRs(ctx, app, version)
	return f.notifyReleasePR(image, version, prs, app)
}

func (f *Flow) generatePRs(ctx context.Context, app *Application, version string) PullRequests {
	var prs PullRequests

	for _, manifest := range app.Manifests {
		if !shouldCreatePR(manifest, version) {
			continue
		}

		prURL, err := f.createReleasePR(ctx, version, *app, manifest)

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

	return prs
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

// createReleasePR submits release PullRequest to manifest repository
func (f *Flow) createReleasePR(ctx context.Context, version string, a Application, m Manifest) (string, error) {
	baseBranch := a.ManifestBaseBranch
	if m.BaseBranch != "" {
		baseBranch = m.BaseBranch
	}

	repo := gitbot.NewRepo(a.ManifestOwner, a.ManifestName, baseBranch)

	// Create PR Body with tag page URL
	prBody := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", a.SourceOwner, a.SourceName, version)
	if m.PRBody != "" {
		prBody += fmt.Sprintf("\n\n%s", m.PRBody)
	}
	release := gitbot.NewRelease(*repo, a.Name, m.Env, version, prBody)

	for _, filePath := range m.Files {
		release.AddChanges(filePath, fmt.Sprintf("%s:.*", a.ImageName), fmt.Sprintf("%s:%s", a.ImageName, version))
		if a.RewriteVersion {
			release.AddChanges(filePath, "version: .*", fmt.Sprintf("version: %s", version))
		}

		if a.RewriteNewTag && strings.Contains(filePath, "kustomization.yaml") {
			release.AddChanges(filePath, "newTag: .*", fmt.Sprintf("newTag: %s", version))
		}
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

func (f *Flow) notifyReleasePR(image, version string, prs PullRequests, app *Application) error {
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
		AppName:    app.Name,
		Image:      image,
		Version:    version,
		PrURL:      prURL,
	}

	return slackbot.NewSlackMessage(f.slackBotToken, cfg.SlackNotifiyChannel, d).Post()
}

func (f *Flow) notifyFalure(e cloudbuildevent.Event, errorMessage string, app *Application) error {
	d := slackbot.MessageDetail{
		IsSuccess:    false,
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
		if eventTriggerID == app.TriggerID {
			return &app, nil
		}
	}
	return nil, errors.New("No application found for " + eventTriggerID)
}

func getApplicationByImage(image string) (*Application, error) {
	for _, app := range cfg.ApplicationList {
		if image == app.ImageName {
			return &app, nil
		}
	}
	return nil, errors.New("No application found for image " + image)
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
