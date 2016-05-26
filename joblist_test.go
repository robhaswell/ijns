package main

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSuperceded(t *testing.T) {
	xml := []byte(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item I Blueprint" endDate="2020-01-01 01:01:01" />
      <row jobID="2" installerName="Fake Character" blueprintTypeName="Test Item I Blueprint" endDate="2020-01-01 01:01:02" />
      <row jobID="3" installerName="Fake Character" blueprintTypeName="Test Item I Blueprint" endDate="2020-01-01 01:02:03" />
      <row jobID="4" installerName="Fake Character" blueprintTypeName="Test Item I Blueprint" endDate="2020-01-01 02:02:03" />
      <row jobID="5" installerName="Fake Character" blueprintTypeName="Test Item I Blueprint" endDate="2020-01-01 02:02:03" />
    </rowset>
  </result>
</eveapi>`)
	jobs, err := ParseXmlApiResponse(xml)
	if err != nil {
		t.Fatal(err)
	}

	jobList := NewJobList(NewTestCharacterConfig(), NewFakeAlerter())
	jobList.SetJobs(jobs)

	expected := true
	result := jobList.IsSuperceded(jobs[0])
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = false
	result = jobList.IsSuperceded(jobs[1])
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = false
	result = jobList.IsSuperceded(jobs[2])
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = true
	result = jobList.IsSuperceded(jobs[3])
	if expected != result {
		t.Fatal("Unexpected result", result)
	}

	expected = false
	result = jobList.IsSuperceded(jobs[4])
	if expected != result {
		t.Fatal("Unexpected result", result)
	}
}

// Jobs which are removed from the feed are forgotten about.
func TestJobsAreRemoved(t *testing.T) {
	xml := []byte(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item I Blueprint" endDate="2020-01-01 01:01:01" />
      <row jobID="2" installerName="Fake Character" blueprintTypeName="Test Item I Blueprint" endDate="2020-01-01 01:01:02" />
    </rowset>
  </result>
</eveapi>`)
	jobs, err := ParseXmlApiResponse(xml)
	if err != nil {
		t.Fatal(err)
	}

	jobList := NewJobList(NewTestCharacterConfig(), NewFakeAlerter())
	jobList.SetJobs(jobs)

	jobs = jobs[1:]
	jobList.SetJobs(jobs)

	if jobList.Count() != 1 {
		t.Fatal("Unexpected elements in job list.")
	}
}

func TestJobEquality(t *testing.T) {
	j1 := Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Fake Character",
		EndDateString: "2020-01-01 01:01:01",
	}
	j2 := Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Fake Character",
		EndDateString: "2020-01-01 01:01:01",
	}
	if j1 != j2 {
		t.Fatal("Jobs are not equal")
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

	alerter := NewSlackAlerter(token)
	j1 := &Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Fake Character",
		EndDateString: "2020-01-01 01:01:01",
	}
	alerter.Alert(j1, "agrakari")
}
