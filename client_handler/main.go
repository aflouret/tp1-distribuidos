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

type Server struct {
	tripsProducer    *middleware.Producer
	stationsProducer *middleware.Producer
	weatherProducer  *middleware.Producer
	resultsConsumer  *middleware.Consumer
}

func NewServer() *Server {
	tripsProducer := middleware.NewProducer("data_dropper")
	stationsProducer := middleware.NewProducer("stations")
	weatherProducer := middleware.NewProducer("weather")
	resultsConsumer := middleware.NewConsumer("results", "")
	return &Server{
		tripsProducer:    tripsProducer,
		stationsProducer: stationsProducer,
		weatherProducer:  weatherProducer,
		resultsConsumer:  resultsConsumer,
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
	s.tripsProducer.Close()
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
	case protocol.EndStaticData:
		s.handleEndStaticData(conn)
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

		station := city + "," + strings.TrimSpace(msg.Payload)
		s.stationsProducer.Produce(station)
	}
}

func (s *Server) handleWeather(conn net.Conn, city string) {
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
			if msg.Type != protocol.EndWeather {
				fmt.Printf("Received wrong message: %v\n", msg)
				return
			}
			fmt.Println("Finished receiving weather from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}

		weather := city + "," + strings.TrimSpace(msg.Payload)
		s.weatherProducer.Produce(weather)
	}
}

func (s *Server) handleTrips(conn net.Conn, city string) {
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
				fmt.Printf("Received wrong message: %v, \n", msg.Type)
				return
			}
			fmt.Println("Finished receiving trips from " + city)
			protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
			return
		}
		lines := strings.Split(msg.Payload, ";")
		for _, line := range lines {
			id := strconv.Itoa(tripCounter)
			trip := id + "," + city + "," + strings.TrimSpace(line)
			s.tripsProducer.Produce(trip)
			if tripCounter%10000 == 0 {
				fmt.Printf("Time: %s Received message %s\n", time.Since(startTime).String(), id)
			}
			tripCounter++
		}
	}
}

func (s *Server) handleEndStaticData(conn net.Conn) {
	s.stationsProducer.Produce("eof")
	s.weatherProducer.Produce("eof")
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
}

func (s *Server) handleResults(conn net.Conn) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
	s.tripsProducer.Produce("eof")
	s.resultsConsumer.Consume(func(msg string) {
		protocol.Send(conn, protocol.NewDataMessage(msg))
	})

}
