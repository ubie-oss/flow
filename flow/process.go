package flow

import (
	"fmt"

	"github.com/sakajunquality/flow/slacklib"
)

func (f *Flow) process(e Event) error {
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

	// @todo Create PulRequest here

	return "https://github.com/xxx/yyy/pulls/zzz", nil
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
