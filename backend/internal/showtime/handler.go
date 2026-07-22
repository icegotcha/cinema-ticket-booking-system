package showtime

// use gin
import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShowTimeHandler struct {
	service ShowTimeService
}

func NewShowTimeHandler(service ShowTimeService) *ShowTimeHandler {
	return &ShowTimeHandler{
		service: service,
	}
}

func (h *ShowTimeHandler) RegisterRoutes(router *gin.RouterGroup) {
	showtimes := router.Group("/showtimes")
	showtimes.GET("/movies/:movie_id", h.GetShowTimes)
	showtimes.GET("/movies/:movie_id/st/:id", h.GetShowTimeByID)
}

func (h *ShowTimeHandler) GetShowTimes(c *gin.Context) {
	movieID := c.Param("movie_id")
	showtimes, err := h.service.GetShowTimes(c.Request.Context(), movieID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, showtimes)
}

func (h *ShowTimeHandler) GetShowTimeByID(c *gin.Context) {
	id := c.Param("id")
	showtime, err := h.service.GetShowTimeByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if showtime == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
		return
	}
	c.JSON(http.StatusOK, showtime)
}
