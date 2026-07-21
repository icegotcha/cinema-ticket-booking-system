package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/auth"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
)

func main() {
	db, err := database.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	router := gin.Default()

	api := router.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	auth.Register(api, db)
	user.Register(api, db)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
