package booking

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/event"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *bookingService) ConfirmBookings(
	ctx context.Context,
	request BookingSeatsRequest,
) ([]BookingResponse, error) {
	bookings, err := s.findOwnedLocks(ctx, request)
	if err != nil {
		return nil, err
	}
	if err := s.validateLocks(ctx, bookings); err != nil {
		return nil, err
	}

	confirmedAt := time.Now().UTC()
	responses := make([]BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		confirmed, err := s.repository.ConfirmLock(ctx, booking.ID, confirmedAt)
		if err != nil {
			return nil, err
		}
		if !confirmed {
			return nil, ErrHoldExpired
		}

		lock := booking
		clearLock(&booking, BookingStatusBooked)
		response := bookingResponse(booking)
		responses = append(responses, response)
		s.notify(booking)
		s.publishBookingSuccess(ctx, response)
		if err := s.unlock(lock); err != nil {
			log.Printf("release confirmed seat lock: %v", err)
		}
	}
	return responses, nil
}

func (s *bookingService) ReleaseBookings(
	ctx context.Context,
	request BookingSeatsRequest,
) ([]BookingResponse, error) {
	bookings, err := s.findOwnedLocks(ctx, request)
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		released, err := s.repository.TransitionLock(
			ctx,
			booking.ID,
			BookingStatusLocked,
			BookingStatusCanceled,
		)
		if err != nil {
			return nil, err
		}
		if !released {
			continue
		}

		lock := booking
		clearLock(&booking, BookingStatusCanceled)
		responses = append(responses, bookingResponse(booking))
		s.notify(booking)
		if err := s.unlock(lock); err != nil {
			log.Printf("release canceled seat lock: %v", err)
		}
	}
	return responses, nil
}

func (s *bookingService) ExpireBookings(ctx context.Context) error {
	bookings, err := s.repository.FindExpiredLocks(ctx, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("find expired seat holds: %w", err)
	}

	for _, booking := range bookings {
		expired, err := s.repository.TransitionLock(
			ctx,
			booking.ID,
			BookingStatusLocked,
			BookingStatusExpired,
		)
		if err != nil {
			return err
		}
		if expired {
			clearLock(&booking, BookingStatusExpired)
			s.notify(booking)
		}
	}
	return nil
}

func (s *bookingService) findOwnedLocks(
	ctx context.Context,
	request BookingSeatsRequest,
) ([]Booking, error) {
	if _, err := bson.ObjectIDFromHex(request.ShowTimeID); err != nil {
		return nil, fmt.Errorf("invalid showtime id: %w", err)
	}

	seats, err := normalizeSeatNumbers(request.SeatNumbers)
	if err != nil {
		return nil, err
	}
	bookings, err := s.repository.FindLockedBySeats(ctx, request.ShowTimeID, request.UserID, seats)
	if err != nil {
		return nil, fmt.Errorf("find seat holds: %w", err)
	}
	if len(bookings) != len(seats) {
		return nil, ErrHoldNotFound
	}
	return bookings, nil
}

func (s *bookingService) validateLocks(ctx context.Context, bookings []Booking) error {
	now := time.Now().UTC()
	for _, booking := range bookings {
		if booking.ExpiresAt == nil || !booking.ExpiresAt.After(now) {
			return ErrHoldExpired
		}
		owned, err := s.seatLocker.IsOwner(
			ctx,
			booking.ShowTimeID,
			[]string{booking.SeatNumber},
			booking.LockToken,
		)
		if err != nil {
			return err
		}
		if !owned {
			return ErrHoldExpired
		}
	}
	return nil
}

func (s *bookingService) publishBookingSuccess(ctx context.Context, booking BookingResponse) {
	payload := event.BookingSuccessEvent{
		BookingID:  booking.ID,
		ShowTimeID: booking.ShowTimeID,
		UserID:     booking.UserID,
		SeatNumber: booking.SeatNumber,
		Status:     string(booking.Status),
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.queuePublisher.Publish(ctx, "booking.success", payload); err != nil {
		log.Printf("publish booking success: %v", err)
	}
}
