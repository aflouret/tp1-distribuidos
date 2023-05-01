package middleware

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	exchangeName string
}

func NewProducer(exchangeName string) *Producer {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		false,        // durable
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
	err := p.ch.PublishWithContext(context.TODO(),
		p.exchangeName,
		"",
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
