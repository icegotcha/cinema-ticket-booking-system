package user

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID          bson.ObjectID `json:"id" bson:"_id,omitempty"`
	FirebaseUID string        `json:"firebase_uid" bson:"firebase_uid,omitempty"`
	Email       string        `json:"email" bson:"email"`
	CreatedAt   time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt" bson:"updatedAt"`
}
