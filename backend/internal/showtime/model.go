package showtime

import (
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ShowTime struct {
	ID        bson.ObjectID  `json:"id" bson:"_id,omitempty"`
	MovieId   bson.ObjectID  `json:"movie_id" bson:"movie_id,omitempty"`
	StartTime bson.Timestamp `json:"start_time" bson:"start_time,omitempty"`
}

type ShowTimeResponse struct {
	ID        string              `json:"id"`
	StartTime int64               `json:"start_time"`
	Movie     movie.MovieResponse `json:"movie"`
}
