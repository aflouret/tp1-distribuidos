package middleware

import (
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type Consumer struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	msgChannel <-chan amqp.Delivery
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func NewConsumer(queueName string) *Consumer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,
		0,
		false,
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a server")

	return &Consumer{
		conn:       conn,
		ch:         ch,
		msgChannel: msgs,
	}
}

func (c *Consumer) Consume(callback func(string)) {
	for d := range c.msgChannel {
		callback(string(d.Body))
		d.Ack(false)
	}
}

func (c *Consumer) Close() {
	c.ch.Close()
	c.conn.Close()
}
