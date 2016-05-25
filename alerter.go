package main

import (
	"log"

	"github.com/nlopes/slack"
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

type FakeAlerter struct {
	Chan chan FakeAlertEvent
}

func NewFakeAlerter() *FakeAlerter {
	fake := FakeAlerter{}
	fake.Chan = make(chan FakeAlertEvent)
	return &fake
}

func (s *FakeAlerter) Alert(job *Job, username string) {
	s.Chan <- FakeAlertEvent{job, username}
}

type FakeAlertEvent struct {
	Job *Job
	Username string
}
