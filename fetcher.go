/* The fetcher periodically fetches the list of industry jobs from CCP and
 * sends them to a channel. */

package main

import (
	"log"
	"time"

	"github.com/jonboulle/clockwork"
)

type Fetcher struct {
	ch        chan ([]Job)
	requester IndustryJobsRequester
	clock     clockwork.Clock
}

func NewFetcher(receiver chan ([]Job), requester IndustryJobsRequester, clock clockwork.Clock) *Fetcher {
	return &Fetcher{receiver, requester, clock}
}

// The main loop of the fetcher. Errors are logged.
func (self *Fetcher) Loop() {
	if err := self.Poll(); err != nil {
		log.Print(err)
	}
	self.clock.Sleep(15 * time.Minute)
}

// Request and parse the list of jobs and write them to the specified channel.
func (self *Fetcher) Poll() error {
	body, err := self.requester.GetXML()
	if err != nil {
		return err
	}

	jobs, err := ParseXmlApiResponse(body)
	if err != nil {
		return err
	}

	log.Printf("Retrieved %d jobs", len(jobs))

	self.ch <- jobs
	return nil
}
