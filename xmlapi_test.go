package main

import (
	"reflect"
	"testing"
)

func TestParseXmlApiResponse(t *testing.T) {
	xml := []byte(`<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <currentTime>2016-05-23 19:51:43</currentTime>
  <result>
    <rowset name="jobs" key="jobID" columns="jobID,installerID,installerName,facilityID,solarSystemID,solarSystemName,stationID,activityID,blueprintID,blueprintTypeID,blueprintTypeName,blueprintLocationID,outputLocationID,runs,cost,teamID,licensedRuns,probability,productTypeID,productTypeName,status,timeInSeconds,startDate,endDate,pauseDate,completedDate,completedCharacterID,successfulRuns">
      <row jobID="1" installerID="1" installerName="Test Character" facilityID="1018140767062" solarSystemID="30001688" solarSystemName="Onazel" stationID="1019017237703" activityID="1" blueprintID="1020936400343" blueprintTypeID="31797" blueprintTypeName="Test Item Blueprint I" blueprintLocationID="1018140767062" outputLocationID="1018140767062" runs="5" cost="2687951.00" teamID="0" licensedRuns="1" probability="1" productTypeID="31796" productTypeName="Medium Core Defense Field Extender II" status="1" timeInSeconds="72364" startDate="2016-01-02 03:04:05" endDate="2016-01-02 06:07:08" pauseDate="0001-01-01 00:00:00" completedDate="0001-01-01 00:00:00" completedCharacterID="0" successfulRuns="0" />
      <row jobID="2" installerID="1" installerName="Test Character" facilityID="1018140767062" solarSystemID="30001688" solarSystemName="Onazel" stationID="1019017237703" activityID="1" blueprintID="1020936400343" blueprintTypeID="31797" blueprintTypeName="Test Item Blueprint I" blueprintLocationID="1018140767062" outputLocationID="1018140767062" runs="5" cost="2687951.00" teamID="0" licensedRuns="1" probability="1" productTypeID="31796" productTypeName="Medium Core Defense Field Extender II" status="1" timeInSeconds="72364" startDate="2016-01-02 03:04:05" endDate="2016-01-02 06:07:08" pauseDate="0001-01-01 00:00:00" completedDate="0001-01-01 00:00:00" completedCharacterID="0" successfulRuns="0" />
    </rowset>
  </result>
  <cachedUntil>2016-05-23 19:59:14</cachedUntil>
</eveapi>`)

	expected := []Job{
		Job{
			ID:            1,
			Blueprint:     "Test Item Blueprint I",
			Installer:     "Test Character",
			EndDateString: "2016-01-02 06:07:08",
		},
		Job{
			ID:            2,
			Blueprint:     "Test Item Blueprint I",
			Installer:     "Test Character",
			EndDateString: "2016-01-02 06:07:08",
		},
	}
	result, err := ParseXmlApiResponse(xml)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(result, expected) {
		t.Fatalf("Unexpected result. Expected:\n\n%v\n\nGot:\n\n%", expected, result)
	}
}
