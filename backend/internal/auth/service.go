package auth

import (
	"context"
	"errors"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/firebase"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrNetworkError       = errors.New("network error")
)

type AuthService interface {
	Login(ctx context.Context, req AuthLoginRequest) (*AuthResponse, error)
	Signup(ctx context.Context, req CreateAuthRequest) error
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
	tokenUID := token.UID
	user, err := s.repository.FindOne(ctx, user.User{FirebaseUID: tokenUID})

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return &AuthResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (s *authService) Signup(ctx context.Context, req CreateAuthRequest) error {
	newUser := user.User{
		FirebaseUID: req.UID,
		Email:       req.Email,
	}

	existing, err := s.repository.FindOne(ctx, user.User{Email: req.Email})
	if err == nil && existing != nil {
		return ErrEmailAlreadyExists
	}

	s.repository.Create(ctx, &newUser)
	return nil
}

func NewService(repository *user.UserRepository) AuthService {
	return &authService{
		repository: repository,
	}
}
