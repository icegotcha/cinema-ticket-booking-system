package admin

import (
	"context"
	"errors"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AdminRepository struct {
	*database.Repository[AdminAccount]
	sessions *database.Repository[AdminSession]
}

func NewAdminRepository(db *mongo.Database) *AdminRepository {
	return &AdminRepository{
		Repository: database.NewRepository[AdminAccount](db, "admins"),
		sessions:   database.NewRepository[AdminSession](db, "admin_sessions"),
	}
}

func (r *AdminRepository) FindAccount(ctx context.Context, username string) (*AdminAccount, error) {
	return r.FindOne(ctx, bson.M{"username": username})
}

func (r *AdminRepository) CreateSession(ctx context.Context, session AdminSession) error {
	_, err := r.sessions.Create(ctx, &session)
	return err
}

func (r *AdminRepository) FindSession(ctx context.Context, tokenHash string) (*AdminSession, error) {
	return r.sessions.FindOne(ctx, bson.M{"token_hash": tokenHash})
}

func (r *AdminRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	err := r.sessions.DeleteOne(ctx, bson.M{"token_hash": tokenHash})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	return err
}
