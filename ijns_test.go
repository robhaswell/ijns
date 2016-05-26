package main

import (
	"fmt"
	"testing"
	"time"

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

	jobList := NewJobList(NewFakeAlerter())
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
	jobList := NewJobList(alerter)

	if err := mainLoop(jobList, requester); err != nil {
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
	jobList := NewJobList(alerter)

	if err := mainLoop(jobList, requester); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-alerter.Chan:
		t.Fatal("Unexpected alert event", event)
	default:
	}
}

// XML containing a job in the far past does not result in an alert
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
	jobList := NewJobList(alerter)

	if err := mainLoop(jobList, requester); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-alerter.Chan:
		t.Fatal("Unexpected alert event", event)
	default:
	}
}

// Duplicate jobs are only added once.
func TestDuplicateJob(t *testing.T) {
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, time.Now().UTC().Add(time.Hour).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()
	jobList := NewJobList(alerter)

	// Consume the XML once
	if err := mainLoop(jobList, requester); err != nil {
		t.Fatal(err)
	}
	// Now consume it again
	if err := mainLoop(jobList, requester); err != nil {
		t.Fatal(err)
	}
	if jobList.Count() != 1 {
		t.Fatal("Unexpected jobs: ", jobList.String())
	}
}
