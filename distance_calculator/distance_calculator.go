package main

import (
	"fmt"
	"github.com/umahmood/haversine"
	"strconv"
	"strings"
	"tp1/common/middleware"
)

const (
	idIndex = iota
	cityIndex
	startStationNameIndex
	startStationLatitudeIndex
	startStationLongitudeIndex
	endStationNameIndex
	endStationLatitudeIndex
	endStationLongitudeIndex
	yearIndex
)

type DistanceCalculator struct {
	producer *middleware.Producer
	consumer *middleware.Consumer
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

	c.consumer.Consume(c.processMessage)
}

func (c *DistanceCalculator) processMessage(msg string) {
	if msg == "eof" {
		c.producer.PublishMessage(msg, "")
		return
	}
	//fmt.Println("Received message " + msg)

	tripWithDistance, err := c.calculateDistance(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.producer.PublishMessage(tripWithDistance, "")
}

func (c *DistanceCalculator) calculateDistance(msg string) (string, error) {
	fields := strings.Split(msg, ",")
	id := fields[idIndex]
	endStationName := fields[endStationNameIndex]

	startStationLatitude, err := strconv.ParseFloat(fields[startStationLatitudeIndex], 64)
	if err != nil {
		return "", fmt.Errorf("error parsing start station latitude: %w", err)
	}
	startStationLongitude, err := strconv.ParseFloat(fields[startStationLongitudeIndex], 64)
	if err != nil {
		return "", fmt.Errorf("error parsing start station longitude: %w", err)
	}
	endStationLatitude, err := strconv.ParseFloat(fields[endStationLatitudeIndex], 64)
	if err != nil {
		return "", fmt.Errorf("error parsing end station latitude: %w", err)
	}
	endStationLongitude, err := strconv.ParseFloat(fields[endStationLongitudeIndex], 64)
	if err != nil {
		return "", fmt.Errorf("error parsing end station longitude: %w", err)
	}

	startCoordinates := haversine.Coord{startStationLatitude, startStationLongitude}
	endCoordinates := haversine.Coord{endStationLatitude, endStationLongitude}
	_, distance := haversine.Distance(startCoordinates, endCoordinates)

	return fmt.Sprintf("%s,%s,%v", id, endStationName, distance), nil
}
