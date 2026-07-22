package booking

import "go.mongodb.org/mongo-driver/v2/bson"

type Booking struct {
	ID         bson.ObjectID `json:"id" bson:"_id,omitempty"`
	ShowTimeId string        `json:"show_time_id" bson:"show_time_id,omitempty"`
	UserId     string        `json:"user_id" bson:"user_id,omitempty"`
	SeatNumber string        `json:"seat_number" bson:"seat_number,omitempty"`
}

type CreateBookingsRequest struct {
	ShowTimeId  string   `json:"show_time_id" binding:"required"`
	UserId      string   `json:"user_id"`
	SeatNumbers []string `json:"seat_numbers" binding:"required,min=1"`
}

type BookingResponse struct {
	ID         string `json:"id"`
	ShowTimeId string `json:"show_time_id"`
	UserId     string `json:"user_id"`
	SeatNumber string `json:"seat_number"`
}
