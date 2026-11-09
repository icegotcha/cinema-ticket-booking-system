package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/firebase"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid or expired Firebase ID token")
	ErrNetworkError       = errors.New("network error")
	ErrDatabase           = errors.New("database error")
)

type AuthService interface {
	Login(ctx context.Context, req AuthLoginRequest) (*AuthResponse, error)
}

type authService struct {
	repository *user.UserRepository
}

func (s *authService) Login(ctx context.Context, req AuthLoginRequest) (*AuthResponse, error) {
	firebaseClient, err := firebase.NewAuthClient(ctx)
	if err != nil {
		return nil, ErrNetworkError
	}
	token, err := firebaseClient.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	email, _ := token.Claims["email"].(string)
	foundUser, err := s.repository.FindOrCreate(ctx, token.UID, email)
	if err != nil {
		return nil, fmt.Errorf("%w: sync user: %v", ErrDatabase, err)
	}

	return &AuthResponse{
		ID:    foundUser.ID.Hex(),
		Email: foundUser.Email,
	}, nil
}

func NewService(repository *user.UserRepository) AuthService {
	return &authService{
		repository: repository,
	}
}
