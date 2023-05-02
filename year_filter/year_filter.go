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
	year1    string
	year2    string
}

func NewYearFilter(producer *middleware.Producer, consumer *middleware.Consumer, year1 string, year2 string) *YearFilter {
	return &YearFilter{
		producer: producer,
		consumer: consumer,
		year1:    year1,
		year2:    year2,
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
	if year == f.year1 || year == f.year2 {
		startStationName := fields[startStationNameIndex]
		id := fields[idIndex]
		messageToSend := id + "," + startStationName
		f.producer.PublishMessage(messageToSend, year)
		//fmt.Printf("Sent message %s,%s\n", id, startStationName)
	}
	return nil
}
