package main

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/deckarep/golang-set"
)

const SLACK_NAME = "agrakari"

var CHARACTERS = map[string]bool{"Maaya Saraki": true, "Indy Drone 4": true, "Fake Character": true}

// Records jobs and creates alerts about them
type JobList struct {
	alerter Alerter
	jobs mapset.Set
}

func NewJobList(alerter Alerter) *JobList {
	jobList := &JobList{}
	jobList.Init(alerter)
	return jobList
}

func (self *JobList) Init(alerter Alerter) {
	self.alerter = alerter
	self.jobs = mapset.NewSet()
}

func (self *JobList) Count() int {
	return self.jobs.Cardinality()
}

func (self *JobList) String() string {
	return self.jobs.String()
}

// Given the current complete list of Jobs read from the jobs API, record
// interesting jobs and initialise alerts.
func (self *JobList) SetJobs(jobs []Job) {
	currentJobs := mapset.NewSet()
	for _, job := range(jobs) {
		if self.isInteresting(&job) {
			currentJobs.Add(job)
		}
	}
	newJobs := currentJobs.Difference(self.jobs)
	// TODO add a lock around this
	self.jobs = currentJobs

	for newJobInterface := range(newJobs.Iter()) {
		newJob := newJobInterface.(Job)
		self.startAlertTimer(&newJob)
	}
}

// A job is interesting if it belongs to a configured character.
func (self *JobList) isInteresting(job *Job) bool {
	_, ok := CHARACTERS[job.Installer]
	return ok
}

// Alert about the job 1 minute before its end date
func (self *JobList) startAlertTimer(job *Job) {
	duration := job.EndDate.Sub(time.Now()) - time.Minute
	// Do not bother to create an alert if it is due in the past
	if duration < 0 {
		return
	}
	time.AfterFunc(duration, func() {
		self.Alert(job)
	})
	log.Printf("Will alert about %s in %s", job.Blueprint, duration)
}

func (self *JobList) Alert(job *Job) {
	// Do not bother if it is superceded
	if self.IsSuperceded(*job) {
		return
	}
	self.alerter.Alert(job, SLACK_NAME)
}

// If there is another job for the same blueprint & character due within the
// next minute, return false
func (self *JobList) IsSuperceded(job Job) bool {
	for otherJobInterface := range(self.jobs.Iter()) {
		otherJob := otherJobInterface.(Job)
		if job == otherJob {
			continue
		}
		if job.Installer != otherJob.Installer {
			continue
		}
		if job.Blueprint != otherJob.Blueprint {
			continue
		}
		delta := otherJob.EndDate.Sub(job.EndDate)
		// For jobs at identical times, the highest ID wins
		if delta == 0 {
			return job.ID < otherJob.ID
		}
		if delta > 0 && delta <= time.Minute {
			return true
		}
	}
	return false
}

func mainLoop(jobList *JobList, requester IndustryJobsRequester) error {
	body, err := requester.GetXML()
	if err != nil {
		return err
	}

	jobs, err := ParseXmlApiResponse(body)
	if err != nil {
		return err
	}

	log.Printf("Retrieved %d jobs", len(jobs))

	jobList.SetJobs(jobs)
	return nil
}

func main() {
	viper.SetEnvPrefix("ijns")
	viper.BindEnv("vcode")
	viper.BindEnv("keyid")
	viper.BindEnv("slack_token")

	requester := NewXmlApiIndustryJobsRequester(viper.GetString("vcode"), viper.GetString("keyid"))
	alerter := NewSlackAlerter(viper.GetString("slack_token"))

	jobList := NewJobList(alerter)

	for {
		if err := mainLoop(jobList, requester); err != nil {
			log.Print(err)
		}
		time.Sleep(15 * time.Minute)
	}
}
