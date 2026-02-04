package middleware

import (
	"net/http"
	"strings"

	"kb-platform-gateway/internal/auth"
	"kb-platform-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtManager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "AUTHENTICATION_ERROR",
					Message: "Missing authorization header",
				},
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "AUTHENTICATION_ERROR",
					Message: "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "AUTHENTICATION_ERROR",
					Message: "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}
