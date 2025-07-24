package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsAuthorized(authorizedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User role not found in token"})
			return
		}

		userRole := role.(string)
		isAuthorized := false
		for _, authorizedRole := range authorizedRoles {
			if userRole == authorizedRole {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not authorized to perform this action"})
			return
		}

		c.Next()
	}
}
