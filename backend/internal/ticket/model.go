package ticket

type Ticket struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	BookingId string `json:"booking_id" bson:"booking_id,omitempty"`
}
