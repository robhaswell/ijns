package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"github.com/nlopes/slack"
	"github.com/spf13/viper"
)

const SLACK_NAME = "agrakari"

var CHARACTERS = map[string]bool{"Maaya Saraki": true, "Indy Drone 4": true}

var slackApi *slack.Client

type EVEAPI struct {
	XMLName xml.Name `xml:"eveapi"`
	Result  Result   `xml:"result"`
}

type Result struct {
	XMLName xml.Name `xml:"result"`
	Rowset  Rowset   `xml:"rowset"`
}

type Rowset struct {
	XMLName xml.Name `xml:"rowset"`
	Job     []Job    `xml:"row"`
}

type Job struct {
	XMLName       xml.Name `xml:"row"`
	ID            int      `xml:"jobID,attr"`
	Blueprint     string   `xml:"blueprintTypeName,attr"`
	Installer     string   `xml:"installerName,attr"`
	EndDateString string   `xml:"endDate,attr"`
	EndDate       time.Time
}
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
	params := slack.PostMessageParameters{
		Username: "ijns",
	}
	_, _, err := slackApi.PostMessage("@"+SLACK_NAME, self.String(), params)
	if err != nil {
		log.Print(err)
	}
}
func (self *Job) String() string {
	return fmt.Sprintf("%s // %s will be delivered in 1 minute", self.Installer, self.Blueprint)
}
// If there is another job for the same blueprint & character due within the
// next minute, return false
func (self Job) IsSuperceded() bool {
	for job, _ := range(allJobs) {
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

func poll(ijr IndustryJobsRequester) error {
	vCode := viper.GetString("vcode")
	keyID := viper.GetString("keyid")

	body, err := ijr.GetXML(vCode, keyID)
	if err != nil {
		return err
	}

	var eveapi EVEAPI
	xml.Unmarshal(body, &eveapi)

	jobs := eveapi.Result.Rowset.Job

	log.Printf("Retrieved %d jobs", len(jobs))

	for _, job := range(jobs) {
		addJob(job)
	}
	// Prune the list of jobs
	for job, _ := range(allJobs) {
		if time.Since(job.EndDate) > time.Minute {
			log.Printf("Deleting job %d", job.ID)
			delete(allJobs, job)
		}
	}
	return nil
}

// Add a job if it is interesting and does not exist, and return whether it
// was added.
func addJob(job Job) bool {
	if _, ok := CHARACTERS[job.Installer]; ok {
		job.ParseDate()
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

	slackApi = slack.New(viper.GetString("slack_token"))

	ijh := new(XmlApiIndustryJobsRequester)

	for {
		if err := poll(ijh); err != nil {
			log.Print(err)
		}
		time.Sleep(15 * time.Minute)
	}
}
