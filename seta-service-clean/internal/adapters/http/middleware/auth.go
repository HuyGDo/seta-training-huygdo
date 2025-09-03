package middleware

import (
	"net/http"
	"seta/internal/application/ports"
	"seta/pkg/errorHandling"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware uses the AuthService port to validate a token.
// It now uses the custom error handler for all failure cases.
func AuthMiddleware(authService ports.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "Authorization header is missing"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		user, err := authService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userId", user.ID.String())
		c.Set("role", user.Role)
		c.Next()
	}
}
