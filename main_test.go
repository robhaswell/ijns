package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
)

// XML containing a job 1m1s in the future results in a job being alerted in 1s.
func TestSimpleE2E(t *testing.T) {
	clock := clockwork.NewFakeClock()
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, clock.Now().UTC().Add(time.Minute+time.Second).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()
	jobList := NewJobList(NewTestCharacterConfig(), clock, alerter)

	fetcher := NewFetcher(clock, jobList.Ch, requester)

	// Read the XML
	if err := fetcher.Poll(); err != nil {
		t.Fatal(err)
	}
	jobList.Tick()

	clock.Advance(time.Second)

	jobList.Tick()

	select {
	case event := <-alerter.Chan:
		if event.Job.ID != 1 {
			t.Fatal("Unexpected alert event", event)
		}
	default:
		t.Fatal("No alert event.")
	}
}

// XML containing a job 30 seconds in the future does not result in an alert
func TestNearFutureE2E(t *testing.T) {
	clock := clockwork.NewFakeClock()
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, clock.Now().UTC().Add(30*time.Second).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()
	jobList := NewJobList(NewTestCharacterConfig(), clock, alerter)

	fetcher := NewFetcher(clock, jobList.Ch, requester)

	// Read the XML
	if err := fetcher.Poll(); err != nil {
		t.Fatal(err)
	}
	jobList.Tick()

	clock.Advance(time.Minute)

	jobList.Tick()

	select {
	case event := <-alerter.Chan:
		t.Fatal("Unexpected alert event", event)
	default:
	}
}

// XML containing a job 2m30s in the future does not result in an alert after 1m
func TestFutureE2E(t *testing.T) {
	clock := clockwork.NewFakeClock()
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, clock.Now().UTC().Add(2*time.Minute + 30*time.Second).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()
	jobList := NewJobList(NewTestCharacterConfig(), clock, alerter)

	fetcher := NewFetcher(clock, jobList.Ch, requester)

	// Read the XML
	if err := fetcher.Poll(); err != nil {
		t.Fatal(err)
	}
	jobList.Tick()

	clock.Advance(time.Minute)

	jobList.Tick()

	select {
	case event := <-alerter.Chan:
		t.Fatal("Unexpected alert event", event)
	default:
	}
}

// XML containing a job in the far past does not result in an alert
func TestFarPastE2E(t *testing.T) {
	clock := clockwork.NewFakeClock()
	xml := []byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerName,blueprintTypeName,endDate">
      <row jobID="1" installerName="Fake Character" blueprintTypeName="Test Item Blueprint I" endDate="%v" />
    </rowset>
  </result>
</eveapi>`, clock.Now().UTC().Add(-5*time.Hour).Format(DateFormat)))
	requester := &FakeIndustryJobsRequester{}
	requester.SetResponse(xml)

	alerter := NewFakeAlerter()
	jobList := NewJobList(NewTestCharacterConfig(), clock, alerter)

	fetcher := NewFetcher(clock, jobList.Ch, requester)

	// Read the XML
	fetcher.Poll()
	jobList.Tick()

	clock.Advance(time.Minute)

	jobList.Tick()

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
	jobList := NewJobList(NewTestCharacterConfig(), clockwork.NewFakeClock(), alerter)

	fetcher := NewFetcher(clockwork.NewFakeClock(), jobList.Ch, requester)

	// Read the XML
	if err := fetcher.Poll(); err != nil {
		t.Fatal(err)
	}
	jobList.Tick()

	// Now consume it again
	if err := fetcher.Poll(); err != nil {
		t.Fatal(err)
	}
	jobList.Tick()

	if jobList.Count() != 1 {
		t.Fatal("Unexpected jobs: ", jobList.String())
	}
}
