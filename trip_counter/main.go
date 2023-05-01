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

	consumer := middleware.NewConsumer("trip_counter", "", 1, instanceID)
	producer := middleware.NewProducer("count_merger", 1, false)

	tripCounter := NewTripCounter(consumer, producer)
	tripCounter.Run()
}
