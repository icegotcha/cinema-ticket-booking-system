package user

import (
	"context"
	"time"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository struct {
	*database.Repository[User]
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		Repository: database.NewRepository[User](db, "users"),
	}
}

func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.Collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "firebase_uid", Value: 1}},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"firebase_uid": bson.M{"$type": "string"}}),
	})
	return err
}

func (r *UserRepository) FindOrCreate(ctx context.Context, uid, email string) (*User, error) {
	now := time.Now().UTC()
	filter := bson.M{"firebase_uid": uid}
	var profile User
	err := r.Collection().FindOneAndUpdate(ctx, filter, bson.M{
		"$set":         bson.M{"email": email, "updatedAt": now},
		"$setOnInsert": bson.M{"firebase_uid": uid, "createdAt": now},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Decode(&profile)
	// Another request may have inserted this UID while our upsert was running.
	if mongo.IsDuplicateKeyError(err) {
		if lookupErr := r.Collection().FindOne(ctx, filter).Decode(&profile); lookupErr != nil {
			return nil, lookupErr
		}
		return &profile, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}
