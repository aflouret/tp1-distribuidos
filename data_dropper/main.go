package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tp1/common/middleware"
)

const (
	idIndex               = 0
	cityIndex             = 1
	startDateIndex        = 2
	startStationCodeIndex = 3
	endStationCodeIndex   = 5
	durationSecIndex      = 6
	yearIdIndex           = 8
)

var columnsForWeatherJoiner = []int{idIndex, cityIndex, startDateIndex, durationSecIndex}
var columnsForStationsJoiner = []int{idIndex, cityIndex, startStationCodeIndex, endStationCodeIndex, yearIdIndex}

type DataDropper struct {
	stationsJoinerProducer *middleware.Producer
	weatherJoinerProducer  *middleware.Producer
	consumer               *middleware.Consumer
	endMessageReceived     bool
	msgCount               int
	startTime              time.Time
}

func main() {
	dataDropper := NewDataDropper()
	dataDropper.Run()
}

func NewDataDropper() *DataDropper {
	consumer := middleware.NewConsumer("data_dropper", "")
	stationsJoinerProducer := middleware.NewProducer("stations_joiner_trips")
	weatherJoinerProducer := middleware.NewProducer("weather_joiner_trips")

	return &DataDropper{
		stationsJoinerProducer: stationsJoinerProducer,
		weatherJoinerProducer:  weatherJoinerProducer,
		consumer:               consumer,
		startTime:              time.Now(),
	}
}

func (d *DataDropper) Run() {
	defer d.consumer.Close()
	defer d.stationsJoinerProducer.Close()
	defer d.weatherJoinerProducer.Close()

	d.consumer.Consume(d.processMessage)
}

func (d *DataDropper) processMessage(msg string) {
	if msg == "eof" {
		if !d.endMessageReceived {
			d.endMessageReceived = true
			d.stationsJoinerProducer.Produce(msg)
			d.weatherJoinerProducer.Produce(msg)
		}
		return
	}
	fields := strings.Split(msg, ",")
	d.sanitize(fields)
	d.sendToWeatherJoiner(fields)
	d.sendToStationsJoiner(fields)
	if d.msgCount%10000 == 0 {
		fmt.Printf("Time: %s Received message %s\n", time.Since(d.startTime).String(), msg)
	}
	d.msgCount++
}

func (d *DataDropper) sanitize(fields []string) {
	duration, err := strconv.ParseFloat(fields[durationSecIndex], 64)
	if err != nil || duration < 0 {
		fields[durationSecIndex] = "0"
	}

	day := strings.Split(fields[startDateIndex], " ")[0]
	fields[startDateIndex] = day
}

func (d *DataDropper) sendToWeatherJoiner(fields []string) {
	var fieldsToSend []string
	for _, col := range columnsForWeatherJoiner {
		fieldsToSend = append(fieldsToSend, fields[col])
	}
	trip := strings.Join(fieldsToSend, ",")
	d.weatherJoinerProducer.Produce(trip)
	//log.Printf("Sent trip to weather joiner: %s\n", trip)
}

func (d *DataDropper) sendToStationsJoiner(fields []string) {
	var fieldsToSend []string
	for _, col := range columnsForStationsJoiner {
		fieldsToSend = append(fieldsToSend, fields[col])
	}
	trip := strings.Join(fieldsToSend, ",")

	d.stationsJoinerProducer.Produce(trip)
	//log.Printf("Sent trip to stations joiner: %s\n", trip)
}
