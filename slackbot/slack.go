package slackbot

import (
	"fmt"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

type slackMessage struct {
	apiKey  string
	channel string
	MessageDetail
}

type MessageDetail struct {
	IsSuccess    bool
	IsPrNotify   bool
	AppName      string
	Images       []string
	LogURL       string
	PrURL        string
	BranchName   *string
	TagName      *string
	Time         time.Duration
	ErrorMessage string
}

func NewSlackMessage(apiKey, channel string, d MessageDetail) *slackMessage {
	return &slackMessage{
		apiKey:        apiKey,
		channel:       channel,
		MessageDetail: d,
	}
}

func (s *slackMessage) Post() error {
	api := slack.New(s.apiKey)

	var title, color string

	title = "Build"

	if s.IsSuccess {
		color = colorSuccess
		title += " Success"
	} else {
		color = colorDanger
		title += " Failure"
	}

	fields := []slack.AttachmentField{}

	fields = append(fields, slack.AttachmentField{
		Title: "App",
		Value: s.AppName,
		Short: true,
	})

	if s.BranchName != nil {
		fields = append(fields, slack.AttachmentField{
			Title: "Branch",
			Value: *s.BranchName,
			Short: true,
		})
	}

	if s.TagName != nil {
		fields = append(fields, slack.AttachmentField{
			Title: "Tag",
			Value: *s.TagName,
			Short: true,
		})
	}

	if len(s.Images) > 0 {
		fields = append(fields, slack.AttachmentField{
			Title: "Images",
			Value: "```\n" + strings.Join(s.Images, "\n") + "\n```",
			Short: false,
		})
	}

	if s.PrURL != "" {
		fields = append(fields, slack.AttachmentField{
			Title: "Deploy Pull Request",
			Value: fmt.Sprintf("Merge this PullRequest for Release\n%s\n", s.PrURL),
			Short: false,
		})
	}

	fields = append(fields, slack.AttachmentField{
		Title: "Logs",
		Value: fmt.Sprintf("<%s|BuildLog>", s.LogURL),
		Short: false,
	})

	if s.ErrorMessage != "" {
		fields = append(fields, slack.AttachmentField{
			Title: "Errors",
			Value: fmt.Sprintf("%s", s.ErrorMessage),
			Short: false,
		})
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Color:  color,
				Title:  title,
				Fields: fields,
			},
		},
		Markdown:  true,
		LinkNames: 1,
		AsUser:    true,
	}

	// ignore channelID and timestamp
	_, _, err := api.PostMessage(s.channel, "", params)
	return err
}
