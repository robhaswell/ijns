package main

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

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
	viper.BindEnv("characters")

	config, err := NewJsonCharacterConfig(viper.GetString("characters"))
	if err != nil {
		log.Fatalf("Invalid JSON in env IJNS_CHARACTERS=%s: %v", viper.GetString("characters"), err)
	}
	requester := NewXmlApiIndustryJobsRequester(viper.GetString("vcode"), viper.GetString("keyid"))
	alerter := NewSlackAlerter(viper.GetString("slack_token"))

	jobList := NewJobList(config, alerter)

	for {
		if err := mainLoop(jobList, requester); err != nil {
			log.Print(err)
		}
		time.Sleep(15 * time.Minute)
	}
}
