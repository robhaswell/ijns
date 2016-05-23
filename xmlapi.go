package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type IndustryJobsRequester interface {
	GetXML(string, string) ([]byte, error)
}

type XmlApiIndustryJobsRequester struct{}

func (xmlapi *XmlApiIndustryJobsRequester) GetXML(vCode string, keyID string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf(
		"https://api.eveonline.com/corp/IndustryJobs.xml.aspx?keyID=%s&vCode=%s",
		keyID, vCode))
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type IndustryJobs struct {
	XMLName xml.Name `xml:"eveapi"`
	Jobs []Job `xml:"result>rowset>row"`
}

type Job struct {
	XMLName       xml.Name `xml:"row"`
	ID            int      `xml:"jobID,attr"`
	Blueprint     string   `xml:"blueprintTypeName,attr"`
	Installer     string   `xml:"installerName,attr"`
	EndDateString string   `xml:"endDate,attr"`
	EndDate       time.Time
	Alerter       Alerter
}

func ParseXmlApiResponse(body []byte) ([]Job, error) {
	jobs := IndustryJobs{}
	err := xml.Unmarshal(body, &jobs)
	return jobs.Jobs, err
}
