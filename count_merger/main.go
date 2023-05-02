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
	consumer := middleware.NewConsumer("count_merger", "", 2*previousStageInstances, instanceID)
	producer := middleware.NewProducer("results", 1, false)
	countMerger := NewCountMerger(consumer, producer)
	countMerger.Run()
}
