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
	nextStageInstances, err := strconv.Atoi(os.Getenv("NEXT_STAGE_INSTANCES"))
	if err != nil {
		nextStageInstances = 1
	}

	producer := middleware.NewProducer("trip_counter", nextStageInstances, true)
	consumer := middleware.NewConsumer("year_filter", "", previousStageInstances, instanceID)

	yearFilter := NewYearFilter(producer, consumer, year1, year2)
	yearFilter.Run()
}
