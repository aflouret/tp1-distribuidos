package middleware

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type Consumer struct {
	conn           *amqp.Connection
	ch             *amqp.Channel
	msgChannel     <-chan amqp.Delivery
	eofsReceived   int
	producerCount  int
	sigtermChannel chan os.Signal
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func NewConsumer(exchangeName string, routingKey string, producerCount int, consumerID string) *Consumer {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		false,        // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,                // queue name
		routingKey+consumerID, // routing key
		exchangeName,          // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	err = ch.QueueBind(
		q.Name,       // queue name
		"eof",        // routing key
		exchangeName, // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	sigtermChannel := make(chan os.Signal, 1)
	signal.Notify(sigtermChannel, syscall.SIGTERM)
	return &Consumer{
		conn:           conn,
		ch:             ch,
		msgChannel:     msgs,
		eofsReceived:   0,
		producerCount:  producerCount,
		sigtermChannel: sigtermChannel,
	}
}

func (c *Consumer) Consume(processMessage func(string)) {
	for {
		select {
		case <-c.sigtermChannel:
			return
		case msg := <-c.msgChannel:
			msgBody := string(msg.Body)
			if msgBody == "eof" {
				fmt.Printf("Received eof %v\n", c.eofsReceived)
				c.eofsReceived++
				if c.eofsReceived == c.producerCount {
					fmt.Println("Received all eofs")
					processMessage(msgBody)
					return
				}
				continue
			}
			processMessage(msgBody)
		}

	}
}

func (c *Consumer) Close() {
	c.ch.Close()
	c.conn.Close()
}
