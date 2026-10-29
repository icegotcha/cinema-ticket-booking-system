package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	booking_auditlog "github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking/auditlog"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/event"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const BookingSuccessQueue = "booking.success"

type BookingSuccessWorker struct {
	channel            *amqp.Channel
	auditLogRepository *booking_auditlog.BookingAuditLogRepository
}

func NewBookingSuccessWorker(
	channel *amqp.Channel,
	db *mongo.Database,
) *BookingSuccessWorker {
	return &BookingSuccessWorker{
		channel:            channel,
		auditLogRepository: booking_auditlog.NewBookingAuditLogRepository(db),
	}
}

func (w *BookingSuccessWorker) Run(ctx context.Context) error {
	_, err := w.channel.QueueDeclare(
		BookingSuccessQueue,
		true,  // durable
		false, // auto delete
		false, // exclusive
		false, // no wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare booking success queue: %w", err)
	}

	if err := w.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("set worker qos: %w", err)
	}

	messages, err := w.channel.Consume(
		BookingSuccessQueue,
		"booking_success_worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume booking success queue: %w", err)
	}

	log.Println("booking success worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("booking success worker stopped")
			return nil

		case message, ok := <-messages:
			if !ok {
				return fmt.Errorf("booking success message channel closed")
			}

			w.handleMessage(ctx, message)
		}
	}
}

func (w *BookingSuccessWorker) handleMessage(
	ctx context.Context,
	message amqp.Delivery,
) {
	var bookingEvent event.BookingSuccessEvent

	if err := json.Unmarshal(message.Body, &bookingEvent); err != nil {
		log.Printf("decode booking success event: %v", err)

		if err := message.Nack(false, false); err != nil {
			log.Printf("nack invalid message: %v", err)
		}

		return
	}

	auditLog := booking_auditlog.BookingAuditLog{
		Event:      "BOOKING_SUCCESS",
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
		"booking success audit recorded: booking_id=%s",
		bookingEvent.BookingID,
	)
}
