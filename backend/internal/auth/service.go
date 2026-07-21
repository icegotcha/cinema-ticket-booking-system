package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/firebase"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrNetworkError       = errors.New("network error")
	ErrDatabase           = errors.New("database error")
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
	foundUser, err := s.repository.FindOne(ctx, bson.M{"firebase_uid": tokenUID})

	if err != nil {
		return nil, fmt.Errorf("%w: find user: %v", ErrDatabase, err)
	}
	if foundUser == nil {
		return nil, ErrUserNotFound
	}

	return &AuthResponse{
		ID:    foundUser.ID.Hex(),
		Email: foundUser.Email,
	}, nil
}

func (s *authService) Signup(ctx context.Context, req CreateAuthRequest) error {
	now := time.Now().UTC()
	newUser := user.User{
		FirebaseUID: req.UID,
		Email:       req.Email,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	existing, err := s.repository.FindOne(ctx, bson.M{"email": req.Email})
	if err != nil {
		return fmt.Errorf("%w: check existing user: %v", ErrDatabase, err)
	}
	if existing != nil {
		return ErrEmailAlreadyExists
	}

	if _, err := s.repository.Create(ctx, &newUser); err != nil {
		return fmt.Errorf("%w: create user: %v", ErrDatabase, err)
	}
	return nil
}

func NewService(repository *user.UserRepository) AuthService {
	return &authService{
		repository: repository,
	}
}
