package user

import "time"

type User struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	FirebaseUID string    `json:"firebase_uid" bson:"firebase_uid"`
	Email       string    `json:"email" bson:"email"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
}
