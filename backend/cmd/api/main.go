package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/auth"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/websocket"
)

func main() {
	db, err := database.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/ws", websocket.HandleConnection)

	api := router.Group("/api")
	auth.Register(api, db)
	user.Register(api, db)
	movie.Register(api, db)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
