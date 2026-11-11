package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/auth"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/notifier"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/queue"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/redis"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime"
	websocket_showtime "github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime/ws"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
)

func main() {
	db, err := database.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	redisClient, err := redis.NewRedisClient()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	channel, err := queue.NewRabbitMQClient()
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer channel.Close()

	router := gin.Default()

	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
	router.GET("/health", healthHandler)

	ws := router.Group("/ws")
	showtimeHub := websocket_showtime.NewHub()
	websocket_showtime.Register(ws, db, showtimeHub)

	notifierService := notifier.NewNotifierService(redisClient)
	subscription, err := notifierService.SubscribeExpirationUpdates(context.Background(), showtimeHub)
	if err != nil {
		log.Fatalf("initialize expiration updates: %v", err)
	}
	defer subscription.Close()

	api := router.Group("/api")
	api.GET("/health", healthHandler)
	auth.Register(api, db)
	user.Register(api, db)
	movie.Register(api, db)
	showtime.Register(api, db)
	if err := booking.Register(api, db, channel.Channel, showtimeHub, redisClient); err != nil {
		log.Fatalf("failed to register booking module: %v", err)
	}

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
