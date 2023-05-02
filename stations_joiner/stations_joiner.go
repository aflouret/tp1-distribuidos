package main

import (
	"fmt"
	"strings"
	"tp1/common/middleware"
)

const (
	idIndex = iota
	tripCityIndex
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
	yearFilterProducer         *middleware.Producer
	distanceCalculatorProducer *middleware.Producer
	tripsConsumer              *middleware.Consumer
	stationsConsumer           *middleware.Consumer
	stations                   map[string]station
}

func NewStationsJoiner(
	tripsConsumer *middleware.Consumer,
	stationsConsumer *middleware.Consumer,
	yearFilterProducer *middleware.Producer,
	distanceCalculatorProducer *middleware.Producer,
) *StationsJoiner {
	stations := make(map[string]station)
	return &StationsJoiner{
		tripsConsumer:              tripsConsumer,
		stationsConsumer:           stationsConsumer,
		yearFilterProducer:         yearFilterProducer,
		distanceCalculatorProducer: distanceCalculatorProducer,
		stations:                   stations,
	}
}

func (j *StationsJoiner) Run() {
	defer j.yearFilterProducer.Close()
	defer j.distanceCalculatorProducer.Close()

	j.stationsConsumer.Consume(j.processStationMessage)
	j.stationsConsumer.Close()
	j.tripsConsumer.Consume(j.processTripMessage)
	j.tripsConsumer.Close()
}

func (j *StationsJoiner) processStationMessage(msg string) {
	//fmt.Printf("Received station: %s\n", msg)
	if msg == "eof" {
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

}

func (j *StationsJoiner) processTripMessage(msg string) {
	//log.Printf("Received trip: %s\n", msg)
	if msg == "eof" {
		j.yearFilterProducer.PublishMessage(msg, "")
		j.distanceCalculatorProducer.PublishMessage(msg, "")
		return
	}
	joinedTrip, err := j.joinStation(msg)
	if err != nil {
		fmt.Printf("Error joining station %s: %s\n", msg, err.Error())
		return
	}
	j.sendToYearFilter(joinedTrip)
	j.sendToDistanceCalculator(joinedTrip)
	//log.Printf("joined trip: %s\n", joinedTrip)
}

func getStationKey(code, year, city string) string {
	return fmt.Sprintf("%s-%s-%s", code, year, city)
}

func (j *StationsJoiner) joinStation(csvTrip string) (string, error) {

	tripFields := strings.Split(csvTrip, ",")

	id := tripFields[idIndex]
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

	joinedTrip := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s",
		id,
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

func (j *StationsJoiner) sendToYearFilter(trip string) {
	fields := strings.Split(trip, ",")

	id := fields[0]
	startStationName := fields[2]
	year := fields[8]

	tripToSend := fmt.Sprintf("%s,%s,%s", id, startStationName, year)
	j.yearFilterProducer.PublishMessage(tripToSend, "")
}

func (j *StationsJoiner) sendToDistanceCalculator(trip string) {
	fields := strings.Split(trip, ",")
	city := fields[1]
	if city == "montreal" {
		j.distanceCalculatorProducer.PublishMessage(trip, "")
	}
}
