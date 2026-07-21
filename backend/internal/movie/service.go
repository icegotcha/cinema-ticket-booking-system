package movie

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MovieService interface {
	GetMovies(ctx context.Context) ([]MovieResponse, error)
	GetMovieByID(ctx context.Context, id string) (*MovieResponse, error)
}

type movieService struct {
	repository *MovieRepository
}

func NewMovieService(repository *MovieRepository) MovieService {
	return &movieService{
		repository: repository,
	}
}

func (s *movieService) GetMovies(ctx context.Context) ([]MovieResponse, error) {
	movies, err := s.repository.FindAllSummaries(ctx)
	if err != nil {
		return nil, fmt.Errorf("get movies: %w", err)
	}

	responses := make([]MovieResponse, 0, len(movies))
	for _, m := range movies {
		id := ""
		if m.ID != bson.NilObjectID {
			id = m.ID.Hex()
		}

		responses = append(responses, MovieResponse{
			ID:    id,
			Title: fmt.Sprint(m.SeriesTitle),
		})
	}

	return responses, nil
}

func (s *movieService) GetMovieByID(
	ctx context.Context,
	id string,
) (*MovieResponse, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}

	m, err := s.repository.FindSummaryByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("get movie by id: %w", err)
	}

	if m == nil {
		return nil, nil
	}

	respID := ""
	if m.ID != bson.NilObjectID {
		respID = m.ID.Hex()
	}

	return &MovieResponse{
		ID:    respID,
		Title: fmt.Sprint(m.SeriesTitle),
	}, nil
}
