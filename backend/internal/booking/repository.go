package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BookingRepository struct {
	*database.Repository[Booking]
}

func (r *BookingRepository) FindByShowtime(ctx context.Context, showtimeID string) ([]Booking, error) {
	return r.FindMany(ctx, bson.M{"show_time_id": showtimeID})
}

func (r *BookingRepository) FindActiveByShowtime(ctx context.Context, showtimeID string) ([]Booking, error) {
	return r.FindMany(ctx, bson.M{
		"show_time_id": showtimeID,
		"status":       bson.M{"$in": activeBookingStatuses()},
	})
}

func (r *BookingRepository) FindLockedBySeats(
	ctx context.Context,
	showtimeID string,
	userID string,
	seatNumbers []string,
) ([]Booking, error) {
	return r.FindMany(ctx, bson.M{
		"show_time_id": showtimeID,
		"user_id":      userID,
		"seat_number":  bson.M{"$in": seatNumbers},
		"status":       BookingStatusLocked,
	})
}

func (r *BookingRepository) FindExpiredLocks(ctx context.Context, now time.Time) ([]Booking, error) {
	return r.FindMany(ctx, bson.M{
		"status":     BookingStatusLocked,
		"expires_at": bson.M{"$lte": now},
	})
}

func (r *BookingRepository) TransitionLock(
	ctx context.Context,
	bookingID bson.ObjectID,
	from BookingStatus,
	to BookingStatus,
) (bool, error) {
	result, err := r.Collection().UpdateOne(
		ctx,
		bson.M{"_id": bookingID, "status": from},
		bson.M{
			"$set": bson.M{"status": to},
			"$unset": bson.M{
				"lock_token": "",
				"expires_at": "",
			},
		},
	)
	if err != nil {
		return false, fmt.Errorf("transition booking status: %w", err)
	}
	return result.ModifiedCount == 1, nil
}

func (r *BookingRepository) ConfirmLock(
	ctx context.Context,
	bookingID bson.ObjectID,
	confirmedAt time.Time,
) (bool, error) {
	result, err := r.Collection().UpdateOne(
		ctx,
		bson.M{
			"_id":        bookingID,
			"status":     BookingStatusLocked,
			"expires_at": bson.M{"$gt": confirmedAt},
		},
		bson.M{
			"$set": bson.M{"status": BookingStatusBooked},
			"$unset": bson.M{
				"lock_token": "",
				"expires_at": "",
			},
		},
	)
	if err != nil {
		return false, fmt.Errorf("confirm booking: %w", err)
	}
	return result.ModifiedCount == 1, nil
}

func (r *BookingRepository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.Collection().UpdateMany(
		ctx,
		bson.M{"status": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"status": BookingStatusBooked}},
	); err != nil {
		return fmt.Errorf("migrate legacy booking statuses: %w", err)
	}

	_, err := r.Collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "show_time_id", Value: 1},
			{Key: "seat_number", Value: 1},
		},
		Options: options.Index().
			SetName("unique_active_showtime_seat").
			SetUnique(true).
			SetPartialFilterExpression(bson.M{
				"status": bson.M{"$in": activeBookingStatuses()},
			}),
	})
	if err != nil {
		return fmt.Errorf("create booking indexes: %w", err)
	}
	_, err = r.Collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "expires_at", Value: 1},
		},
		Options: options.Index().SetName("booking_hold_expiry"),
	})
	if err != nil {
		return fmt.Errorf("create booking hold expiry index: %w", err)
	}
	return nil
}

func (r *BookingRepository) CreateMany(ctx context.Context, bookings []Booking) ([]bson.ObjectID, error) {
	ids := make([]bson.ObjectID, len(bookings))
	documents := make([]any, len(bookings))
	for i := range bookings {
		ids[i] = bson.NewObjectID()
		bookings[i].ID = ids[i]
		documents[i] = bookings[i]
	}

	_, err := r.Collection().InsertMany(ctx, documents)
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, cleanupErr := r.Collection().DeleteMany(
			cleanupCtx,
			bson.M{"_id": bson.M{"$in": ids}},
		); cleanupErr != nil {
			return nil, fmt.Errorf("insert bookings: %w (compensation failed: %v)", err, cleanupErr)
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrSeatUnavailable
		}
		return nil, fmt.Errorf("insert bookings: %w", err)
	}

	return ids, nil
}

func (r *BookingRepository) ReplaceByID(ctx context.Context, id bson.ObjectID, booking *Booking) error {
	err := r.Repository.ReplaceByID(ctx, id, booking)
	if mongo.IsDuplicateKeyError(err) {
		return ErrSeatUnavailable
	}
	return err
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		Repository: database.NewRepository[Booking](db, "bookings"),
	}
}
