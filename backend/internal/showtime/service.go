package showtime

import (
	"context"
	"fmt"

	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/movie"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ShowTimeService interface {
	GetShowTimes(ctx context.Context, movieID string) ([]ShowTimeResponse, error)
	GetShowTimeByID(ctx context.Context, id string) (*ShowTimeResponse, error)
}

type showTimeService struct {
	repository      *ShowTimeRepository
	movieRepository *movie.MovieRepository
}

func NewShowTimeService(repository *ShowTimeRepository, movieRepository *movie.MovieRepository) ShowTimeService {
	return &showTimeService{
		repository:      repository,
		movieRepository: movieRepository,
	}
}

func (s *showTimeService) GetShowTimes(ctx context.Context, movieID string) ([]ShowTimeResponse, error) {
	movieObjectID, err := bson.ObjectIDFromHex(movieID)
	if err != nil {
		return nil, nil
	}

	showtimes, err := s.repository.FindMany(ctx, bson.M{"movie_id": movieObjectID})
	if err != nil {
		return nil, fmt.Errorf("get showtimes: %w", err)
	}

	m, err := s.movieRepository.FindSummaryByID(ctx, movieObjectID)
	if err != nil {
		return nil, fmt.Errorf("get movie by id: %w", err)
	}

	if m == nil {
		return nil, fmt.Errorf("movie not found")
	}

	responses := make([]ShowTimeResponse, 0, len(showtimes))
	for _, st := range showtimes {
		responses = append(responses, ShowTimeResponse{
			ID:        st.ID.Hex(),
			StartTime: int64(st.StartTime.T),
			Movie: movie.MovieResponse{
				ID:    m.ID.Hex(),
				Title: fmt.Sprint(m.SeriesTitle),
			},
		})
	}

	return responses, nil
}

func (s *showTimeService) GetShowTimeByID(ctx context.Context, id string) (*ShowTimeResponse, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}

	showtime, err := s.repository.FindOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return nil, fmt.Errorf("get showtime by id: %w", err)
	}

	if showtime == nil {
		return nil, nil
	}

	m, err := s.movieRepository.FindSummaryByID(ctx, showtime.MovieId)
	if err != nil {
		return nil, fmt.Errorf("get movie by id: %w", err)
	}

	if m == nil {
		return nil, fmt.Errorf("movie not found")
	}

	respID := ""
	if showtime.ID != bson.NilObjectID {
		respID = showtime.ID.Hex()
	}

	return &ShowTimeResponse{
		ID:        respID,
		StartTime: int64(showtime.StartTime.T),
		Movie: movie.MovieResponse{
			ID:    m.ID.Hex(),
			Title: fmt.Sprint(m.SeriesTitle),
		},
	}, nil
}
