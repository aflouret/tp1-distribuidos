package main

import (
	"strings"
	"tp1/common/middleware"
)

const (
	idIndex = iota
	startStationNameIndex
	yearIndex
)

type YearFilter struct {
	producer *middleware.Producer
	consumer *middleware.Consumer
}

func NewYearFilter(producer *middleware.Producer, consumer *middleware.Consumer) *YearFilter {
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
		f.producer.PublishMessage(msg, "")
		return
	}
	//fmt.Println("Received message " + msg)

	f.filterAndSend(msg)
}

func (f *YearFilter) filterAndSend(msg string) error {
	fields := strings.Split(msg, ",")
	year := fields[yearIndex]
	if year == "2016" || year == "2017" {
		startStationName := fields[startStationNameIndex]
		id := fields[idIndex]
		messageToSend := id + "," + startStationName
		f.producer.PublishMessage(messageToSend, year)
		//fmt.Printf("Sent message %s,%s\n", id, startStationName)
	}
	return nil
}
