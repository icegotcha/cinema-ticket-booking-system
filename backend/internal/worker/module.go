package worker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/locker"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/notifier"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/queue"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/redis"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupExpirationWorker(db *mongo.Database) (*BookingExpirationWorker, func(), error) {
	interval, err := readExpirationInterval()
	if err != nil {
		return nil, nil, err
	}

	publisherChannel, err := queue.NewRabbitMQClient()
	if err != nil {
		return nil, nil, fmt.Errorf("create publisher channel: %w", err)
	}
	publisher, err := queue.NewPublisher(publisherChannel.Channel)
	if err != nil {
		_ = publisherChannel.Close()
		return nil, nil, fmt.Errorf("create publisher: %w", err)
	}

	redisClient, err := redis.NewRedisClient()
	if err != nil {
		_ = publisherChannel.Close()
		return nil, nil, fmt.Errorf("connect to Redis: %w", err)
	}
	cleanup := func() {
		_ = redisClient.Close()
		_ = publisherChannel.Close()
	}

	repository := booking.NewBookingRepository(db)
	if err := repository.EnsureIndexes(context.Background()); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("initialize booking indexes: %w", err)
	}
	showtimeRepository := showtime.NewShowTimeRepository(db)
	notifierService := notifier.NewNotifierService(redisClient)
	seatLocker := locker.NewBookingSeatLocker(redisClient)
	service := booking.NewBookingService(
		repository,
		showtimeRepository,
		publisher,
		notifierService,
		seatLocker,
	)
	return NewBookingExpirationWorker(service, interval), cleanup, nil
}

func readExpirationInterval() (time.Duration, error) {
	value := os.Getenv("BOOKING_EXPIRATION_INTERVAL")
	if value == "" {
		return time.Second, nil
	}
	interval, err := time.ParseDuration(value)
	if err != nil || interval <= 0 {
		return 0, fmt.Errorf("BOOKING_EXPIRATION_INTERVAL must be a positive duration")
	}
	return interval, nil
}
