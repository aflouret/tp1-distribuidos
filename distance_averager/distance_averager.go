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
	endStationNameIndex
	distanceIndex
)

type average struct {
	avg   float64
	count int
}

type DistanceAverager struct {
	producer              *middleware.Producer
	consumer              *middleware.Consumer
	avgDistancesByStation map[string]average
	msgCount              int
	startTime             time.Time
}

func NewDistanceAverager(consumer *middleware.Consumer, producer *middleware.Producer) *DistanceAverager {
	avgDistancesByStation := make(map[string]average)

	return &DistanceAverager{
		producer:              producer,
		consumer:              consumer,
		avgDistancesByStation: avgDistancesByStation,
	}
}

func (a *DistanceAverager) Run() {
	defer a.consumer.Close()
	defer a.producer.Close()

	a.startTime = time.Now()
	a.consumer.Consume(a.processMessage)
	a.sendResults()
}

func (a *DistanceAverager) processMessage(msg string) {
	if msg == "eof" {
		return
	}
	//fmt.Println(msg)
	if a.msgCount%10000 == 0 {
		fmt.Printf("Time: %s Received message %s\n", time.Since(a.startTime).String(), msg)
	}
	a.msgCount++
	a.updateAverage(msg)
}

func (a *DistanceAverager) updateAverage(msg string) error {
	fields := strings.Split(msg, ",")
	endStationName := fields[endStationNameIndex]
	distance, err := strconv.ParseFloat(fields[distanceIndex], 64)
	if err != nil {
		return err
	}

	if d, ok := a.avgDistancesByStation[endStationName]; ok {
		newAvg := (d.avg*float64(d.count) + distance) / float64(d.count+1)
		d.avg = newAvg
		d.count++
		a.avgDistancesByStation[endStationName] = d
	} else {
		a.avgDistancesByStation[endStationName] = average{avg: distance, count: 1}
	}
	return nil
}

func (a *DistanceAverager) sendResults() {
	for k, v := range a.avgDistancesByStation {
		result := fmt.Sprintf("%s,%v,%v", k, v.avg, v.count)
		a.producer.PublishMessage(result, "")
	}
	a.producer.PublishMessage("eof", "")
}
