package websocket_showtime

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Handler struct {
	bookingRepository *booking.BookingRepository
	hub               *Hub
}

type client struct {
	connection *websocket.Conn
	writeMu    sync.Mutex
}

func (c *client) writeJSON(message ShowtimeWebsocketMessage) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.connection.WriteJSON(message)
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*client]struct{})}
}

func NewHandler(repository *booking.BookingRepository, hub *Hub) *Handler {
	return &Handler{bookingRepository: repository, hub: hub}
}

func (h *Hub) add(showtimeID string, connectedClient *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[showtimeID] == nil {
		h.clients[showtimeID] = make(map[*client]struct{})
	}
	h.clients[showtimeID][connectedClient] = struct{}{}
}

func (h *Hub) remove(showtimeID string, connectedClient *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients[showtimeID], connectedClient)
	if len(h.clients[showtimeID]) == 0 {
		delete(h.clients, showtimeID)
	}
}

func (h *Hub) NotifyBookingUpdated(updated booking.BookingResponse) {
	seatStatus := string(updated.Status)
	if updated.Status != booking.BookingStatusLocked && updated.Status != booking.BookingStatusBooked {
		seatStatus = "AVAILABLE"
	}
	message := ShowtimeWebsocketMessage{
		Type: MessageTypeSeatUpdated,
		Bookings: []ShowtimeBooking{{
			ID:         updated.ID,
			SeatNumber: updated.SeatNumber,
			SeatStatus: seatStatus,
		}},
	}

	h.mu.RLock()
	clients := make([]*client, 0, len(h.clients[updated.ShowTimeID]))
	for client := range h.clients[updated.ShowTimeID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.writeJSON(message); err != nil {
			h.remove(updated.ShowTimeID, client)
			_ = client.connection.Close()
		}
	}
}

func (h *Handler) HandleConnection(c *gin.Context) {
	showtimeID := c.Param("showtime_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	client := &client{connection: conn}
	client.writeMu.Lock()
	h.hub.add(showtimeID, client)
	defer h.hub.remove(showtimeID, client)

	bookings, err := h.bookingRepository.FindActiveByShowtime(
		c.Request.Context(),
		showtimeID,
	)
	if err != nil {
		_ = conn.WriteJSON(ShowtimeWebsocketMessage{
			Type:     MessageTypeError,
			Message:  err.Error(),
			Bookings: []ShowtimeBooking{},
		})
		client.writeMu.Unlock()
		return
	}

	messageBookings := make([]ShowtimeBooking, len(bookings))
	for i := range bookings {
		messageBookings[i] = ShowtimeBooking{
			ID:         bookings[i].ID.Hex(),
			SeatNumber: bookings[i].SeatNumber,
			SeatStatus: string(bookings[i].Status),
		}
	}

	err = conn.WriteJSON(ShowtimeWebsocketMessage{
		Type:     MessageTypeSeatSnapshot,
		Bookings: messageBookings,
	})
	client.writeMu.Unlock()
	if err != nil {
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
