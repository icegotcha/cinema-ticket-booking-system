package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	booking_auditlog "github.com/icegotcha/cinema-ticket-booking-system/backend/internal/auditlog"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/event"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type auditLogWriter interface {
	Create(context.Context, *booking_auditlog.BookingAuditLog) (bson.ObjectID, error)
}

type BookingAuditWorker struct {
	channel            *amqp.Channel
	auditLogRepository auditLogWriter
}

func NewBookingAuditWorker(
	channel *amqp.Channel,
	db *mongo.Database,
) *BookingAuditWorker {
	return &BookingAuditWorker{
		channel:            channel,
		auditLogRepository: booking_auditlog.NewBookingAuditLogRepository(db),
	}
}

func (w *BookingAuditWorker) Run(ctx context.Context) error {
	consume := func(queueName string) (<-chan amqp.Delivery, error) {
		if _, err := w.channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
			return nil, fmt.Errorf("declare %s: %w", queueName, err)
		}
		return w.channel.Consume(queueName, "", false, false, false, false, nil)
	}
	if err := w.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("set worker qos: %w", err)
	}
	messages, err := consume(event.BookingAuditQueue)
	if err != nil {
		return fmt.Errorf("consume booking audit queue: %w", err)
	}

	log.Println("booking audit worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("booking audit worker stopped")
			return nil

		case message, ok := <-messages:
			if !ok {
				return fmt.Errorf("booking audit message channel closed")
			}
			w.handleMessage(ctx, message)
		}
	}
}

func (w *BookingAuditWorker) handleMessage(
	ctx context.Context,
	message amqp.Delivery,
) {
	var bookingEvent event.BookingAuditEvent

	if err := json.Unmarshal(message.Body, &bookingEvent); err != nil {
		log.Printf("decode booking audit event: %v", err)

		if err := message.Nack(false, false); err != nil {
			log.Printf("nack invalid message: %v", err)
		}

		return
	}
	switch bookingEvent.Event {
	case event.BookingSuccess, event.BookingTimeout, event.SeatReleased, event.SystemError:
	default:
		log.Printf("reject unknown booking audit event: %q", bookingEvent.Event)
		if err := message.Nack(false, false); err != nil {
			log.Printf("nack invalid event: %v", err)
		}
		return
	}

	auditLog := booking_auditlog.BookingAuditLog{
		Event:      bookingEvent.Event,
		Message:    bookingEvent.Message,
		BookingID:  bookingEvent.BookingID,
		UserID:     bookingEvent.UserID,
		ShowTimeID: bookingEvent.ShowTimeID,
		SeatNumber: bookingEvent.SeatNumber,
		CreatedAt:  bookingEvent.CreatedAt,
	}

	if _, err := w.auditLogRepository.Create(ctx, &auditLog); err != nil {
		log.Printf("insert audit log: %v", err)

		if err := message.Nack(false, true); err != nil {
			log.Printf("nack message: %v", err)
		}

		return
	}

	if err := message.Ack(false); err != nil {
		log.Printf("ack message: %v", err)
		return
	}

	log.Printf(
		"booking audit recorded: event=%s booking_id=%s",
		bookingEvent.Event, bookingEvent.BookingID,
	)
}
