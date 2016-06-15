package main

import (
	"log"

	"github.com/jonboulle/clockwork"
	"github.com/spf13/viper"
)

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

	clock := clockwork.NewRealClock()

	requester := NewXmlApiIndustryJobsRequester(viper.GetString("vcode"), viper.GetString("keyid"))
	alerter := NewSlackAlerter(viper.GetString("slack_token"))

	jobList := NewJobList(config, clock, alerter)
	fetcher := NewFetcher(clock, jobList.Ch, requester)

	// Begin the job requesting thread
	go fetcher.Loop()

	// Begin the primary job collection and alerting thread.
	jobList.Loop()
}
