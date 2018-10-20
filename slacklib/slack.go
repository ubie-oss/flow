package slacklib

import (
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

	title := s.AppName

	if s.IsPrNotify {
		title += " Build"
	} else {
		title += " Deploy"
	}

	color := colorSuccess
	if !s.IsSuccess {
		color = colorDanger
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Color: color,
				Title: title,
				Fields: []slack.AttachmentField{
					slack.AttachmentField{
						Title: "Images",
						Value: "```\n" + strings.Join(s.Images, "\n") + "\n```",
						Short: false,
					},
					slack.AttachmentField{
						Title: "Deploy Pull Request",
						Value: "Merge this PullRequest for Production Relase\n" + s.PrURL,
						Short: false,
					},
					slack.AttachmentField{
						Title: "Logs",
						Value: s.LogURL,
						Short: false,
					},
				},
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
