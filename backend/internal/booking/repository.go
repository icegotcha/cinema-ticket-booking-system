package booking

import (
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BookingRepository struct {
	*database.Repository[Booking]
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		Repository: database.NewRepository[Booking](db, "bookings"),
	}
}
