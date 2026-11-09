package booking

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/user"
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
	bookings.POST("/holds", user.CheckAuth(), h.HoldBookings)
	bookings.POST("/confirm", user.CheckAuth(), h.ConfirmBookings)
	bookings.DELETE("/holds", user.CheckAuth(), h.ReleaseBookings)
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

func (h *BookingHandler) HoldBookings(c *gin.Context) {
	var request HoldBookingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seat hold request"})
		return
	}
	request.UserID = c.GetString("firebase_uid")

	bookings, err := h.service.HoldBookings(c.Request.Context(), request)
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

func (h *BookingHandler) ConfirmBookings(c *gin.Context) {
	h.handleSeatAction(c, h.service.ConfirmBookings)
}

func (h *BookingHandler) ReleaseBookings(c *gin.Context) {
	h.handleSeatAction(c, h.service.ReleaseBookings)
}

func (h *BookingHandler) handleSeatAction(
	c *gin.Context,
	action func(context.Context, BookingSeatsRequest) ([]BookingResponse, error),
) {
	var request BookingSeatsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking request"})
		return
	}
	request.UserID = c.GetString("firebase_uid")

	bookings, err := action(c.Request.Context(), request)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrHoldNotFound) || errors.Is(err, ErrHoldExpired) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}
