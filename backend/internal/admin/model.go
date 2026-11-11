package admin

import (
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime"
)

type AdminAccount struct {
	Username     string `bson:"username"`
	PasswordHash string `bson:"password_hash"`
}

type AdminSession struct {
	TokenHash string    `bson:"token_hash"`
	Username  string    `bson:"username"`
	ExpiresAt time.Time `bson:"expires_at"`
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required,max=128"`
	Password string `json:"password" binding:"required,max=72"`
}

type AdminLoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type AdminBookingsResponse struct {
	Bookings  []booking.Booking   `json:"bookings"`
	Showtimes []showtime.ShowTime `json:"showtimes"`
	Movies    []movie.Movie       `json:"movies"`
}
