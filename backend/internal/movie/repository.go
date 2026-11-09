package movie

import (
	"context"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MovieRepository struct {
	*database.Repository[Movie]
}

func NewMovieRepository(db *mongo.Database) *MovieRepository {
	return &MovieRepository{
		Repository: database.NewRepository[Movie](db, "movies"),
	}
}

var movieSummaryProjection = bson.M{
	"_id":          1,
	"poster_link":  1,
	"series_title": 1,
}

func (r *MovieRepository) FindAllSummaries(ctx context.Context) ([]Movie, error) {
	return r.FindAll(
		ctx,
		options.Find().SetProjection(movieSummaryProjection),
	)
}

func (r *MovieRepository) FindSummaryByID(
	ctx context.Context,
	id bson.ObjectID,
) (*Movie, error) {
	return r.FindOne(
		ctx,
		bson.M{"_id": id},
		options.FindOne().SetProjection(movieSummaryProjection),
	)
}
