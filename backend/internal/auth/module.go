package auth

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(rg *gin.RouterGroup, db *mongo.Database) error {
	repository := user.NewUserRepository(db)
	if err := repository.EnsureIndexes(context.Background()); err != nil {
		return fmt.Errorf("initialize user repository: %w", err)
	}
	service := NewService(repository)
	handler := NewHandler(service)
	handler.RegisterRoutes(rg)
	return nil
}
