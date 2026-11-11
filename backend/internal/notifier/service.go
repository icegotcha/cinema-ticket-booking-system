package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
	"github.com/redis/go-redis/v9"
)

const expirationUpdatesChannel = "booking:expiration_updates"

type NotifierService interface {
	NotifyBookingUpdated(booking.BookingResponse)
	SubscribeExpirationUpdates(context.Context, booking.BookingNotifier) (*redis.PubSub, error)
}

type notifierService struct {
	client *redis.Client
}

func NewNotifierService(client *redis.Client) NotifierService {
	return &notifierService{client: client}
}

func (s *notifierService) NotifyBookingUpdated(updated booking.BookingResponse) {
	payload, err := json.Marshal(updated)
	if err != nil {
		log.Printf("encode expiration update: %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.client.Publish(ctx, expirationUpdatesChannel, payload).Err(); err != nil {
		log.Printf("publish expiration update: %v", err)
	}
}

func (s *notifierService) SubscribeExpirationUpdates(ctx context.Context, listener booking.BookingNotifier) (*redis.PubSub, error) {
	subscription := s.client.Subscribe(ctx, expirationUpdatesChannel)
	if _, err := subscription.Receive(ctx); err != nil {
		_ = subscription.Close()
		return nil, fmt.Errorf("subscribe expiration updates: %w", err)
	}
	go func() {
		defer subscription.Close()
		messages := subscription.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-messages:
				if !ok {
					return
				}
				var updated booking.BookingResponse
				if err := json.Unmarshal([]byte(message.Payload), &updated); err != nil {
					log.Printf("decode expiration update: %v", err)
					continue
				}
				if listener != nil {
					listener.NotifyBookingUpdated(updated)
				}
			}
		}
	}()
	return subscription, nil
}
