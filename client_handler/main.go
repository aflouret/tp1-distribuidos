package main

import (
	"fmt"
	"net"
	"strings"
	"tp1/common/middleware"
	"tp1/common/protocol"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
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

	producer := middleware.NewProducer("data_dropper")
	defer producer.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("New connection from: %v\n", conn.RemoteAddr())
		s.handleConnection(conn, producer)
	}
}

func (s *Server) handleConnection(conn net.Conn, producer *middleware.Producer) {
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
		s.handleTrips(conn, producer, msg.Payload)
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
	}
}

func (s *Server) handleTrips(conn net.Conn, producer *middleware.Producer, city string) {
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

		trip := city + "," + strings.TrimSpace(msg.Payload)

		producer.Produce(trip)
	}
}

func (s *Server) handleResults(conn net.Conn) {
	protocol.Send(conn, protocol.Message{Type: protocol.Ack, Payload: ""})
	protocol.Send(conn, protocol.NewDataMessage("OK"))
	protocol.Send(conn, protocol.NewDataMessage("OK"))
	protocol.Send(conn, protocol.NewDataMessage("OK"))
}
