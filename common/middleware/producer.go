package middleware

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
)

type Producer struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	config ProducerConfig
}

type ProducerConfig struct {
	connectionString   string
	exchangeName       string
	nextStageInstances int
	routeByID          bool
}

func newProducerConfig(configID string) (ProducerConfig, error) {
	v := viper.New()
	v.SetConfigFile("./middleware_config.yaml")
	if err := v.ReadInConfig(); err != nil {
		return ProducerConfig{}, fmt.Errorf("Configuration for producer %s could not be read from config file\n", configID)
	}
	exchangeName := v.GetString(fmt.Sprintf("%s.exchange_name", configID))
	routeByID := v.GetBool(fmt.Sprintf("%s.route_by_id", configID))
	nextStageInstancesEnv := v.GetString(fmt.Sprintf("%s.next_stage_instances_env", configID))
	nextStageInstances, err := strconv.Atoi(os.Getenv(nextStageInstancesEnv))
	if err != nil {
		nextStageInstances = 1
	}
	connectionString := os.Getenv("RABBITMQ_CONNECTION_STRING")

	return ProducerConfig{
		connectionString:   connectionString,
		exchangeName:       exchangeName,
		nextStageInstances: nextStageInstances,
		routeByID:          routeByID,
	}, nil
}

func NewProducer(configID string) (*Producer, error) {
	config, err := newProducerConfig(configID)
	if err != nil {
		return nil, err
	}
	conn, err := amqp.Dial(config.connectionString)
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		config.exchangeName, // name
		"direct",            // type
		false,               // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	return &Producer{
		conn:   conn,
		ch:     ch,
		config: config,
	}, nil
}

func (p *Producer) PublishMessage(msg string, routingKey string) {
	if msg == "eof" {
		routingKey = "eof"
	} else if p.config.routeByID {
		msgIDString := strings.Split(msg, ",")[0]
		msgID, _ := strconv.Atoi(msgIDString)
		consumerID := msgID % p.config.nextStageInstances
		routingKey += fmt.Sprintf("%v", consumerID)
	}

	//fmt.Printf("Routing key: %s, message: %s\n", routingKey, msg)
	err := p.ch.PublishWithContext(context.TODO(),
		p.config.exchangeName,
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
