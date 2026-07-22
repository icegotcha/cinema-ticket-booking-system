package booking

type Booking struct {
	ID         string `json:"id" bson:"_id,omitempty"`
	ShowTimeId string `json:"show_time_id" bson:"show_time_id,omitempty"`
	UserId     string `json:"user_id" bson:"user_id,omitempty"`
	SeatNumber string `json:"seat_number" bson:"seat_number,omitempty"`
}

type BookingResponse struct {
	ID         string `json:"id"`
	ShowTimeId string `json:"show_time_id"`
	UserId     string `json:"user_id"`
	SeatNumber string `json:"seat_number"`
}
