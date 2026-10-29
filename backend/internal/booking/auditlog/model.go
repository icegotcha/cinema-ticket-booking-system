package booking_auditlog

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingAuditLog struct {
	ID         bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Event      string        `json:"event" bson:"event"`
	BookingID  string        `json:"booking_id" bson:"booking_id"`
	UserID     string        `json:"user_id" bson:"user_id"`
	ShowTimeID string        `json:"show_time_id" bson:"show_time_id"`
	SeatNumber string        `json:"seat_number" bson:"seat_number"`
	CreatedAt  time.Time     `json:"created_at" bson:"created_at"`
}
