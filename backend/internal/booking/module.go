package booking

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking/locker"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/queue"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(
	rg *gin.RouterGroup,
	db *mongo.Database,
	channel *amqp.Channel,
	notifier BookingNotifier,
	redisClient *redis.Client,
) error {
	repository := NewBookingRepository(db)
	if err := repository.EnsureIndexes(context.Background()); err != nil {
		return fmt.Errorf("initialize booking repository: %w", err)
	}
	showtimeRepository := showtime.NewShowTimeRepository(db)
	publisher, err := queue.NewPublisher(channel)
	if err != nil {
		return fmt.Errorf("initialize queue publisher: %w", err)
	}
	seatLocker := locker.NewBookingSeatLocker(redisClient)
	service := NewBookingService(repository, showtimeRepository, publisher, notifier, seatLocker)
	handler := NewBookingHandler(service)
	handler.RegisterRoutes(rg)
	go expireSeatHolds(service)
	return nil
}

func expireSeatHolds(service BookingService) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := service.ExpireBookings(ctx)
		cancel()
		if err != nil {
			log.Printf("expire seat holds: %v", err)
		}
	}
}
