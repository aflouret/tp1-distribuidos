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
	nextStageInstances, err := strconv.Atoi(os.Getenv("NEXT_STAGE_INSTANCES"))
	if err != nil {
		nextStageInstances = 1
	}

	producer := middleware.NewProducer("distance_averager", nextStageInstances, true)
	consumer := middleware.NewConsumer("distance_calculator", "", previousStageInstances, instanceID)

	distanceCalculator := NewDistanceCalculator(producer, consumer)
	distanceCalculator.Run()
}
