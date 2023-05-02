package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"tp1/common/middleware"
	"tp1/common/protocol"
)

type ClientHandler struct {
	tripsProducer    *middleware.Producer
	stationsProducer *middleware.Producer
	weatherProducer  *middleware.Producer
	resultsConsumer  *middleware.Consumer
}

func NewClientHandler(
	tripsProducer *middleware.Producer,
	stationsProducer *middleware.Producer,
	weatherProducer *middleware.Producer,
	resultsConsumer *middleware.Consumer,
) *ClientHandler {
	return &ClientHandler{
		tripsProducer:    tripsProducer,
		stationsProducer: stationsProducer,
		weatherProducer:  weatherProducer,
		resultsConsumer:  resultsConsumer,
	}
}

func (s *ClientHandler) Run() {
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("ClientHandler listening on port 12345")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("New connection from: %v\n", conn.RemoteAddr())
		s.handleConnection(conn)
	}
	s.tripsProducer.Close()
}

func (s *ClientHandler) handleConnection(conn net.Conn) {
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
	case protocol.EndStaticData:
		s.handleEndStaticData(conn)
	case protocol.BeginTrips:
		s.handleTrips(conn, msg.Payload)
	case protocol.GetResults:
		s.handleResults(conn)
	}
}

func (s *ClientHandler) handleStations(conn net.Conn, city string) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})

	for {
		msg, err := protocol.Recv(conn)
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			return
		}
		if msg.Type != protocol.Data {
			if msg.Type != protocol.EndStations {
				fmt.Printf("Received invalid message: %v\n", msg)
				return
			}
			fmt.Println("Finished receiving stations from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}
		protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
		lines := strings.Split(msg.Payload, ";")
		for _, line := range lines {
			station := city + "," + strings.TrimSpace(line)
			s.stationsProducer.PublishMessage(station, "")
		}
	}
}

func (s *ClientHandler) handleWeather(conn net.Conn, city string) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})

	for {
		msg, err := protocol.Recv(conn)
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			return
		}
		if msg.Type != protocol.Data {
			if msg.Type != protocol.EndWeather {
				fmt.Printf("Received invalid message: %v\n", msg)
				return
			}
			fmt.Println("Finished receiving weather from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}
		protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
		lines := strings.Split(msg.Payload, ";")
		for _, line := range lines {
			weather := city + "," + strings.TrimSpace(line)
			s.weatherProducer.PublishMessage(weather, "")
		}

	}
}

func (s *ClientHandler) handleTrips(conn net.Conn, city string) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})

	startTime := time.Now()
	tripCounter := 0
	for {
		msg, err := protocol.Recv(conn)
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			return
		}
		if msg.Type != protocol.Data {
			if msg.Type != protocol.EndTrips {
				fmt.Printf("Received invalid message: %v, \n", msg.Type)
				return
			}
			fmt.Println("Finished receiving trips from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}
		protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
		lines := strings.Split(msg.Payload, ";")
		for _, line := range lines {
			id := strconv.Itoa(tripCounter)
			trip := id + "," + city + "," + strings.TrimSpace(line)
			s.tripsProducer.PublishMessage(trip, "")
			if tripCounter%10000 == 0 {
				fmt.Printf("Time: %s Received message %s\n", time.Since(startTime).String(), id)
			}
			tripCounter++
		}
	}
}

func (s *ClientHandler) handleEndStaticData(conn net.Conn) {
	s.stationsProducer.PublishMessage("eof", "")
	s.weatherProducer.PublishMessage("eof", "")
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
}

func (s *ClientHandler) handleResults(conn net.Conn) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
	s.tripsProducer.PublishMessage("eof", "")
	s.resultsConsumer.Consume(func(msg string) {
		protocol.Send(conn, protocol.NewDataMessage(msg))
	})

}
