package queue

import (
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQClient() (*amqp.Channel, error) {
	mqURL := os.Getenv("RABBITMQ_URL")
	if mqURL == "" {
		mqURL = "amqp://guest:guest@localhost:5672/"
	}
	rabbitMQConn, err := amqp.Dial(mqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := rabbitMQConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return channel, nil
}
