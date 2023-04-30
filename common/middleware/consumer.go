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

func NewConsumer(exchangeName string, routingKey string) *Consumer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

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
	failOnError(err, "Failed to register a consumer")

	return &Consumer{
		conn:       conn,
		ch:         ch,
		msgChannel: msgs,
	}
}

func (c *Consumer) Consume(processMessage func(string)) {
	for msg := range c.msgChannel {
		msgBody := string(msg.Body)
		processMessage(msgBody)
		msg.Ack(false)
		if msgBody == "eof" {
			return
		}
	}
}

func (c *Consumer) Close() {
	c.ch.Close()
	c.conn.Close()
}
