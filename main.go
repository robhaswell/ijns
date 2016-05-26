package main

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

const SLACK_NAME = "agrakari"

var CHARACTERS = map[string]bool{"Maaya Saraki": true, "Indy Drone 4": true, "Fake Character": true}

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
