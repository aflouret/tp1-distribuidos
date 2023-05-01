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
	consumer := middleware.NewConsumer("count_merger", "", 1, instanceID)
	producer := middleware.NewProducer("results", 1, false)
	countMerger := NewCountMerger(consumer, producer)
	countMerger.Run()
}
