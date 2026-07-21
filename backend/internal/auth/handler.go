package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service AuthService
}

func NewHandler(service AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	authApi := rg.Group("/auth")
	{
		authApi.POST("/login", h.Login)
		authApi.POST("/signup", h.Signup)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		switch err {
		case ErrNetworkError:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			if errors.Is(err, ErrDatabase) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": ErrDatabase.Error()})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			}
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req CreateAuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Signup(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrDatabase.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, nil)
}
