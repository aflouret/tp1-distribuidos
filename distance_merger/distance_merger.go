package main

import (
	"fmt"
	"strconv"
	"strings"
	"tp1/common/middleware"
)

const (
	endStationNameIndex = iota
	averageIndex
	countIndex
)

type average struct {
	avg   float64
	count int
}

type DistanceMerger struct {
	producer                  *middleware.Producer
	consumer                  *middleware.Consumer
	averageDistancesByStation map[string]average
}

func NewDistanceMerger(consumer *middleware.Consumer, producer *middleware.Producer) *DistanceMerger {
	averageDistancesByStation := make(map[string]average)

	return &DistanceMerger{
		producer:                  producer,
		consumer:                  consumer,
		averageDistancesByStation: averageDistancesByStation,
	}
}

func (m *DistanceMerger) Run() {
	defer m.consumer.Close()
	defer m.producer.Close()

	m.consumer.Consume(m.processMessage)
	m.sendResults()
}

func (m *DistanceMerger) processMessage(msg string) {
	if msg == "eof" {
		return
	}

	//fmt.Println("Received message " + msg)

	m.mergeResults(msg)
}

func (m *DistanceMerger) mergeResults(msg string) error {
	fields := strings.Split(msg, ",")
	endStationName := fields[endStationNameIndex]
	avg, err := strconv.ParseFloat(fields[averageIndex], 64)
	if err != nil {
		return err
	}
	count, err := strconv.Atoi(fields[countIndex])
	if err != nil {
		return err
	}

	if d, ok := m.averageDistancesByStation[endStationName]; ok {
		newAvg := (d.avg*float64(d.count) + avg*float64(count)) / float64(d.count+count)
		d.avg = newAvg
		d.count += count
		m.averageDistancesByStation[endStationName] = d
	} else {
		m.averageDistancesByStation[endStationName] = average{avg: avg, count: count}
	}
	return nil
}

func (m *DistanceMerger) sendResults() {
	result := "Stations with more than 6km average to arrive at them:\n"
	result += "end_station_name,average_distance\n"

	for k, v := range m.averageDistancesByStation { // DISTANCE MERGER
		if v.avg > 5 {
			result += fmt.Sprintf("%s,%v\n", k, v.avg)
		}
	}
	m.producer.PublishMessage(result, "")
	fmt.Println(result)
}
