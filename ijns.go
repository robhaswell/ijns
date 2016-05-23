package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

const SLACK_NAME = "agrakari"

var CHARACTERS = map[string]bool{"Maaya Saraki": true, "Indy Drone 4": true}

func (self *Job) ParseDate() error {
	endDate, err := time.Parse(DateFormat, self.EndDateString)
	self.EndDate = endDate
	return err
}

// Alert 1 minute before the job is due to complete
func (self *Job) MakeAlert() {
	duration := self.EndDate.Sub(time.Now()) - time.Minute
	time.AfterFunc(duration, self.Alert)
	log.Printf("Will alert about %s in %s", self.Blueprint, duration)
}
func (self *Job) Alert() {
	if self.IsSuperceded() {
		return
	}
	self.Alerter.Alert(self, SLACK_NAME)
}
func (self *Job) String() string {
	return fmt.Sprintf("%s // %s will be delivered in 1 minute", self.Installer, self.Blueprint)
}

// If there is another job for the same blueprint & character due within the
// next minute, return false
func (self Job) IsSuperceded() bool {
	for job, _ := range allJobs {
		if job == self {
			continue
		}
		if job.Installer != self.Installer {
			continue
		}
		if job.Blueprint != self.Blueprint {
			continue
		}
		delta := job.EndDate.Sub(self.EndDate)
		// For jobs at identical times, the highest ID wins
		if delta == 0 {
			return self.ID < job.ID
		}
		if delta > 0 && delta <= time.Minute {
			return true
		}
	}
	return false
}

var allJobs = make(map[Job]bool)

const DateFormat = "2006-01-02 15:04:05"

func poll(requester IndustryJobsRequester, alerter Alerter) error {
	vCode := viper.GetString("vcode")
	keyID := viper.GetString("keyid")

	body, err := requester.GetXML(vCode, keyID)
	if err != nil {
		return err
	}

	jobs, err := ParseXmlApiResponse(body)
	if err != nil {
		return err
	}

	log.Printf("Retrieved %d jobs", len(jobs))

	activeJobIDs := make(map[int]bool)

	for _, job := range jobs {
		addJob(job, alerter)
		activeJobIDs[job.ID] = true
	}
	// Prune the list of jobs
	for job, _ := range allJobs {
		if !activeJobIDs[job.ID] {
			log.Printf("Deleting job %d", job.ID)
			delete(allJobs, job)
		}
	}
	return nil
}

// Add a job if it is interesting and does not exist, and return whether it
// was added.
func addJob(job Job, alerter Alerter) bool {
	if _, ok := CHARACTERS[job.Installer]; ok {
		// TODO stop poking into the job here
		job.ParseDate()
		// TODO job should not alert themselves
		job.Alerter = alerter
		if _, ok := allJobs[job]; !ok {
			job.MakeAlert()
			allJobs[job] = true
			return true
		}
	}
	return false
}

func main() {
	viper.SetEnvPrefix("ijns")
	viper.BindEnv("vcode")
	viper.BindEnv("keyid")
	viper.BindEnv("slack_token")

	requester := new(XmlApiIndustryJobsRequester)
	alerter := NewSlackAlerter(viper.GetString("slack_token"))

	for {
		if err := poll(requester, alerter); err != nil {
			log.Print(err)
		}
		time.Sleep(15 * time.Minute)
	}
}
