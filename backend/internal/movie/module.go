package movie

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(rg *gin.RouterGroup, db *mongo.Database) {
	repository := NewMovieRepository(db)
	service := NewMovieService(repository)
	handler := NewMovieHandler(service)
	handler.RegisterRoutes(rg)
}
