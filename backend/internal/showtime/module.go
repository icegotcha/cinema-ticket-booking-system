package showtime

import (
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(rg *gin.RouterGroup, db *mongo.Database) {
	repository := NewShowTimeRepository(db)
	movieRepository := movie.NewMovieRepository(db)
	service := NewShowTimeService(repository, movieRepository)
	handler := NewShowTimeHandler(service)
	handler.RegisterRoutes(rg)
}
