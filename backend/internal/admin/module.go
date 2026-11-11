package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(router *gin.RouterGroup, db *mongo.Database) {
	repository := NewAdminRepository(db)
	service := NewAdminService(
		repository,
		booking.NewBookingRepository(db),
		showtime.NewShowTimeRepository(db),
		movie.NewMovieRepository(db),
	)
	handler := NewAdminHandler(service)
	handler.RegisterRoutes(router)
}
