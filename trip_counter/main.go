package main

import (
	"fmt"
	"strings"
	"tp1/common/middleware"
)

const (
	startStationNameIndex = iota
)

type TripCounter struct {
	producer           *middleware.Producer
	consumer           *middleware.Consumer
	countByStation     map[string]int
	endMessageReceived bool
	year               string
}

func main() {
	tripCounter := NewTripCounter()
	tripCounter.Run()
}

func NewTripCounter() *TripCounter {
	consumer := middleware.NewConsumer("trip_counter", "")
	producer := middleware.NewProducer("count_merger")
	countByStation := make(map[string]int)

	return &TripCounter{
		producer:       producer,
		consumer:       consumer,
		countByStation: countByStation,
		year:           "2016",
	}
}

func (a *TripCounter) Run() {
	defer a.consumer.Close()
	defer a.producer.Close()

	a.consumer.Consume(a.processMessage)
	a.sendResults()
	for k, v := range a.countByStation {
		fmt.Printf("%s,%v\n", k, v)
	}
}

func (a *TripCounter) processMessage(msg string) {
	if msg == "eof" {
		if !a.endMessageReceived {
			a.endMessageReceived = true
		}
		return
	}
	fmt.Println("Received message " + msg)

	a.updateCount(msg)
}

func (a *TripCounter) updateCount(msg string) error {
	fields := strings.Split(msg, ",")
	startStationName := fields[startStationNameIndex]
	if c, ok := a.countByStation[startStationName]; ok {
		a.countByStation[startStationName] = c + 1
	} else {
		a.countByStation[startStationName] = 1
	}
	return nil
}

func (a *TripCounter) sendResults() {
	for k, v := range a.countByStation {
		result := fmt.Sprintf("%s,%s,%v", a.year, k, v)
		a.producer.Produce(result)
		fmt.Println(result)
	}
	a.producer.Produce("eof")
}
