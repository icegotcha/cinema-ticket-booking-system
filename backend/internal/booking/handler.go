package booking

import (
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
}

func (h *BookingHandler) GetBookings(c *gin.Context) {
	bookings, err := h.service.GetBookings(c.Request.Context())
	if err != nil {
		log.Printf("failed to get bookings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get bookings",
		})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
