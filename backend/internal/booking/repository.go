package booking

import (
	"context"
	"fmt"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BookingRepository struct {
	*database.Repository[Booking]
}

func (r *BookingRepository) FindByShowtime(ctx context.Context, showtimeID string) ([]Booking, error) {
	return r.FindMany(ctx, bson.M{"show_time_id": showtimeID})
}

func (r *BookingRepository) CreateMany(ctx context.Context, bookings []Booking) ([]bson.ObjectID, error) {
	documents := make([]any, len(bookings))
	for i := range bookings {
		documents[i] = bookings[i]
	}

	result, err := r.Collection().InsertMany(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("insert bookings: %w", err)
	}

	ids := make([]bson.ObjectID, 0, len(result.InsertedIDs))
	for _, insertedID := range result.InsertedIDs {
		id, ok := insertedID.(bson.ObjectID)
		if !ok {
			return nil, fmt.Errorf("inserted booking id has unexpected type")
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		Repository: database.NewRepository[Booking](db, "bookings"),
	}
}
