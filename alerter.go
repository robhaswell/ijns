package main

import (
	"github.com/nlopes/slack"
	"log"
)

type Alerter interface {
	Alert(*Job, string)
}

type SlackAlerter struct {
	api *slack.Client
}

func NewSlackAlerter(token string) *SlackAlerter {
	return &SlackAlerter{
		api: slack.New(token),
	}
}

func (s *SlackAlerter) Alert(job *Job, username string) {
	params := slack.PostMessageParameters{
		Username: "ijns",
	}
	_, _, err := s.api.PostMessage("@"+username, job.String(), params)
	if err != nil {
		log.Print(err)
	}
}

type FakeAlerter struct {}

func (s *FakeAlerter) Alert(job *Job, username string) {
	// no-op
}
