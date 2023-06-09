package main

import (
	"log"
	"os"
	"strconv"
	"tp1/common/middleware"
)

func main() {
	minimumDistance, err := strconv.ParseFloat(os.Getenv("MIN_DISTANCE"), 64)
	if err != nil {
		minimumDistance = 6
	}

	consumer, err := middleware.NewConsumer("consumer")
	if err != nil {
		log.Fatal(err)
	}
	producer, err := middleware.NewProducer("producer")
	if err != nil {
		log.Fatal(err)
	}
	distanceMerger := NewDistanceMerger(consumer, producer, minimumDistance)
	distanceMerger.Run()
}
