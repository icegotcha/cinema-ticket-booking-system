package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(rg *gin.RouterGroup, db *mongo.Database) {
	repository := user.NewUserRepository(db)
	service := NewService(repository)
	handler := NewHandler(service)
	handler.RegisterRoutes(rg)
}
