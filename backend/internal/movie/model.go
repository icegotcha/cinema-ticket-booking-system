package movie

import "go.mongodb.org/mongo-driver/v2/bson"

type Movie struct {
	ID           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	PosterLink   string        `bson:"poster_link" json:"poster_link"`
	SeriesTitle  any           `bson:"series_title" json:"series_title"`
	ReleasedYear int           `bson:"released_year" json:"released_year"`
	Certificate  string        `bson:"certificate" json:"certificate"`
	Runtime      string        `bson:"runtime" json:"runtime"`
	Genre        string        `bson:"genre" json:"genre"`
	Rating       float64       `bson:"rating" json:"rating"`
	Overview     string        `bson:"overview" json:"overview"`
	MetaScore    string        `bson:"meta_score" json:"meta_score"`
	Director     string        `bson:"director" json:"director"`
	Star1        string        `bson:"star1" json:"star1"`
	Star2        string        `bson:"star2" json:"star2"`
	Star3        string        `bson:"star3" json:"star3"`
	Star4        string        `bson:"star4" json:"star4"`
	NoOfVotes    int64         `bson:"no_of_votes" json:"no_of_votes"`
	Gross        string        `bson:"gross" json:"gross"`
}

type MovieResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	PosterLink string `json:"poster_link"`
}
