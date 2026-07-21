package user

import (
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepository struct {
	*database.Repository[User]
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		Repository: database.NewRepository[User](db, "users"),
	}
}
