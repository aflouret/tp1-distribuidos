package main

import (
	"fmt"
	"strconv"
	"strings"
	"tp1/common/middleware"
)

const (
	startDateIndex = iota
	durationIndex
	precipitationsIndex
)

type PrecipitationFilter struct {
	producer           *middleware.Producer
	consumer           *middleware.Consumer
	endMessageReceived bool
}

func main() {
	precipitationFilter := NewPrecipitationFilter()
	precipitationFilter.Run()
}

func NewPrecipitationFilter() *PrecipitationFilter {
	consumer := middleware.NewConsumer("precipitation_filter", "")
	producer := middleware.NewProducer("duration_averager")

	return &PrecipitationFilter{
		producer: producer,
		consumer: consumer,
	}
}

func (f *PrecipitationFilter) Run() {
	defer f.consumer.Close()
	defer f.producer.Close()

	f.consumer.Consume(f.processMessage)
}

func (f *PrecipitationFilter) processMessage(msg string) {
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

func (f *PrecipitationFilter) filterAndSend(msg string) error {
	fields := strings.Split(msg, ",")
	precipitationsString := fields[precipitationsIndex]
	precipitations, err := strconv.ParseFloat(precipitationsString, 64)
	if err != nil {
		return err
	}
	if precipitations > 30 {
		startDate := fields[startDateIndex]
		duration := fields[durationIndex]
		msgToSend := fmt.Sprintf("%s,%s", startDate, duration)
		f.producer.Produce(msgToSend)
		fmt.Printf("Sent message %s\n", msgToSend)
	}
	return nil
}
