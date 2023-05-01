package main

import (
	"fmt"
	"strings"
	"time"
	"tp1/common/middleware"
)

const (
	idIndex = iota
	tripCityIndex
	tripStartDateIndex
	tripDurationIndex
)

type WeatherJoiner struct {
	producer                   *middleware.Producer
	tripsConsumer              *middleware.Consumer
	weatherConsumer            *middleware.Consumer
	precipitationsByDateByCity map[string]map[string]string
	endMessageReceived         bool
	msgCount                   int
	startTime                  time.Time
}

func main() {
	weatherJoiner := NewWeatherJoiner()
	weatherJoiner.Run()
}

func NewWeatherJoiner() *WeatherJoiner {
	producer := middleware.NewProducer("precipitation_filter")
	weatherConsumer := middleware.NewConsumer("weather", "")
	tripsConsumer := middleware.NewConsumer("weather_joiner_trips", "")
	precipitationsByDateByCity := make(map[string]map[string]string)
	return &WeatherJoiner{
		producer:                   producer,
		tripsConsumer:              tripsConsumer,
		weatherConsumer:            weatherConsumer,
		precipitationsByDateByCity: precipitationsByDateByCity,
		startTime:                  time.Now(),
	}
}

func (j *WeatherJoiner) Run() {
	defer j.producer.Close()

	j.weatherConsumer.Consume(j.processWeatherMessage)
	j.weatherConsumer.Close()
	j.tripsConsumer.Consume(j.processTripMessage)
	j.tripsConsumer.Close()
}

func (j *WeatherJoiner) processWeatherMessage(msg string) {
	if msg == "eof" {
		return
	}

	fields := strings.Split(msg, ",")
	city := fields[0]
	date := fields[1]
	precipitations := fields[2]

	if _, ok := j.precipitationsByDateByCity[city]; !ok {
		j.precipitationsByDateByCity[city] = make(map[string]string)
	}
	j.precipitationsByDateByCity[city][date] = precipitations

	//log.Printf("Received weather: %s\n", msg)
}

func (j *WeatherJoiner) processTripMessage(msg string) {
	if msg == "eof" {
		if !j.endMessageReceived {
			j.endMessageReceived = true
			j.producer.Produce(msg)
		}
		return
	}
	joinedTrip, err := j.joinWeather(msg)
	if err != nil {
		//fmt.Println("ERROR " + err.Error())
	} else {
		j.producer.Produce(joinedTrip)
	}
	if j.msgCount%10000 == 0 {
		fmt.Printf("Time: %s Received message %s\n", time.Since(j.startTime).String(), msg)
	}
	j.msgCount++
}

func (j *WeatherJoiner) joinWeather(csvTrip string) (string, error) {
	tripFields := strings.Split(csvTrip, ",")
	id := tripFields[idIndex]
	city := tripFields[tripCityIndex]
	startDate := tripFields[tripStartDateIndex]
	duration := tripFields[tripDurationIndex]

	precipitations, ok := j.precipitationsByDateByCity[city][startDate]
	if !ok {
		return "", fmt.Errorf("weather not found: %s %s", city, startDate)
	}

	joinedTrip := fmt.Sprintf("%s,%s,%s,%s",
		id,
		startDate,
		duration,
		precipitations,
	)

	return joinedTrip, nil
}
