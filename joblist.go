package main

import (
	"log"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/jonboulle/clockwork"
)

// Records jobs and creates alerts about them
type JobList struct {
	config  CharacterConfig
	clock   clockwork.Clock
	alerter Alerter
	jobs    mapset.Set
	Ch      chan ([]Job)
}

func NewJobList(config CharacterConfig, clock clockwork.Clock, alerter Alerter) *JobList {
	return &JobList{
		config:  config,
		clock:   clock,
		alerter: alerter,
		jobs:    mapset.NewSet(),
		Ch:      make(chan ([]Job)),
	}
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
	for _, job := range jobs {
		if self.isInteresting(&job) {
			currentJobs.Add(job)
		}
	}
	newJobs := currentJobs.Difference(self.jobs)
	// TODO add a lock around this
	self.jobs = currentJobs

	for newJobInterface := range newJobs.Iter() {
		newJob := newJobInterface.(Job)
		self.startAlertTimer(&newJob)
	}
}

// A job is interesting if it belongs to a configured character.
func (self *JobList) isInteresting(job *Job) bool {
	return self.config.CharacterSet().Contains(job.Installer)
}

// Alert about the job 1 minute before its end date
func (self *JobList) startAlertTimer(job *Job) {
	duration := job.EndDate.Sub(self.clock.Now()) - time.Minute
	// Do not bother to create an alert if it is due in the past
	if duration < 0 {
		return
	}
	c := self.clock.After(duration)
	go func() {
		_ = <-c
		self.Alert(job)
	}()
	log.Printf("Will alert about %s in %s", job.Blueprint, duration)
}

func (self *JobList) Alert(job *Job) {
	// Do not bother if it is superceded
	if self.IsSuperceded(*job) {
		return
	}
	self.alerter.Alert(job, self.config.AlertUsername(job.Installer))
}

// If there is another job for the same blueprint & character due within the
// next minute, return false
func (self *JobList) IsSuperceded(job Job) bool {
	for otherJobInterface := range self.jobs.Iter() {
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
