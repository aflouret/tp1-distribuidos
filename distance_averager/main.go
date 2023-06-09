package main

import (
	"log"
	"tp1/common/middleware"
)

func main() {
	consumer, err := middleware.NewConsumer("consumer")
	if err != nil {
		log.Fatal(err)
	}
	producer, err := middleware.NewProducer("producer")
	if err != nil {
		log.Fatal(err)
	}
	distanceAverager := NewDistanceAverager(consumer, producer)
	distanceAverager.Run()
}
