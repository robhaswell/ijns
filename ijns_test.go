package main

import (
	"testing"

	"github.com/nlopes/slack"
	"github.com/spf13/viper"
)

func TestSuperceded(t *testing.T) {
	allJobs = make(map[Job]bool)
	j1 := Job{
		ID: 1,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	j1.ParseDate()
	j2 := Job{
		ID: 2,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:02",
	}
	j2.ParseDate()
	j3 := Job{
		ID: 3,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:02:03",
	}
	j3.ParseDate()
	j4 := Job{
		ID: 4,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 02:02:03",
	}
	j4.ParseDate()
	j5 := Job{
		ID: 5,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 02:02:03",
	}
	j5.ParseDate()
	allJobs[j1] = true
	allJobs[j2] = true
	allJobs[j3] = true
	allJobs[j4] = true
	allJobs[j5] = true

	expected := true
	result := j1.IsSuperceded()
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = false
	result = j2.IsSuperceded()
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = false
	result = j3.IsSuperceded()
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = true
	result = j4.IsSuperceded()
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = false
	result = j5.IsSuperceded()
	if expected != result {
		t.Fatal("Unexpected result", result)
	}
}

func TestJobEquality(t *testing.T) {
	j1 := Job{
		ID: 1,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	j2 := Job{
		ID: 1,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	if j1 != j2 {
		t.Fatal("Jobs are not equal")
	}
}

func TestAddJob(t *testing.T) {
	allJobs = make(map[Job]bool)
	j1 := Job{
		ID: 1,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	j2 := Job{
		ID: 1,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}

	if j1 != j2 {
		t.Fatal("Jobs are not equal")
	}

	var added bool

	added = addJob(j1)
	if !added {
		t.Fatal("Should have added first job")
	}

	added = addJob(j2)
	if added {
		t.Fatal("Should not have added second job")
	}
}

// This will actualy post a message to "agrakari" on Slack.
func TestSlackAlert(t *testing.T) {
	viper.SetEnvPrefix("ijns")
	viper.BindEnv("slack_token")
	token := viper.GetString("slack_token")

	if token == "" {
		t.Skip("Token not specified")
	}

	slackApi = slack.New(token)
	j1 := Job{
		ID: 1,
		Blueprint: "Test Item Blueprint I",
		Installer: "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	j1.Alert()
}
