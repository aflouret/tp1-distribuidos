package main

import (
	"os"
	"tp1/common/middleware"
)

func main() {
	instanceID := os.Getenv("ID")
	if instanceID == "" {
		instanceID = "0"
	}

	stationsConsumer := middleware.NewConsumer("stations", "", 1, "0")
	tripsConsumer := middleware.NewConsumer("stations_joiner_trips", "", 1, instanceID)
	yearFilterProducer := middleware.NewProducer("year_filter", 1, false)

	stationsJoiner := NewStationsJoiner(stationsConsumer, tripsConsumer, yearFilterProducer)
	stationsJoiner.Run()
}
