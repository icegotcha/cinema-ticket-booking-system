package booking

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking/locker"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *bookingService) HoldBookings(
	ctx context.Context,
	request HoldBookingsRequest,
) ([]BookingResponse, error) {
	seats, err := normalizeSeatNumbers(request.SeatNumbers)
	if err != nil {
		return nil, err
	}
	if err := s.validateShowtime(ctx, request.ShowTimeID); err != nil {
		return nil, err
	}

	token, locked, err := s.seatLocker.Lock(ctx, request.ShowTimeID, seats)
	if err != nil {
		return nil, fmt.Errorf("lock seats: %w", err)
	}
	if !locked {
		return nil, ErrSeatUnavailable
	}

	persisted := false
	defer func() {
		if persisted {
			return
		}
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.seatLocker.Unlock(unlockCtx, request.ShowTimeID, seats, token); err != nil {
			log.Printf("release failed seat hold: %v", err)
		}
	}()

	if err := s.ensureSeatsAvailable(ctx, request.ShowTimeID, seats); err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(locker.SeatLockTTL)
	bookings := newLockedBookings(request, seats, token, expiresAt)
	ids, err := s.repository.CreateMany(ctx, bookings)
	if err != nil {
		return nil, fmt.Errorf("create seat holds: %w", err)
	}
	persisted = true

	responses := make([]BookingResponse, len(bookings))
	for index := range bookings {
		bookings[index].ID = ids[index]
		responses[index] = bookingResponse(bookings[index])
		s.notify(bookings[index])
	}
	return responses, nil
}

func (s *bookingService) validateShowtime(ctx context.Context, rawID string) error {
	id, err := bson.ObjectIDFromHex(rawID)
	if err != nil {
		return fmt.Errorf("invalid showtime id: %w", err)
	}

	showtime, err := s.showtimeRepository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find showtime: %w", err)
	}
	if showtime == nil {
		return errors.New("showtime not found")
	}
	if !time.Unix(int64(showtime.StartTime.T), 0).After(time.Now()) {
		return errors.New("showtime has already started")
	}
	return nil
}

func (s *bookingService) ensureSeatsAvailable(
	ctx context.Context,
	showtimeID string,
	seats []string,
) error {
	bookings, err := s.repository.FindActiveByShowtime(ctx, showtimeID)
	if err != nil {
		return fmt.Errorf("check booked seats: %w", err)
	}

	requested := make(map[string]struct{}, len(seats))
	for _, seat := range seats {
		requested[seat] = struct{}{}
	}
	for _, booking := range bookings {
		if _, exists := requested[booking.SeatNumber]; exists {
			return ErrSeatUnavailable
		}
	}
	return nil
}

func newLockedBookings(
	request HoldBookingsRequest,
	seats []string,
	token string,
	expiresAt time.Time,
) []Booking {
	bookings := make([]Booking, len(seats))
	for index, seat := range seats {
		bookings[index] = Booking{
			ShowTimeID: request.ShowTimeID,
			UserID:     request.UserID,
			SeatNumber: seat,
			Status:     BookingStatusLocked,
			LockToken:  token,
			ExpiresAt:  &expiresAt,
		}
	}
	return bookings
}
