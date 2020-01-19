package flow

import (
	"context"
	"errors"
	"fmt"
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
	if version == "" {
		return false
	}
	// ignore latest tag
	if version == "latest" {
		return false
	}

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
		release.AddChanges(filePath, fmt.Sprintf("%s:.*", a.Image), fmt.Sprintf("%s:%s", a.Image, version))
		if a.RewriteVersion {
			release.AddChanges(filePath, "version: .*", fmt.Sprintf("version: %s", version))
		}

		if a.RewriteNewTag && strings.Contains(filePath, "kustomization.yaml") {
			release.AddChanges(filePath, "newTag: .*", fmt.Sprintf("newTag: %s", version))
		}
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

func getApplicationByEventRepoName(eventRepoName string) (*Application, error) {
	for _, app := range cfg.ApplicationList {
		// CloudBuild Repo Names
		if eventRepoName == fmt.Sprintf("github-%s-%s", app.SourceOwner, app.SourceName) {
			return &app, nil
		}
	}
	return nil, errors.New("No application found for " + eventRepoName)
}

func getApplicationByImage(image string) (*Application, error) {
	for _, app := range cfg.ApplicationList {
		if image == app.Image {
			return &app, nil
		}
	}
	return nil, errors.New("No application found for image " + image)
}
