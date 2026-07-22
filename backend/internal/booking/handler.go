package booking

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	service BookingService
}

func NewBookingHandler(service BookingService) *BookingHandler {
	return &BookingHandler{
		service: service,
	}
}

func (h *BookingHandler) RegisterRoutes(router *gin.RouterGroup) {
	bookings := router.Group("/bookings")

	bookings.GET("", h.GetBookings)
	bookings.POST("", h.CreateBookings)
}

func (h *BookingHandler) GetBookings(c *gin.Context) {
	bookings, err := h.service.GetBookings(c.Request.Context(), c.Query("show_time_id"))
	if err != nil {
		log.Printf("failed to get bookings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get bookings",
		})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) CreateBookings(c *gin.Context) {
	var request CreateBookingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking request"})
		return
	}

	bookings, err := h.service.CreateBookings(c.Request.Context(), request)
	if err != nil {
		if errors.Is(err, ErrSeatUnavailable) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		log.Printf("failed to create bookings: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bookings)
}
