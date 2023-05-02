package middleware

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
	"strings"
)

type Producer struct {
	conn          *amqp.Connection
	ch            *amqp.Channel
	exchangeName  string
	consumerCount int
	routeByID     bool
}

func NewProducer(exchangeName string, consumerCount int, routeByID bool) *Producer {
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

	return &Producer{
		conn:          conn,
		ch:            ch,
		exchangeName:  exchangeName,
		consumerCount: consumerCount,
		routeByID:     routeByID,
	}
}

func (p *Producer) PublishMessage(msg string, routingKey string) {
	if msg == "eof" {
		routingKey = "eof"
	} else if p.routeByID {
		msgIDString := strings.Split(msg, ",")[0]
		msgID, _ := strconv.Atoi(msgIDString)
		consumerID := msgID % p.consumerCount
		routingKey += fmt.Sprintf("%v", consumerID)
	} else {
		routingKey += "0"
	}

	//fmt.Printf("Routing key: %s, message: %s\n", routingKey, msg)
	err := p.ch.PublishWithContext(context.TODO(),
		p.exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "text/plain",
			Body:         []byte(msg),
		},
	)
	failOnError(err, "Failed to publish a message")
}

func (p *Producer) Close() {
	p.ch.Close()
	p.conn.Close()
}
