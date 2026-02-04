package middleware

import (
	"net/http"

	"kb-platform-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the x-user-name header set by upstream gateway
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.GetHeader("x-user-name")
		if userName == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "AUTHENTICATION_ERROR",
					Message: "Missing x-user-name header",
				},
			})
			c.Abort()
			return
		}

		c.Set("username", userName)
		c.Next()
	}
}
