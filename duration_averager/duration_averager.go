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
)

type average struct {
	avg   float64
	count int
}

type DurationAverager struct {
	producer           *middleware.Producer
	consumer           *middleware.Consumer
	avgDurationsByDate map[string]average
	msgCount           int
	startTime          time.Time
}

func NewDurationAverager(consumer *middleware.Consumer, producer *middleware.Producer) *DurationAverager {
	avgDurationsByDate := make(map[string]average)

	return &DurationAverager{
		producer:           producer,
		consumer:           consumer,
		avgDurationsByDate: avgDurationsByDate,
	}
}

func (a *DurationAverager) Run() {
	defer a.consumer.Close()
	defer a.producer.Close()

	a.startTime = time.Now()
	a.consumer.Consume(a.processMessage)
	a.sendResults()
	for k, v := range a.avgDurationsByDate {
		fmt.Printf("%s %v %v\n", k, v.avg, v.count)
	}
}

func (a *DurationAverager) processMessage(msg string) {
	if msg == "eof" {
		return
	}
	//fmt.Println(msg)
	if a.msgCount%100 == 0 {
		fmt.Printf("Time: %s Received message %s\n", time.Since(a.startTime).String(), msg)
	}
	a.msgCount++
	a.updateAverage(msg)
}

func (a *DurationAverager) updateAverage(msg string) error {
	fields := strings.Split(msg, ",")
	startDate := fields[startDateIndex]
	duration, err := strconv.ParseFloat(fields[durationIndex], 64)
	if err != nil {
		return err
	}

	if d, ok := a.avgDurationsByDate[startDate]; ok {
		newAvg := (d.avg*float64(d.count) + duration) / float64(d.count+1)
		d.avg = newAvg
		d.count++
		a.avgDurationsByDate[startDate] = d
	} else {
		a.avgDurationsByDate[startDate] = average{avg: duration, count: 1}
	}
	return nil
}

func (a *DurationAverager) sendResults() {
	for k, v := range a.avgDurationsByDate {
		result := fmt.Sprintf("%s,%v,%v", k, v.avg, v.count)
		a.producer.PublishMessage(result)
	}
	a.producer.PublishMessage("eof")
}
