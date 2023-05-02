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

	minimumPrecipitations, err := strconv.ParseFloat(os.Getenv("MIN_PRECIPITATIONS"), 64)
	if err != nil {
		minimumPrecipitations = 30
	}

	previousStageInstances, err := strconv.Atoi(os.Getenv("PREV_STAGE_INSTANCES"))
	if err != nil {
		previousStageInstances = 1
	}
	nextStageInstances, err := strconv.Atoi(os.Getenv("NEXT_STAGE_INSTANCES"))
	if err != nil {
		nextStageInstances = 1
	}
	consumer := middleware.NewConsumer("precipitation_filter", "", previousStageInstances, instanceID)
	producer := middleware.NewProducer("duration_averager", nextStageInstances, true)

	precipitationFilter := NewPrecipitationFilter(consumer, producer, minimumPrecipitations)
	precipitationFilter.Run()
}
