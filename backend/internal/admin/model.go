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

type AdminBookingsQuery struct {
	Page     int    `form:"page" binding:"min=1,max=1000000"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	MovieID  string `form:"movie_id"`
	UserID   string `form:"user_id"`
	DateFrom int64  `form:"date_from" binding:"min=0,max=4294967295"`
	DateTo   int64  `form:"date_to" binding:"min=0,max=4294967295"`
}

type AdminBookingsResponse struct {
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
	Total     int64               `json:"total"`
	Bookings  []booking.Booking   `json:"bookings"`
	Showtimes []showtime.ShowTime `json:"showtimes"`
	Movies    []movie.Movie       `json:"movies"`
}
