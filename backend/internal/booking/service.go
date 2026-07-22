package booking

import (
	"context"
)

type BookingService interface {
	GetBookings(ctx context.Context) ([]BookingResponse, error)
}

type bookingService struct {
	repository *BookingRepository
}

func NewBookingService(repository *BookingRepository) BookingService {
	return &bookingService{
		repository: repository,
	}
}

func (s *bookingService) GetBookings(ctx context.Context) ([]BookingResponse, error) {
	bookings, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, 0, len(bookings))

	for _, b := range bookings {
		responses = append(responses, BookingResponse{
			ID:         b.ID,
			ShowTimeId: b.ShowTimeId,
			UserId:     b.UserId,
			SeatNumber: b.SeatNumber,
		})
	}

	return responses, nil
}
