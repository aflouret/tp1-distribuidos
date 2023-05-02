package main

import (
	"log"
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
	nextStageInstances, err := strconv.Atoi(os.Getenv("NEXT_STAGE_INSTANCES"))
	if err != nil {
		nextStageInstances = 1
	}
	year := os.Getenv("YEAR")
	if year == "" {
		log.Fatal("Year not defined")
	}

	consumer := middleware.NewConsumer("trip_counter", year, previousStageInstances, instanceID)
	producer := middleware.NewProducer("count_merger", nextStageInstances, false)

	tripCounter := NewTripCounter(year, consumer, producer)
	tripCounter.Run()
}
