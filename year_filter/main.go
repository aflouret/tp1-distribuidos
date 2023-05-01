package main

import (
	"fmt"
	"strings"
	"tp1/common/middleware"
)

const (
	startStationNameIndex = iota
	yearIndex
)

type YearFilter struct {
	producer           *middleware.Producer
	consumer           *middleware.Consumer
	endMessageReceived bool
}

func main() {
	yearFilter := NewYearFilter()
	yearFilter.Run()
}

func NewYearFilter() *YearFilter {
	consumer := middleware.NewConsumer("year_filter", "")
	producer := middleware.NewProducer("trip_counter")

	return &YearFilter{
		producer: producer,
		consumer: consumer,
	}
}

func (f *YearFilter) Run() {
	defer f.consumer.Close()
	defer f.producer.Close()

	f.consumer.Consume(f.processMessage)
}

func (f *YearFilter) processMessage(msg string) {
	if msg == "eof" {
		if !f.endMessageReceived {
			f.endMessageReceived = true
			f.producer.Produce(msg)
		}
		return
	}
	fmt.Println("Received message " + msg)

	f.filterAndSend(msg)
}

func (f *YearFilter) filterAndSend(msg string) error {
	fields := strings.Split(msg, ",")
	year := fields[yearIndex]
	if year == "2016" || year == "2017" {
		startStationName := fields[startStationNameIndex]
		f.producer.Produce(startStationName)
		fmt.Printf("Sent message %s\n", startStationName)
	}
	return nil
}
