package booking

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrSeatUnavailable = errors.New("one or more seats are unavailable")

type BookingService interface {
	GetBookings(ctx context.Context, showtimeID string) ([]BookingResponse, error)
	CreateBookings(ctx context.Context, request CreateBookingsRequest) (*[]BookingResponse, error)
}

type bookingService struct {
	repository *BookingRepository
}

func NewBookingService(repository *BookingRepository) BookingService {
	return &bookingService{
		repository: repository,
	}
}

func (s *bookingService) GetBookings(ctx context.Context, showtimeID string) ([]BookingResponse, error) {
	var bookings []Booking
	var err error
	if showtimeID == "" {
		bookings, err = s.repository.FindAll(ctx)
	} else {
		bookings, err = s.repository.FindByShowtime(ctx, showtimeID)
	}
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, 0, len(bookings))

	for _, b := range bookings {
		responses = append(responses, BookingResponse{
			ID:         objectIDString(b.ID),
			ShowTimeId: b.ShowTimeId,
			UserId:     b.UserId,
			SeatNumber: b.SeatNumber,
		})
	}

	return responses, nil
}

func (s *bookingService) CreateBookings(
	ctx context.Context,
	request CreateBookingsRequest,
) (*[]BookingResponse, error) {
	return nil, nil
}

func objectIDString(id bson.ObjectID) string {
	if id == bson.NilObjectID {
		return ""
	}
	return id.Hex()
}
