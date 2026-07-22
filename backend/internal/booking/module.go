package booking

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(rg *gin.RouterGroup, db *mongo.Database) {
	repository := NewBookingRepository(db)
	service := NewBookingService(repository)
	handler := NewBookingHandler(service)
	handler.RegisterRoutes(rg)
}
