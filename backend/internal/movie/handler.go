package movie

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	service MovieService
}

func NewMovieHandler(service MovieService) *MovieHandler {
	return &MovieHandler{
		service: service,
	}
}

func (h *MovieHandler) RegisterRoutes(router *gin.RouterGroup) {
	movies := router.Group("/movies")

	movies.GET("", h.GetMovies)
	movies.GET("/:id", h.GetMovieByID)
}

func (h *MovieHandler) GetMovies(c *gin.Context) {
	movies, err := h.service.GetMovies(c.Request.Context())
	if err != nil {
		log.Printf("failed to get movies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get movies",
		})
		return
	}

	c.JSON(http.StatusOK, movies)
}

func (h *MovieHandler) GetMovieByID(c *gin.Context) {
	movie, err := h.service.GetMovieByID(
		c.Request.Context(),
		c.Param("id"),
	)
	if err != nil {
		log.Printf("failed to get movie: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get movie",
		})
		return
	}
	if movie == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "movie not found",
		})
		return
	}

	c.JSON(http.StatusOK, movie)
}
