package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"tp1/common/middleware"
)

const (
	tripCityIndex = iota
	tripStartDateIndex
	tripDurationIndex
)

type WeatherJoiner struct {
	producer                   *middleware.Producer
	tripsConsumer              *middleware.Consumer
	weatherConsumer            *middleware.Consumer
	precipitationsByDateByCity map[string]map[string]string
	endMessageReceived         bool
}

func main() {
	weatherJoiner := NewWeatherJoiner()
	weatherJoiner.Run()
}

func NewWeatherJoiner() *WeatherJoiner {
	weatherConsumer := middleware.NewConsumer("weather", "")
	tripsConsumer := middleware.NewConsumer("weather_joiner_trips", "")
	//producer := middleware.NewProducer("weather_joiner_trips")
	precipitationsByDateByCity := make(map[string]map[string]string)
	return &WeatherJoiner{
		tripsConsumer:              tripsConsumer,
		weatherConsumer:            weatherConsumer,
		precipitationsByDateByCity: precipitationsByDateByCity,
	}
}

func (j *WeatherJoiner) Run() {
	//defer j.producer.Close()

	j.weatherConsumer.Consume(j.processWeatherMessage)
	j.weatherConsumer.Close()
	j.tripsConsumer.Consume(j.processTripMessage)
	j.tripsConsumer.Close()
}

func (j *WeatherJoiner) processWeatherMessage(msg string) {
	if msg == "eof" {
		if !j.endMessageReceived {
			j.endMessageReceived = true
			//j.producer.Produce(msg)
		}
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

	log.Printf("Received weather: %s\n", msg)
}

func (j *WeatherJoiner) processTripMessage(msg string) {
	if msg == "eof" {
		if !j.endMessageReceived {
			j.endMessageReceived = true
			//j.producer.Produce(msg)
		}
		return
	}
	joinedTrip, _ := j.joinWeather(msg)
	log.Printf("joined trip: %s\n", joinedTrip)
}

func (j *WeatherJoiner) joinWeather(csvTrip string) (string, error) {
	tripFields := strings.Split(csvTrip, ",")
	city := tripFields[tripCityIndex]
	startDate := tripFields[tripStartDateIndex]
	duration := tripFields[tripDurationIndex]

	precipitations, ok := j.precipitationsByDateByCity[city][startDate]
	if !ok {
		return "", fmt.Errorf("weather not found: %s %s", city, startDate)
	}

	joinedTrip := fmt.Sprintf("%s,%s,%s",
		startDate,
		duration,
		precipitations,
	)

	return joinedTrip, nil
}
