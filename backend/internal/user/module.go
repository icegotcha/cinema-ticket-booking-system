package user

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(rg *gin.RouterGroup, db *mongo.Database) {
	handler := NewHandler()
	handler.RegisterRoutes(rg)

}
