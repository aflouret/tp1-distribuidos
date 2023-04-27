package client

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tp1/client/internal/utils"
	"tp1/common/protocol"
)

var cities = []string{"montreal", "toronto", "washington"}
var ErrClientTerminated = errors.New("client terminated")

type Config struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	BatchSize     int
}

type Client struct {
	config Config
	conn   net.Conn
}

func NewClient(config Config) *Client {
	client := &Client{
		config: config,
	}
	return client
}

func (c *Client) connectToServer() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) StartClient() {
	err := c.sendStationsToServer()
	if err != nil {
		log.Errorf("action: send_stations | result: fail | error: %v", err)
		return
	}
	err = c.sendWeatherToServer()
	if err != nil {
		log.Errorf("action: send_weather | result: fail | error: %v", err)
		return
	}
	err = c.sendTripsToServer()
	if err != nil {
		log.Errorf("action: send_trips | result: fail | error: %v", err)
		return
	}
	err = c.getResults()
	if err != nil {
		log.Errorf("action: request_results | result: fail | error: %v", err)
		return
	}
	log.Infof("action: send_trips | result: success")
}

func (c *Client) sendStationsToServer() error {
	sigtermNotifier := make(chan os.Signal, 1)
	signal.Notify(sigtermNotifier, syscall.SIGTERM)

	for _, city := range cities {
		file, err := os.Open(fmt.Sprintf("data/%s/stations.csv", city))
		if err != nil {
			log.Fatal(err)
		}

		err = c.connectToServer()
		if err != nil {
			return err
		}

		err = c.notifyBeginStations(city)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		for {
			station, err := utils.ReadLine(scanner)
			if err != nil {
				return err
			}
			if len(station) == 0 {
				err = c.notifyEndStations(city)
				if err != nil {
					return err
				}
				break
			}

			if err = c.sendDataMessage(station); err != nil {
				return err
			}

			select {
			case <-sigtermNotifier:
				log.Debugf("action: terminate_client | result: success | client_id: %v", c.config.ID)
				return ErrClientTerminated
			default:
			}
		}
		c.conn.Close()
		file.Close()
	}

	return nil
}

func (c *Client) sendWeatherToServer() error {
	sigtermNotifier := make(chan os.Signal, 1)
	signal.Notify(sigtermNotifier, syscall.SIGTERM)

	for _, city := range cities {
		file, err := os.Open(fmt.Sprintf("data/%s/weather.csv", city))
		if err != nil {
			log.Fatal(err)
		}

		err = c.connectToServer()
		if err != nil {
			return err
		}

		err = c.notifyBeginWeather(city)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		for {
			weather, err := utils.ReadLine(scanner)
			if err != nil {
				return err
			}
			if len(weather) == 0 {
				err = c.notifyEndWeather(city)
				if err != nil {
					return err
				}
				break
			}

			if err = c.sendDataMessage(weather); err != nil {
				return err
			}

			select {
			case <-sigtermNotifier:
				log.Debugf("action: terminate_client | result: success | client_id: %v", c.config.ID)
				return ErrClientTerminated
			default:
			}
		}
		c.conn.Close()
		file.Close()
	}

	return nil
}

func (c *Client) sendTripsToServer() error {
	sigtermNotifier := make(chan os.Signal, 1)
	signal.Notify(sigtermNotifier, syscall.SIGTERM)

	for _, city := range cities {
		file, err := os.Open(fmt.Sprintf("data/%s/trips-small.csv", city))
		if err != nil {
			log.Fatal(err)
		}

		err = c.connectToServer()
		if err != nil {
			return err
		}

		err = c.notifyBeginTrips(city)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		for {
			trip, err := utils.ReadLine(scanner)
			if err != nil {
				return err
			}
			if len(trip) == 0 {
				err = c.notifyEndTrips(city)
				if err != nil {
					return err
				}
				break
			}

			if err = c.sendDataMessage(trip); err != nil {
				return err
			}

			select {
			case <-sigtermNotifier:
				log.Debugf("action: terminate_client | result: success | client_id: %v", c.config.ID)
				return ErrClientTerminated
			default:
			}
		}
		c.conn.Close()
		file.Close()
	}

	return nil
}

func (c *Client) getResults() error {
	err := c.connectToServer()
	if err != nil {
		return err
	}

	err = c.sendResultsRequest()
	if err != nil {
		return err
	}

	result1, err := c.getResult()
	if err != nil {
		return err
	}
	fmt.Println(result1)

	result2, err := c.getResult()
	if err != nil {
		return err
	}
	fmt.Println(result2)

	result3, err := c.getResult()
	if err != nil {
		return err
	}
	fmt.Println(result3)
	return nil
}

func (c *Client) notifyBeginStations(city string) error {
	return utils.SendControlMessage(c.conn, protocol.BeginStations, city)
}

func (c *Client) notifyEndStations(city string) error {
	return utils.SendControlMessage(c.conn, protocol.EndStations, city)
}

func (c *Client) notifyBeginWeather(city string) error {
	return utils.SendControlMessage(c.conn, protocol.BeginWeather, city)
}

func (c *Client) notifyEndWeather(city string) error {
	return utils.SendControlMessage(c.conn, protocol.EndWeather, city)
}

func (c *Client) notifyBeginTrips(city string) error {
	return utils.SendControlMessage(c.conn, protocol.BeginTrips, city)
}

func (c *Client) notifyEndTrips(city string) error {
	return utils.SendControlMessage(c.conn, protocol.EndTrips, city)
}

func (c *Client) sendResultsRequest() error {
	return utils.SendControlMessage(c.conn, protocol.GetResults, "")
}

func (c *Client) getResult() (string, error) {
	msg, err := protocol.Recv(c.conn)
	if err != nil {
		return "", err
	}
	return msg.Payload, nil
}

func (c *Client) sendDataMessage(payload string) error {
	return utils.SendDataMessage(c.conn, payload)
}
