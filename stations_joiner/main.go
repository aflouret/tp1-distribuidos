package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"tp1/common/middleware"
)

const (
	tripCityIndex = iota
	tripStartStationCodeIndex
	tripEndStationCodeIndex
	tripYearIdIndex
)

type station struct {
	name      string
	latitude  string
	longitude string
}

type StationsJoiner struct {
	producer           *middleware.Producer
	tripsConsumer      *middleware.Consumer
	stationsConsumer   *middleware.Consumer
	stations           map[string]station
	endMessageReceived bool
}

func main() {
	stationsJoiner := NewStationsJoiner()
	stationsJoiner.Run()
}

func NewStationsJoiner() *StationsJoiner {
	stationsConsumer := middleware.NewConsumer("stations", "")
	tripsConsumer := middleware.NewConsumer("stations_joiner_trips", "")
	//producer := middleware.NewProducer("stations_joiner_trips")
	stations := make(map[string]station)
	return &StationsJoiner{
		tripsConsumer:    tripsConsumer,
		stationsConsumer: stationsConsumer,
		stations:         stations,
	}
}

func (j *StationsJoiner) Run() {
	//defer j.producer.Close()

	j.stationsConsumer.Consume(j.processStationMessage)
	j.stationsConsumer.Close()
	j.tripsConsumer.Consume(j.processTripMessage)
	j.tripsConsumer.Close()
}

func (j *StationsJoiner) processStationMessage(msg string) {
	if msg == "eof" {
		if !j.endMessageReceived {
			j.endMessageReceived = true
			//j.producer.Produce(msg)
		}
		return
	}

	fields := strings.Split(msg, ",")
	city := fields[0]
	code := fields[1]
	name := fields[2]
	latitude := fields[3]
	longitude := fields[4]
	year := fields[5]

	key := getStationKey(code, year, city)
	j.stations[key] = station{name, latitude, longitude}

	log.Printf("Received station: %s\n", msg)
}

func (j *StationsJoiner) processTripMessage(msg string) {
	if msg == "eof" {
		if !j.endMessageReceived {
			j.endMessageReceived = true
			//j.producer.Produce(msg)
		}
		return
	}
	joinedTrip, _ := j.joinStation(msg)
	log.Printf("joined trip: %s\n", joinedTrip)
}

func getStationKey(code, year, city string) string {
	return fmt.Sprintf("%s-%s-%s", code, year, city)
}

func (j *StationsJoiner) joinStation(csvTrip string) (string, error) {

	tripFields := strings.Split(csvTrip, ",")

	city := tripFields[tripCityIndex]
	startStationCode := tripFields[tripStartStationCodeIndex]
	endStationCode := tripFields[tripEndStationCodeIndex]
	year := tripFields[tripYearIdIndex]

	startStationKey := getStationKey(startStationCode, year, city)
	startStation, ok := j.stations[startStationKey]
	if !ok {
		return "", fmt.Errorf("station not found: %s", startStationKey)
	}
	endStationKey := getStationKey(endStationCode, year, city)
	endStation, ok := j.stations[endStationKey]
	if !ok {
		return "", fmt.Errorf("station not found: %s", endStationKey)
	}

	joinedTrip := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s",
		city,
		startStation.name,
		startStation.latitude,
		startStation.longitude,
		endStation.name,
		endStation.latitude,
		endStation.longitude,
		year,
	)

	return joinedTrip, nil
}
