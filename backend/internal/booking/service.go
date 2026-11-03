package booking

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/event"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/locker"
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

	auditBooking := BookingResponse{ShowTimeID: request.ShowTimeID, UserID: request.UserID}
	token, locked, err := s.seatLocker.Lock(ctx, request.ShowTimeID, seats)
	if err != nil {
		s.publishAudit(event.SystemError, auditBooking, fmt.Sprintf("lock seats %v: %v", seats, err))
		return nil, fmt.Errorf("lock seats: %w", err)
	}
	if !locked {
		s.publishAudit(event.SystemError, auditBooking, fmt.Sprintf("lock seats %v: seats unavailable", seats))
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
			s.publishAudit(event.SystemError, auditBooking, fmt.Sprintf("release failed seat hold: %v", err))
		}
	}()

	if err := s.ensureSeatsAvailable(ctx, request.ShowTimeID, seats); err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(locker.SeatLockTTL)
	bookings := newLockedBookings(request, seats, token, expiresAt)
	ids, err := s.repository.CreateMany(ctx, bookings)
	if err != nil {
		s.publishAudit(event.SystemError, auditBooking, fmt.Sprintf("create seat holds: %v", err))
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
	unavailable, err := s.repository.HasActiveSeats(ctx, showtimeID, seats)
	if err != nil {
		return fmt.Errorf("check booked seats: %w", err)
	}
	if unavailable {
		return ErrSeatUnavailable
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
			s.publishAudit(event.SystemError, bookingResponse(booking), err.Error())
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
		s.publishAudit(event.BookingSuccess, response, "")
		if err := s.unlock(lock); err != nil {
			s.publishAudit(event.SystemError, response, fmt.Sprintf("release confirmed seat lock: %v", err))
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
			s.publishAudit(event.SystemError, bookingResponse(booking), err.Error())
			return nil, err
		}
		if !released {
			continue
		}

		lock := booking
		clearLock(&booking, BookingStatusCanceled)
		responses = append(responses, bookingResponse(booking))
		s.notify(booking)
		s.publishAudit(event.SeatReleased, bookingResponse(booking), "user released seat hold")
		if err := s.unlock(lock); err != nil {
			s.publishAudit(event.SystemError, bookingResponse(booking), fmt.Sprintf("release canceled seat lock: %v", err))
		}
	}
	return responses, nil
}

func (s *bookingService) ExpireBookings(ctx context.Context) error {
	bookings, err := s.repository.FindExpiredLocks(ctx, time.Now().UTC())
	if err != nil {
		s.publishAudit(event.SystemError, BookingResponse{}, fmt.Sprintf("find expired seat holds: %v", err))
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
			s.publishAudit(event.SystemError, bookingResponse(booking), err.Error())
			return err
		}
		if expired {
			clearLock(&booking, BookingStatusExpired)
			s.notify(booking)
			s.publishAudit(event.BookingTimeout, bookingResponse(booking), "seat hold expired")
			s.publishAudit(event.SeatReleased, bookingResponse(booking), "seat released after timeout")
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
	type lockGroup struct {
		showtimeID string
		token      string
	}
	now := time.Now().UTC()
	type heldSeats struct {
		booking Booking
		seats   []string
	}
	groups := make(map[lockGroup]heldSeats)
	for _, booking := range bookings {
		if booking.ExpiresAt == nil || !booking.ExpiresAt.After(now) {
			return ErrHoldExpired
		}
		group := lockGroup{showtimeID: booking.ShowTimeID, token: booking.LockToken}
		held := groups[group]
		held.booking = booking
		held.seats = append(held.seats, booking.SeatNumber)
		groups[group] = held
	}

	for group, held := range groups {
		owned, err := s.seatLocker.IsOwner(ctx, group.showtimeID, held.seats, group.token)
		if err != nil {
			s.publishAudit(event.SystemError, bookingResponse(held.booking), err.Error())
			return err
		}
		if !owned {
			return ErrHoldExpired
		}
	}
	return nil
}

// Use a bounded independent context so a disconnected HTTP client does not
// cancel the audit of an operation that has already changed booking state.
func (s *bookingService) publishAudit(kind string, booking BookingResponse, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	payload := event.BookingAuditEvent{
		Event: kind, BookingID: booking.ID, ShowTimeID: booking.ShowTimeID,
		UserID: booking.UserID, SeatNumber: booking.SeatNumber,
		Message: message, CreatedAt: time.Now().UTC(),
	}
	if err := s.queuePublisher.Publish(ctx, event.BookingAuditQueue, payload); err != nil {
		// Do not recursively publish failures of the audit transport itself.
		log.Printf("publish booking audit event=%s booking_id=%s: %v", kind, booking.ID, err)
	}
}
