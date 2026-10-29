package booking

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingStatus string

const (
	BookingStatusPending  BookingStatus = "PENDING"
	BookingStatusLocked   BookingStatus = "LOCKED"
	BookingStatusBooked   BookingStatus = "BOOKED"
	BookingStatusExpired  BookingStatus = "EXPIRED"
	BookingStatusCanceled BookingStatus = "CANCELED"
)

type Booking struct {
	ID         bson.ObjectID `json:"id" bson:"_id,omitempty"`
	ShowTimeID string        `json:"show_time_id" bson:"show_time_id,omitempty"`
	UserID     string        `json:"user_id" bson:"user_id,omitempty"`
	SeatNumber string        `json:"seat_number" bson:"seat_number,omitempty"`
	Status     BookingStatus `json:"status" bson:"status"`
	LockToken  string        `json:"-" bson:"lock_token,omitempty"`
	ExpiresAt  *time.Time    `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
}

type HoldBookingsRequest struct {
	ShowTimeID  string   `json:"show_time_id" binding:"required"`
	UserID      string   `json:"user_id"`
	SeatNumbers []string `json:"seat_numbers" binding:"required,min=1"`
}

type BookingSeatsRequest struct {
	ShowTimeID  string   `json:"show_time_id" binding:"required"`
	UserID      string   `json:"user_id"`
	SeatNumbers []string `json:"seat_numbers" binding:"required,min=1"`
}

type BookingResponse struct {
	ID         string        `json:"id"`
	ShowTimeID string        `json:"show_time_id"`
	UserID     string        `json:"user_id"`
	SeatNumber string        `json:"seat_number"`
	Status     BookingStatus `json:"status"`
	ExpiresAt  *time.Time    `json:"expires_at,omitempty"`
}
