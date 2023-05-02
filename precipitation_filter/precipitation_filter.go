package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tp1/common/middleware"
)

const (
	idIndex = iota
	startDateIndex
	durationIndex
	precipitationsIndex
)

type PrecipitationFilter struct {
	producer  *middleware.Producer
	consumer  *middleware.Consumer
	msgCount  int
	startTime time.Time
}

func NewPrecipitationFilter(consumer *middleware.Consumer, producer *middleware.Producer) *PrecipitationFilter {
	return &PrecipitationFilter{
		producer: producer,
		consumer: consumer,
	}
}

func (f *PrecipitationFilter) Run() {
	defer f.consumer.Close()
	defer f.producer.Close()
	f.startTime = time.Now()

	f.consumer.Consume(f.processMessage)
}

func (f *PrecipitationFilter) processMessage(msg string) {
	if msg == "eof" {
		f.producer.PublishMessage(msg, "")
		return
	}
	//fmt.Println("Received message " + msg)

	if f.msgCount%10000 == 0 {
		fmt.Printf("Time: %s Received message %s\n", time.Since(f.startTime).String(), msg)
	}
	f.msgCount++
	f.filterAndSend(msg)
}

func (f *PrecipitationFilter) filterAndSend(msg string) error {
	fields := strings.Split(msg, ",")
	precipitationsString := fields[precipitationsIndex]
	precipitations, err := strconv.ParseFloat(precipitationsString, 64)
	if err != nil {
		return err
	}
	if precipitations > 30 {
		id := fields[idIndex]
		startDate := fields[startDateIndex]
		duration := fields[durationIndex]
		msgToSend := fmt.Sprintf("%s,%s,%s", id, startDate, duration)
		f.producer.PublishMessage(msgToSend, "")
		//fmt.Printf("Sent message %s\n", msgToSend)
	}
	return nil
}
