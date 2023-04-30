package middleware

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Producer struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	exchangeName string
}

func NewProducer(exchangeName string) *Producer {
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

	return &Producer{
		conn:         conn,
		ch:           ch,
		exchangeName: exchangeName,
	}
}

func (p *Producer) Produce(msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := p.ch.PublishWithContext(ctx,
		p.exchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
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
