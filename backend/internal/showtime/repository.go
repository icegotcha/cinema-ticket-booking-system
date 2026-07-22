package showtime

import (
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ShowTimeRepository struct {
	*database.Repository[ShowTime]
}

func NewShowTimeRepository(db *mongo.Database) *ShowTimeRepository {
	return &ShowTimeRepository{
		Repository: database.NewRepository[ShowTime](db, "showtimes"),
	}
}
