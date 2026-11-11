package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) RegisterRoutes(router *gin.RouterGroup) {
	admins := router.Group("/admin")
	admins.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
	})
	admins.POST("/login", h.Login)

	protected := admins.Group("", CheckAuth(h.service))
	protected.POST("/logout", h.Logout)
	protected.GET("/bookings", h.GetBookings)
}

func (h *AdminHandler) Login(c *gin.Context) {
	var request AdminLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}
	response, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidCredentials.Error()})
		case errors.Is(err, ErrCreateSession):
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrCreateSession.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to sign in"})
		}
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) Logout(c *gin.Context) {
	if err := h.service.Logout(c.Request.Context(), c.GetString("admin_token_hash")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to sign out. Please try again"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) GetBookings(c *gin.Context) {
	response, err := h.service.GetBookings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
