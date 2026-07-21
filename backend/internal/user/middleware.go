package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/firebase"
)

func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")

		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		const bearerPrefix = "Bearer "

		if !strings.HasPrefix(authorization, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header must use Bearer token",
			})
			return
		}

		idToken := strings.TrimSpace(
			strings.TrimPrefix(authorization, bearerPrefix),
		)

		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "firebase ID token is required",
			})
			return
		}
		client, err := firebase.NewAuthClient(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to create Firebase Auth client",
			})
			return
		}
		token, err := client.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired Firebase ID token",
			})
			return
		}

		c.Set("firebase_uid", token.UID)
		c.Set("firebase_token", token)

		c.Next()
	}
}
