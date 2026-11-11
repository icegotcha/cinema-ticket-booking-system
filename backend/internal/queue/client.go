package queue

import (
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client owns both the AMQP connection and its channel.
type Client struct {
	*amqp.Channel
	connection *amqp.Connection
}

func (c *Client) Close() error {
	return c.connection.Close()
}

func NewRabbitMQClient() (*Client, error) {
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
		_ = rabbitMQConn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &Client{Channel: channel, connection: rabbitMQConn}, nil
}
