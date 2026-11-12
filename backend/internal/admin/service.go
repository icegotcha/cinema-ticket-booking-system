package admin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/booking"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/showtime"
	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 8 * time.Hour

var (
	ErrInvalidCredentials = errors.New("Incorrect username or password")
	ErrSessionExpired     = errors.New("Session expired. Please sign in again")
	ErrCreateSession      = errors.New("Unable to create session")
)

type AdminService interface {
	Login(context.Context, AdminLoginRequest) (*AdminLoginResponse, error)
	VerifyToken(context.Context, string) (string, error)
	Logout(context.Context, string) error
	GetBookings(context.Context, AdminBookingsQuery) (*AdminBookingsResponse, error)
}

type adminAuthRepository interface {
	FindAccount(context.Context, string) (*AdminAccount, error)
	CreateSession(context.Context, AdminSession) error
	FindSession(context.Context, string) (*AdminSession, error)
	DeleteSession(context.Context, string) error
}

type adminService struct {
	repository         adminAuthRepository
	bookingRepository  *booking.BookingRepository
	showtimeRepository *showtime.ShowTimeRepository
	movieRepository    *movie.MovieRepository
}

func NewAdminService(
	repository adminAuthRepository,
	bookingRepository *booking.BookingRepository,
	showtimeRepository *showtime.ShowTimeRepository,
	movieRepository *movie.MovieRepository,
) AdminService {
	return &adminService{
		repository:         repository,
		bookingRepository:  bookingRepository,
		showtimeRepository: showtimeRepository,
		movieRepository:    movieRepository,
	}
}

func (s *adminService) Login(ctx context.Context, request AdminLoginRequest) (*AdminLoginResponse, error) {
	account, err := s.repository.FindAccount(ctx, request.Username)
	if err != nil {
		return nil, fmt.Errorf("find admin account: %w", err)
	}
	if account == nil || bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(request.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateSession, err)
	}
	token := base64.RawURLEncoding.EncodeToString(secret)
	expiresAt := time.Now().Add(sessionDuration)
	if err := s.repository.CreateSession(ctx, AdminSession{
		TokenHash: hashToken(token), Username: account.Username, ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateSession, err)
	}
	return &AdminLoginResponse{AccessToken: token, TokenType: "Bearer", ExpiresAt: expiresAt}, nil
}

func (s *adminService) VerifyToken(ctx context.Context, token string) (string, error) {
	tokenHash := hashToken(token)
	session, err := s.repository.FindSession(ctx, tokenHash)
	if err != nil {
		return "", fmt.Errorf("find admin session: %w", err)
	}
	if session == nil || !time.Now().Before(session.ExpiresAt) {
		return "", ErrSessionExpired
	}
	return tokenHash, nil
}

func (s *adminService) Logout(ctx context.Context, tokenHash string) error {
	return s.repository.DeleteSession(ctx, tokenHash)
}

func (s *adminService) GetBookings(ctx context.Context, query AdminBookingsQuery) (*AdminBookingsResponse, error) {
	bookings, total, err := s.bookingRepository.FindAdminPage(ctx, booking.AdminPageQuery{
		Page: query.Page, PageSize: query.PageSize, MovieID: query.MovieID,
		UserID: query.UserID, DateFrom: query.DateFrom, DateTo: query.DateTo,
	})
	if err != nil {
		return nil, errors.New("Unable to load bookings")
	}
	showtimes, err := s.showtimeRepository.FindAll(ctx)
	if err != nil {
		return nil, errors.New("Unable to load showtimes")
	}
	movies, err := s.movieRepository.FindAllSummaries(ctx)
	if err != nil {
		return nil, errors.New("Unable to load movies")
	}
	return &AdminBookingsResponse{Bookings: bookings, Showtimes: showtimes, Movies: movies, Page: query.Page, PageSize: query.PageSize, Total: total}, nil
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
