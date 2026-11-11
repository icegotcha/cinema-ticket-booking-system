package admin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CheckAuth(service AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, found := strings.Cut(c.GetHeader("Authorization"), " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || len(token) != 43 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Admin login required"})
			return
		}
		tokenHash, err := service.VerifyToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, ErrSessionExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrSessionExpired.Error()})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to verify session"})
			}
			return
		}
		c.Set("admin_token_hash", tokenHash)
		c.Next()
	}
}
