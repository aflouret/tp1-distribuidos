package main

import (
	"fmt"
	"github.com/umahmood/haversine"
	"log"
	"net"
	"strconv"
	"strings"
	"tp1/common/protocol"
)

type trip struct {
	city             string
	startDate        string
	startStationCode string
	endDate          string
	endStationCode   string
	durationSec      string
	year             string
}

type joinedTrip struct {
	city                  string
	startStationName      string
	startStationLatitude  float64
	startStationLongitude float64
	endStationName        string
	endStationLatitude    float64
	endStationLongitude   float64
	year                  string
}

type average struct {
	avg   float64
	count int
}

type station struct {
	name      string
	latitude  float64
	longitude float64
}

type Server struct {
	stations                   map[string]station
	precipitationDatesByCity   map[string]map[string]bool
	avgDurationsByDate         map[string]average
	numberOfTripsByStation2016 map[string]int
	numberOfTripsByStation2017 map[string]int
	avgDistanceByEndStation    map[string]average
}

func NewServer() *Server {
	return &Server{
		stations:                   make(map[string]station),
		precipitationDatesByCity:   make(map[string]map[string]bool),
		avgDurationsByDate:         make(map[string]average),
		numberOfTripsByStation2016: make(map[string]int),
		numberOfTripsByStation2017: make(map[string]int),
		avgDistanceByEndStation:    make(map[string]average),
	}
}

func main() {
	server := NewServer()
	server.Run()
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server listening on port 12345")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("New connection from: %v\n", conn.RemoteAddr())
		s.handleConnection(conn)
	}
}

func getStationKey(code, year, city string) string {
	return fmt.Sprintf("%s-%s-%s", code, year, city)
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	msg, err := protocol.Recv(conn)
	if err != nil {
		fmt.Printf("Error reading from connection: %v\n", err)
		return
	}
	switch msg.Type {
	case protocol.BeginStations:
		s.handleStations(conn, msg.Payload)
	case protocol.BeginWeather:
		s.handleWeather(conn, msg.Payload)
	//case protocol.EndStaticData:
	//	s.handleEndStaticData(conn)
	case protocol.BeginTrips:
		s.handleTrips(conn, msg.Payload)
	case protocol.GetResults:
		s.handleResults(conn)
	}
}

func (s *Server) handleStations(conn net.Conn, city string) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})

	_, err := protocol.Recv(conn)
	if err != nil {
		fmt.Printf("Error reading from connection: %v\n", err)
		return
	}

	for {
		msg, err := protocol.Recv(conn)
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			return
		}
		if msg.Type != protocol.Data {
			if msg.Type != protocol.EndStations {
				fmt.Printf("Received wrong message: %v\n", msg)
				return
			}
			fmt.Println("Finished receiving stations from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}
		line := strings.TrimSpace(msg.Payload)
		s.storeStation(line, city)
	}
}

func (s *Server) handleWeather(conn net.Conn, city string) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})

	s.precipitationDatesByCity[city] = make(map[string]bool)

	_, err := protocol.Recv(conn)
	if err != nil {
		fmt.Printf("Error reading from connection: %v\n", err)
		return
	}

	for {
		msg, err := protocol.Recv(conn)
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			return
		}
		if msg.Type != protocol.Data {
			if msg.Type != protocol.EndWeather {
				fmt.Printf("Received wrong message: %v\n", msg)
				return
			}
			fmt.Println("Finished receiving weather from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}
		line := strings.TrimSpace(msg.Payload)
		s.storeWeather(line, city)
	}
}

func (s *Server) handleTrips(conn net.Conn, city string) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})

	_, err := protocol.Recv(conn)
	if err != nil {
		fmt.Printf("Error reading from connection: %v\n", err)
		return
	}

	for {
		msg, err := protocol.Recv(conn)
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			return
		}
		if msg.Type != protocol.Data {
			if msg.Type != protocol.EndTrips {
				fmt.Printf("Received wrong message: %v\n", msg)
				return
			}
			fmt.Println("Finished receiving trips from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}
		line := strings.TrimSpace(msg.Payload)
		t := parseTrip(line, city)
		s.processQuery1(t) // DURATION AVERAGER

		tripWithStation, err := s.joinStation(t) // STATION JOINER
		if err != nil {
			log.Fatal("not found")
		}
		if tripWithStation.year == "2016" { // YEAR FILTER
			s.count2016(tripWithStation) // 2016 COUNTER
		} else if tripWithStation.year == "2017" { // YEAR FILTER
			s.count2017(tripWithStation) // 2016 COUNTER

		}
		distance := calculateDistance(tripWithStation)               // DISTANCE CALCULATOR
		s.averageDistances(tripWithStation.endStationName, distance) // DISTANCE AVERAGER
	}
}

func (s *Server) handleResults(conn net.Conn) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})

	res1 := s.calculateResult1()
	protocol.Send(conn, protocol.NewDataMessage(res1))
	res2 := s.calculateResult2()
	protocol.Send(conn, protocol.NewDataMessage(res2))
	res3 := s.calculateResult3()
	protocol.Send(conn, protocol.NewDataMessage(res3))
}

