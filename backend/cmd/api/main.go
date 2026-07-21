package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/auth"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

	db, err := database.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	router := gin.Default()
	api := router.Group("/api")
	auth.Register(api, db)
	user.Register(api, db)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
