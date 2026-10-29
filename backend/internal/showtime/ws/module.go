package websocket_showtime

import (
	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(rg *gin.RouterGroup, db *mongo.Database, showtimeHub *Hub) {
	bookingRepository := booking.NewBookingRepository(db)
	showtimeWebsocketHandler := NewHandler(bookingRepository, showtimeHub)
	rg.GET("/showtimes/:showtime_id", showtimeWebsocketHandler.HandleConnection)
}
