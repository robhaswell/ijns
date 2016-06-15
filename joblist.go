package main

import (
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
		Ch:      make(chan ([]Job), 1),
	}
}

func (self *JobList) Count() int {
	return self.jobs.Cardinality()
}

func (self *JobList) String() string {
	return self.jobs.String()
}

// Once per second, call the Tick() function.
func (self *JobList) Loop() {
	for {
		self.Tick()
		self.clock.Sleep(time.Second)
	}
}

// Look for new lists of jobs and process any alerts.
func (self *JobList) Tick() {
	select {
	case jobs := <-self.Ch:
		self.SetJobs(jobs)
	default:
	}

	// Search for jobs which are ripe and have not yet been alerted.
	for jobInterface := range self.jobs.Iter() {
		job := jobInterface.(Job)
		if !self.IsRipe(&job) {
			continue
		}
		if job.Alerted {
			continue
		}
		self.Alert(&job)
	}
}

// Given the current complete list of Jobs read from the jobs API, record
// interesting jobs and initialise alerts.
func (self *JobList) SetJobs(jobs []Job) {
	// TODO Prune the list of jobs
	validJobs := mapset.NewSet()
	for _, job := range jobs {
		validJobs.Add(job)
		if self.IsInteresting(&job) {
			// If the Job is ripe then we mark it as already alerted, in case
			// we are just restarting.
			if self.IsRipe(&job) {
				job.Alerted = true
			}
			self.jobs.Add(job)
		}
	}
	self.jobs = self.jobs.Intersect(validJobs)
}

// A job is interesting if it belongs to a configured character.
func (self *JobList) IsInteresting(job *Job) bool {
	return self.config.CharacterSet().Contains(job.Installer)
}

// A job is ripe if it is ready to be alerted about.
func (self *JobList) IsRipe(job *Job) bool {
	return job.EndDate.Sub(self.clock.Now()) <= time.Minute
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

func (self *JobList) Alert(job *Job) {
	self.alerter.Alert(job, self.config.AlertUsername(job.Installer))
	self.SetAlerted(job)
}

// Set the job as being alered by removing and adding the job
func (self *JobList) SetAlerted(job *Job) {
	job.Alerted = true
	self.jobs.Remove(job)
	self.jobs.Add(job)
}
