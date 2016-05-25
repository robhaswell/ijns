package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestSuperceded(t *testing.T) {
	allJobs = make(map[Job]bool)
	j1 := Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	j1.ParseDate()
	j2 := Job{
		ID:            2,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:02",
	}
	j2.ParseDate()
	j3 := Job{
		ID:            3,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:02:03",
	}
	j3.ParseDate()
	j4 := Job{
		ID:            4,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 02:02:03",
	}
	j4.ParseDate()
	j5 := Job{
		ID:            5,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
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
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	j2 := Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	if j1 != j2 {
		t.Fatal("Jobs are not equal")
	}
}

func TestAddJob(t *testing.T) {
	alerter := &FakeAlerter{}
	allJobs = make(map[Job]bool)
	j1 := Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	j2 := Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}

	if j1 != j2 {
		t.Fatal("Jobs are not equal")
	}

	var added bool

	added = addJob(j1, alerter)
	if !added {
		t.Fatal("Should have added first job")
	}

	added = addJob(j2, alerter)
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

	alerter := NewSlackAlerter(token)
	j1 := &Job{
		ID:            1,
		Blueprint:     "Test Item Blueprint I",
		Installer:     "Maaya Saraki",
		EndDateString: "2020-01-01 01:01:01",
	}
	alerter.Alert(j1, "agrakari")
}

// XML containing a job 1m1s in the future results in a job being alerted in 1s.
func TestSimpleE2E (t *testing.T) {
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, time.Now().UTC().Add(time.Minute + time.Second).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()

	if err := mainLoop(requester, alerter); err != nil {
		t.Fatal(err)
	}

	event := <-alerter.Chan

	if event.Job.ID != 1 {
		t.Fatal("Unexpected alert event", event)
	}
}

// XML containing a job 30 seconds in the future does not result in an alert
func TestNearFutureE2E (t *testing.T) {
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, time.Now().UTC().Add(30 * time.Second).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()

	if err := mainLoop(requester, alerter); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-alerter.Chan:
		t.Fatal("Unexpected alert event", event)
	default:
	}
}

// XML containing a job in the far pastt does not result in an alert
func TestFarPastE2E (t *testing.T) {
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, time.Now().UTC().Add(-5 * time.Hour).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()

	if err := mainLoop(requester, alerter); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-alerter.Chan:
		t.Fatal("Unexpected alert event", event)
	default:
	}
}
