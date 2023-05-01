package main

import (
	"fmt"
	"strconv"
	"strings"
	"tp1/common/middleware"
)

const (
	yearIndex = iota
	startStationNameIndex
	countIndex
)

type CountMerger struct {
	producer            *middleware.Producer
	consumer            *middleware.Consumer
	year1               string
	year2               string
	countByStationYear1 map[string]int
	countByStationYear2 map[string]int
	endMessageReceived  bool
}

func main() {
	countMerger := NewCountMerger()
	countMerger.Run()
}

func NewCountMerger() *CountMerger {
	consumer := middleware.NewConsumer("count_merger", "")
	producer := middleware.NewProducer("results")
	countByStationYear1 := make(map[string]int)
	countByStationYear2 := make(map[string]int)
	return &CountMerger{
		producer:            producer,
		consumer:            consumer,
		year1:               "2016",
		year2:               "2017",
		countByStationYear1: countByStationYear1,
		countByStationYear2: countByStationYear2,
	}
}

func (m *CountMerger) Run() {
	defer m.consumer.Close()
	defer m.producer.Close()

	m.consumer.Consume(m.processMessage)
	m.sendResults()
}

func (m *CountMerger) processMessage(msg string) {
	if msg == "eof" {
		if !m.endMessageReceived {
			m.endMessageReceived = true
		}
		return
	}
	fmt.Println("Received message " + msg)

	m.mergeResults(msg)
}

func (m *CountMerger) mergeResults(msg string) error {
	fields := strings.Split(msg, ",")
	year := fields[yearIndex]
	startStationName := fields[startStationNameIndex]
	count, err := strconv.Atoi(fields[countIndex])
	if err != nil {
		return err
	}

	if year == m.year1 {
		if c, ok := m.countByStationYear1[startStationName]; ok {
			m.countByStationYear1[startStationName] = c + count
		} else {
			m.countByStationYear1[startStationName] = count
		}
	} else if year == m.year2 {
		if c, ok := m.countByStationYear2[startStationName]; ok {
			m.countByStationYear2[startStationName] = c + count
		} else {
			m.countByStationYear2[startStationName] = count
		}
	}

	return nil
}

func (m *CountMerger) sendResults() {
	result := fmt.Sprintf("Stations that doubled the number of trips between %s and %s:\n", m.year1, m.year2)
	result += fmt.Sprintf("start_station_name,trips_count_%s,trips_count_%s\n", m.year2, m.year1)

	for k, countYear2 := range m.countByStationYear2 {
		countYear1 := m.countByStationYear1[k]
		if countYear2 >= countYear1 {
			result += fmt.Sprintf("%s,%v,%v\n", k, countYear2, countYear1)
		}
	}

	m.producer.Produce(result)
	fmt.Println(result)
}
