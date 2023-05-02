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
	yearFilterInstances, err := strconv.Atoi(os.Getenv("YEAR_FILTER_INSTANCES"))
	if err != nil {
		yearFilterInstances = 1
	}
	distanceCalculatorInstances, err := strconv.Atoi(os.Getenv("DISTANCE_CALCULATOR_INSTANCES"))
	if err != nil {
		distanceCalculatorInstances = 1
	}

	stationsConsumer := middleware.NewConsumer("stations", "", 1, "0")
	tripsConsumer := middleware.NewConsumer("stations_joiner_trips", "", previousStageInstances, instanceID)
	yearFilterProducer := middleware.NewProducer("year_filter", yearFilterInstances, true)
	distanceCalculatorProducer := middleware.NewProducer("distance_calculator", distanceCalculatorInstances, true)

	stationsJoiner := NewStationsJoiner(tripsConsumer, stationsConsumer, yearFilterProducer, distanceCalculatorProducer)
	stationsJoiner.Run()
}
