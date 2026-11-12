package event

import "time"

const (
	BookingAuditQueue = "booking.audit"
	BookingSuccess    = "BOOKING_SUCCESS"
	BookingTimeout    = "BOOKING_TIMEOUT"
	SeatReleased      = "SEAT_RELEASED"
	SystemError       = "SYSTEM_ERROR"
)

type BookingAuditEvent struct {
	Event      string    `json:"event"`
	BookingID  string    `json:"booking_id"`
	UserID     string    `json:"user_id"`
	ShowTimeID string    `json:"show_time_id"`
	SeatNumber string    `json:"seat_number"`
	Message    string    `json:"message,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
