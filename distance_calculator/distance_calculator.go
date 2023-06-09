package main

import (
	"fmt"
	"github.com/umahmood/haversine"
	"strconv"
	"strings"
	"time"
	"tp1/common/middleware"
	"tp1/common/utils"
)

const (
	startStationNameIndex = iota
	startStationLatitudeIndex
	startStationLongitudeIndex
	endStationNameIndex
	endStationLatitudeIndex
	endStationLongitudeIndex
	yearIndex
)

type DistanceCalculator struct {
	producer  *middleware.Producer
	consumer  *middleware.Consumer
	msgCount  int
	startTime time.Time
}

func NewDistanceCalculator(producer *middleware.Producer, consumer *middleware.Consumer) *DistanceCalculator {
	return &DistanceCalculator{
		producer: producer,
		consumer: consumer,
	}
}

func (c *DistanceCalculator) Run() {
	defer c.consumer.Close()
	defer c.producer.Close()

	c.startTime = time.Now()
	c.consumer.Consume(c.processMessage)
}

func (c *DistanceCalculator) processMessage(msg string) {
	if msg == "eof" {
		c.producer.PublishMessage(msg, "")
		return
	}

	id, _, trips := utils.ParseBatch(msg)

	tripsWithDistance := c.calculateDistance(trips)

	if len(tripsWithDistance) > 0 {
		tripsWithDistanceBatch := utils.CreateBatch(id, "", tripsWithDistance)
		c.producer.PublishMessage(tripsWithDistanceBatch, "")

		if c.msgCount%20000 == 0 {
			fmt.Printf("Time: %s Received batch %v\n", time.Since(c.startTime).String(), id)
		}
	}

	c.msgCount++

}

func (c *DistanceCalculator) calculateDistance(trips []string) []string {
	tripsWithDistance := make([]string, 0, len(trips))
	for _, trip := range trips {
		fields := strings.Split(trip, ",")
		endStationName := fields[endStationNameIndex]

		startStationLatitude, err := strconv.ParseFloat(fields[startStationLatitudeIndex], 64)
		if err != nil {
			fmt.Println(fmt.Errorf("error parsing start station latitude: %w", err))
			continue
		}
		startStationLongitude, err := strconv.ParseFloat(fields[startStationLongitudeIndex], 64)
		if err != nil {
			fmt.Println(fmt.Errorf("error parsing start station longitude: %w", err))
			continue
		}
		endStationLatitude, err := strconv.ParseFloat(fields[endStationLatitudeIndex], 64)
		if err != nil {
			fmt.Println(fmt.Errorf("error parsing end station latitude: %w", err))
			continue
		}
		endStationLongitude, err := strconv.ParseFloat(fields[endStationLongitudeIndex], 64)
		if err != nil {
			fmt.Println(fmt.Errorf("error parsing end station longitude: %w", err))
			continue
		}

		startCoordinates := haversine.Coord{startStationLatitude, startStationLongitude}
		endCoordinates := haversine.Coord{endStationLatitude, endStationLongitude}
		_, distance := haversine.Distance(startCoordinates, endCoordinates)

		tripWithDistance := fmt.Sprintf("%s,%v", endStationName, distance)
		tripsWithDistance = append(tripsWithDistance, tripWithDistance)
	}
	return tripsWithDistance
}
