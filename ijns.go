package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

const SLACK_NAME = "agrakari"

var CHARACTERS = map[string]bool{"Maaya Saraki": true, "Indy Drone 4": true}

type Result struct {
	XMLName xml.Name `xml:"result"`
	Rowset  Rowset   `xml:"rowset"`
}

type Rowset struct {
	XMLName xml.Name `xml:"rowset"`
	Row     []Row    `xml:"row"`
}

type Row struct {
	XMLName   xml.Name `xml:"row"`
	Blueprint string   `xml:"blueprintTypeName,attr"`
	Installer string   `xml:"installerName,attr"`
}

func poll() {
	vCode := viper.GetString("vcode")
	keyID := viper.GetString("keyid")

	resp, err := http.Get(fmt.Sprintf(
		"https://api.eveonline.com/corp/IndustryJobs.xml.aspx?keyID=%s&vCode=%s",
			keyID, vCode))
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(string(body))

	var result Result
	xml.Unmarshal(body, &result)
	log.Print(result)
}

func main() {
	viper.SetEnvPrefix("ijns")
	viper.BindEnv("vcode")
	viper.BindEnv("keyid")

	for {
		poll()
		time.Sleep(15 * time.Minute)
	}
}
