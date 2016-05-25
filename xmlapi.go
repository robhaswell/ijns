package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type IndustryJobsRequester interface {
	GetXML() ([]byte, error)
}

type XmlApiIndustryJobsRequester struct {
	vCode, keyID string
}

func NewXmlApiIndustryJobsRequester(vCode, keyID string) *XmlApiIndustryJobsRequester {
	xmlapi := &XmlApiIndustryJobsRequester{}
	xmlapi.SetApiCredentials(vCode, keyID)
	return xmlapi
}

func (xmlapi *XmlApiIndustryJobsRequester) SetApiCredentials(vCode, keyID string) {
	xmlapi.vCode = vCode
	xmlapi.keyID = keyID
}

func (xmlapi *XmlApiIndustryJobsRequester) GetXML() ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf(
		"https://api.eveonline.com/corp/IndustryJobs.xml.aspx?keyID=%s&vCode=%s",
		xmlapi.keyID, xmlapi.vCode))
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type FakeIndustryJobsRequester struct {
	xmlResponse []byte
}

func (fake *FakeIndustryJobsRequester) SetResponse(xml []byte) {
	fake.xmlResponse = xml
}

func (fake *FakeIndustryJobsRequester) GetXML() ([]byte, error) {
	return fake.xmlResponse, nil
}

type IndustryJobs struct {
	XMLName xml.Name `xml:"eveapi"`
	Jobs    []Job    `xml:"result>rowset>row"`
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
