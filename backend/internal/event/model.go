package event

import "time"

type BookingSuccessEvent struct {
	BookingID  string    `json:"booking_id"`
	UserID     string    `json:"user_id"`
	ShowTimeID string    `json:"show_time_id"`
	SeatNumber string    `json:"seat_number"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
