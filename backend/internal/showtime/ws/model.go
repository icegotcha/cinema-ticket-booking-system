package websocket_showtime

type MessageType string

const (
	MessageTypeSeatSnapshot MessageType = "SEAT_SNAPSHOT"
	MessageTypeSeatUpdated  MessageType = "SEAT_UPDATED"
	MessageTypeError        MessageType = "ERROR"
)

type ShowtimeBooking struct {
	ID         string `json:"id"`
	SeatNumber string `json:"seat_number"`
	SeatStatus string `json:"seat_status"`
}

type ShowtimeWebsocketMessage struct {
	Type     MessageType       `json:"type"`
	Message  string            `json:"message,omitempty"`
	Bookings []ShowtimeBooking `json:"bookings"`
}
