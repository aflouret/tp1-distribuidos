package main

import (
	"fmt"
	"strings"
	"tp1/common/middleware"
)

const (
	idIndex = iota
	startStationNameIndex
)

type TripCounter struct {
	producer       *middleware.Producer
	consumer       *middleware.Consumer
	countByStation map[string]int
	year           string
}

func NewTripCounter(year string, consumer *middleware.Consumer, producer *middleware.Producer) *TripCounter {
	countByStation := make(map[string]int)

	return &TripCounter{
		producer:       producer,
		consumer:       consumer,
		countByStation: countByStation,
		year:           year,
	}
}

func (a *TripCounter) Run() {
	defer a.consumer.Close()
	defer a.producer.Close()

	a.consumer.Consume(a.processMessage)
	a.sendResults()
}

func (a *TripCounter) processMessage(msg string) {
	if msg == "eof" {
		return
	}
	//fmt.Println("Received message " + msg)

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
		a.producer.PublishMessage(result, "")
		//fmt.Println(result)
	}
	a.producer.PublishMessage("eof", "")
}
