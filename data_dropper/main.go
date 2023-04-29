package main

import (
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"tp1/common/middleware"
)

const (
	cityIndex             = 0
	startDateIndex        = 1
	startStationCodeIndex = 2
	endStationCodeIndex   = 4
	durationSecIndex      = 5
	yearIdIndex           = 7
)

var columnsForWeatherJoiner = []int{cityIndex, startDateIndex, durationSecIndex}
var columnsForStationsJoiner = []int{cityIndex, startStationCodeIndex, endStationCodeIndex, yearIdIndex}

func main() {
	consumer := middleware.NewConsumer("data_dropper")
	defer consumer.Close()

	consumer.Consume(processMessage)
}

func processMessage(msg string) {
	fields := strings.Split(msg, ",")
	sanitize(fields)
	sendToWeatherJoiner(fields)
	sendToStationsJoiner(fields)
}

func sanitize(fields []string) {
	duration, err := strconv.ParseFloat(fields[durationSecIndex], 64)
	if err != nil || duration < 0 {
		fields[durationSecIndex] = "0"
	}

	day := strings.Split(fields[startDateIndex], " ")[0]
	fields[startDateIndex] = day
}

func sendToWeatherJoiner(fields []string) {
	var fieldsToSend []string
	for _, col := range columnsForWeatherJoiner {
		fieldsToSend = append(fieldsToSend, fields[col])
	}
	trip := strings.Join(fieldsToSend, ",")
	log.Printf("Sent trip to weather joiner: %s\n", trip)
}

func sendToStationsJoiner(fields []string) {
	var fieldsToSend []string
	for _, col := range columnsForStationsJoiner {
		fieldsToSend = append(fieldsToSend, fields[col])
	}
	trip := strings.Join(fieldsToSend, ",")
	log.Printf("Sent trip to stations joiner: %s\n", trip)
}
