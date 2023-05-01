package main

import (
	"fmt"
	"os"
	"strconv"
	"tp1/common/middleware"
)

func main() {
	nextStageInstances, err := strconv.Atoi(os.Getenv("NEXT_STAGE_INSTANCES"))
	if err != nil {
		nextStageInstances = 1
	}
	fmt.Printf("NEXT_STAGE_INSTANCES %v\n", nextStageInstances)
	tripsProducer := middleware.NewProducer("data_dropper", nextStageInstances, true)
	stationsProducer := middleware.NewProducer("stations", 1, false)
	weatherProducer := middleware.NewProducer("weather", 1, false)
	resultsConsumer := middleware.NewConsumer("results", "", 1, "0")

	clientHandler := NewClientHandler(tripsProducer, stationsProducer, weatherProducer, resultsConsumer)
	clientHandler.Run()
}
