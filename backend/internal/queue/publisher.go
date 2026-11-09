package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	channel       *amqp.Channel
	confirmations <-chan amqp.Confirmation
	returns       <-chan amqp.Return
	mu            sync.Mutex
}

func NewPublisher(ch *amqp.Channel) (*Publisher, error) {
	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("enable publisher confirms: %w", err)
	}

	return &Publisher{
		channel:       ch,
		confirmations: ch.NotifyPublish(make(chan amqp.Confirmation, 1)),
		returns:       ch.NotifyReturn(make(chan amqp.Return, 1)),
	}, nil
}

func (p *Publisher) Publish(
	ctx context.Context,
	queue string,
	payload any,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := p.channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare queue %s: %w", queue, err)
	}

	deliveryTag := p.channel.GetNextPublishSeqNo()
	if err := p.channel.PublishWithContext(
		ctx,
		"",
		queue,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("publish to queue %s: %w", queue, err)
	}

	returned := false
	for {
		select {
		case _, ok := <-p.returns:
			if !ok {
				return fmt.Errorf("publisher return channel closed")
			}
			returned = true
		case confirmation, ok := <-p.confirmations:
			if !ok {
				return fmt.Errorf("publisher confirmation channel closed")
			}
			// A previous publish may have timed out while its confirmation was
			// in flight. Discard that stale confirmation before waiting for ours.
			if confirmation.DeliveryTag < deliveryTag {
				continue
			}
			if confirmation.DeliveryTag != deliveryTag {
				return fmt.Errorf("unexpected publisher confirmation tag %d", confirmation.DeliveryTag)
			}
			if !confirmation.Ack {
				return fmt.Errorf("broker rejected message for queue %s", queue)
			}
			// RabbitMQ sends basic.return before the corresponding confirm. The
			// buffered return notification is therefore ready by this point.
			select {
			case <-p.returns:
				returned = true
			default:
			}
			if returned {
				return fmt.Errorf("message was unroutable for queue %s", queue)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("wait for publisher confirmation: %w", ctx.Err())
		}
	}
}