func (s *Server) storeStation(csv string, city string) {
	csvFields := strings.Split(csv, ",")
	code := csvFields[0]
	name := csvFields[1]
	latitude, _ := strconv.ParseFloat(csvFields[2], 64)
	longitude, _ := strconv.ParseFloat(csvFields[3], 64)
	year := csvFields[4]

	key := getStationKey(code, year, city)
	s.stations[key] = station{name, latitude, longitude}
}

func (s *Server) storeWeather(csv string, city string) {
	csvFields := strings.Split(csv, ",")
	date := csvFields[0]
	prec, _ := strconv.ParseFloat(csvFields[1], 64)
	if prec > 30 {
		s.precipitationDatesByCity[city][date] = true
	}
}

func parseTrip(csv string, city string) trip {
	csvFields := strings.Split(csv, ",")
	startDate := strings.Split(csvFields[0], " ")[0]
	endDate := strings.Split(csvFields[2], " ")[0]
	return trip{
		city:             city,
		startDate:        startDate,
		startStationCode: csvFields[1],
		endDate:          endDate,
		endStationCode:   csvFields[3],
		durationSec:      csvFields[4],
		year:             csvFields[6],
	}
}

func (s *Server) joinStation(t trip) (joinedTrip, error) {
	startStationKey := getStationKey(t.startStationCode, t.year, t.city)
	startStation, ok := s.stations[startStationKey]
	if !ok {
		return joinedTrip{}, fmt.Errorf("station not found: %s", startStationKey)
	}
	endStationKey := getStationKey(t.endStationCode, t.year, t.city)
	endStation, ok := s.stations[endStationKey]
	if !ok {
		return joinedTrip{}, fmt.Errorf("station not found: %s", endStationKey)
	}

	return joinedTrip{
		city:                  t.city,
		startStationName:      startStation.name,
		startStationLatitude:  startStation.latitude,
		startStationLongitude: startStation.longitude,
		endStationName:        endStation.name,
		endStationLatitude:    endStation.latitude,
		endStationLongitude:   endStation.longitude,
		year:                  t.year,
	}, nil

}

func (s *Server) processQuery1(t trip) {
	if _, ok := s.precipitationDatesByCity[t.city][t.startDate]; ok {
		duration, _ := strconv.ParseFloat(t.durationSec, 64)
		if d, ok := s.avgDurationsByDate[t.startDate]; ok {
			newAvg := (d.avg*float64(d.count) + duration) / float64(d.count+1)
			d.avg = newAvg
			d.count++
			s.avgDurationsByDate[t.startDate] = d
		} else {
			s.avgDurationsByDate[t.startDate] = average{avg: duration, count: 1}
		}
	}
}

func (s *Server) count2016(t joinedTrip) {
	if count, ok := s.numberOfTripsByStation2016[t.startStationName]; ok {
		s.numberOfTripsByStation2016[t.startStationName] = count + 1
	} else {
		s.numberOfTripsByStation2016[t.startStationName] = 1
	}
}

func (s *Server) count2017(t joinedTrip) {
	if count, ok := s.numberOfTripsByStation2017[t.startStationName]; ok {
		s.numberOfTripsByStation2017[t.startStationName] = count + 1
	} else {
		s.numberOfTripsByStation2017[t.startStationName] = 1
	}
}

func calculateDistance(t joinedTrip) float64 {
	startCoordinates := haversine.Coord{t.startStationLatitude, t.startStationLongitude}
	endCoordinates := haversine.Coord{t.endStationLatitude, t.endStationLongitude}
	_, distance := haversine.Distance(startCoordinates, endCoordinates)
	return distance
}

func (s *Server) averageDistances(endStationName string, distance float64) {
	if d, ok := s.avgDistanceByEndStation[endStationName]; ok {
		newAvg := (d.avg*float64(d.count) + distance) / float64(d.count+1)
		d.avg = newAvg
		d.count++
		s.avgDistanceByEndStation[endStationName] = d
	} else {
		s.avgDistanceByEndStation[endStationName] = average{distance, 1}
	}
}

func (s *Server) calculateResult1() string {
	result := "Average duration of trips during >30mm precipitation days:\n"
	result += "start_date,average_duration\n"

	for k, v := range s.avgDurationsByDate { // DURATION MERGER
		result += fmt.Sprintf("%s,%f,%v\n", k, v.avg, v.count)
	}
	return result
}

func (s *Server) calculateResult2() string {
	result := "Stations that doubled the number of trips between 2016 and 2017:\n"
	result += "start_station_name,trips_count_2017,trips_count_2016\n"

	for k, count2017 := range s.numberOfTripsByStation2017 { // COUNT MERGER
		count2016 := s.numberOfTripsByStation2016[k]
		if count2017 >= count2016 {
			result += fmt.Sprintf("%s,%v,%v\n", k, count2017, count2016)
		}
	}
	return result
}

func (s *Server) calculateResult3() string {
	result := "Stations with more than 6km average to arrive at them:\n"
	result += "end_station_name,average_distance\n"

	for k, v := range s.avgDistanceByEndStation { // DISTANCE MERGER
		if v.avg > 1 {
			result += fmt.Sprintf("%s,%v,%v\n", k, v.avg, v.count)
		}
	}
	return result
}
