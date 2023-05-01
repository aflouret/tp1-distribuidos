package main

import (
	"os"
	"strconv"
	"tp1/common/middleware"
)

func main() {
	instanceID := os.Getenv("ID")
	if instanceID == "" {
		instanceID = "0"
	}

	previousStageInstances, err := strconv.Atoi(os.Getenv("PREV_STAGE_INSTANCES"))
	if err != nil {
		previousStageInstances = 1
	}
	stationsJoinerInstances, err := strconv.Atoi(os.Getenv("STATIONS_JOINER_INSTANCES"))
	if err != nil {
		stationsJoinerInstances = 1
	}
	weatherJoinerInstances, err := strconv.Atoi(os.Getenv("WEATHER_JOINER_INSTANCES"))
	if err != nil {
		weatherJoinerInstances = 1
	}

	consumer := middleware.NewConsumer("data_dropper", "", previousStageInstances, instanceID)
	stationsJoinerProducer := middleware.NewProducer("stations_joiner_trips", stationsJoinerInstances, true)
	weatherJoinerProducer := middleware.NewProducer("weather_joiner_trips", weatherJoinerInstances, true)

	dataDropper := NewDataDropper(consumer, stationsJoinerProducer, weatherJoinerProducer)
	dataDropper.Run()
}
