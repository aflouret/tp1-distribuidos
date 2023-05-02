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

	year1 := os.Getenv("YEAR_1")
	if year1 == "" {
		year1 = "2016"
	}

	year2 := os.Getenv("YEAR_2")
	if year2 == "" {
		year2 = "2017"
	}

	previousStageInstances, err := strconv.Atoi(os.Getenv("PREV_STAGE_INSTANCES"))
	if err != nil {
		previousStageInstances = 1
	}
	consumer := middleware.NewConsumer("count_merger", "", 2*previousStageInstances, instanceID)
	producer := middleware.NewProducer("results", 1, false)
	countMerger := NewCountMerger(consumer, producer, year1, year2)
	countMerger.Run()
}
