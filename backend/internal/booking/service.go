package booking

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking/locker"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/queue"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrSeatUnavailable = errors.New("one or more seats are unavailable")
	ErrInvalidSeat     = errors.New("one or more seat numbers are invalid")
	ErrHoldNotFound    = errors.New("active seat hold not found")
	ErrHoldExpired     = errors.New("seat hold has expired")
)

var seatNumberPattern = regexp.MustCompile(`^[A-H](?:[1-9]|10)$`)

type BookingService interface {
	GetBookings(context.Context, string) ([]BookingResponse, error)
	HoldBookings(context.Context, HoldBookingsRequest) ([]BookingResponse, error)
	ConfirmBookings(context.Context, BookingSeatsRequest) ([]BookingResponse, error)
	ReleaseBookings(context.Context, BookingSeatsRequest) ([]BookingResponse, error)
	ExpireBookings(context.Context) error
}

type BookingNotifier interface {
	NotifyBookingUpdated(BookingResponse)
}

type bookingService struct {
	repository         *BookingRepository
	showtimeRepository *showtime.ShowTimeRepository
	queuePublisher     *queue.Publisher
	notifier           BookingNotifier
	seatLocker         locker.SeatLocker
}

func NewBookingService(
	repository *BookingRepository,
	showtimeRepository *showtime.ShowTimeRepository,
	publisher *queue.Publisher,
	notifier BookingNotifier,
	seatLocker locker.SeatLocker,
) BookingService {
	return &bookingService{
		repository:         repository,
		showtimeRepository: showtimeRepository,
		queuePublisher:     publisher,
		notifier:           notifier,
		seatLocker:         seatLocker,
	}
}

func (s *bookingService) GetBookings(ctx context.Context, showtimeID string) ([]BookingResponse, error) {
	bookings, err := s.findBookings(ctx, showtimeID)
	if err != nil {
		return nil, err
	}

	responses := make([]BookingResponse, len(bookings))
	for index, booking := range bookings {
		responses[index] = bookingResponse(booking)
	}
	return responses, nil
}

func (s *bookingService) findBookings(ctx context.Context, showtimeID string) ([]Booking, error) {
	if showtimeID == "" {
		return s.repository.FindAll(ctx)
	}
	return s.repository.FindActiveByShowtime(ctx, showtimeID)
}

func activeBookingStatuses() []BookingStatus {
	return []BookingStatus{BookingStatusPending, BookingStatusLocked, BookingStatusBooked}
}

func normalizeSeatNumbers(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, ErrInvalidSeat
	}

	unique := make(map[string]struct{}, len(raw))
	seats := make([]string, 0, len(raw))
	for _, value := range raw {
		seat := strings.ToUpper(strings.TrimSpace(value))
		if !seatNumberPattern.MatchString(seat) {
			return nil, ErrInvalidSeat
		}
		if _, exists := unique[seat]; exists {
			return nil, ErrInvalidSeat
		}
		unique[seat] = struct{}{}
		seats = append(seats, seat)
	}
	sort.Strings(seats)
	return seats, nil
}

func bookingResponse(booking Booking) BookingResponse {
	return BookingResponse{
		ID:         objectIDString(booking.ID),
		ShowTimeID: booking.ShowTimeID,
		UserID:     booking.UserID,
		SeatNumber: booking.SeatNumber,
		Status:     booking.Status,
		ExpiresAt:  booking.ExpiresAt,
	}
}

func objectIDString(id bson.ObjectID) string {
	if id == bson.NilObjectID {
		return ""
	}
	return id.Hex()
}

func (s *bookingService) notify(booking Booking) {
	if s.notifier != nil {
		s.notifier.NotifyBookingUpdated(bookingResponse(booking))
	}
}

func clearLock(booking *Booking, status BookingStatus) {
	booking.Status = status
	booking.LockToken = ""
	booking.ExpiresAt = nil
}

func (s *bookingService) unlock(booking Booking) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.seatLocker.Unlock(ctx, booking.ShowTimeID, []string{booking.SeatNumber}, booking.LockToken)
}
