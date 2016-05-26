package main

import (
	"fmt"
	"testing"
	"time"
)

// XML containing a job 1m1s in the future results in a job being alerted in 1s.
func TestSimpleE2E(t *testing.T) {
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, time.Now().UTC().Add(time.Minute+time.Second).Format(DateFormat)))
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
func TestNearFutureE2E(t *testing.T) {
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, time.Now().UTC().Add(30*time.Second).Format(DateFormat)))
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
func TestFarPastE2E(t *testing.T) {
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, time.Now().UTC().Add(-5*time.Hour).Format(DateFormat)))
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
func TestDuplicateJobE2E(t *testing.T) {
	xml := []byte(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="2020-01-01 01:01:01" />
    </rowset>
  </result>
</eveapi>`)
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
