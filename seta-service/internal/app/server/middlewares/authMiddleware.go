package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"seta/internal/pkg/errorHandling"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a gin middleware for JWT authentication.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "Authorization header is missing"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Prepare the GraphQL query
		type GQLVariables struct {
			Token string `json:"token"`
		}

		// Prepare the GraphQL query using a struct that will marshal correctly
		query := gin.H{
			"query": `
                query VerifyToken($token: String!) {
                    verifyToken(token: $token) {
                        success
                        user {
                            userId
                            role
                        }
                    }
                }
            `,
			"variables": GQLVariables{
				Token: tokenString,
			},
		}

		jsonQuery, err := json.Marshal(query)
		if err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to create GraphQL query"})
			c.Abort()
			return
		}

		// Make the request to the user-service
		resp, err := http.Post("http://localhost:4000/users", "application/json", bytes.NewBuffer(jsonQuery))
		if err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusServiceUnavailable, Message: "Failed to connect to user service"})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		var result struct {
			Data struct {
				VerifyToken struct {
					Success bool `json:"success"`
					User    struct {
						UserID string `json:"userId"`
						Role   string `json:"role"`
					} `json:"user"`
				} `json:"verifyToken"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusInternalServerError, Message: "Failed to decode user service response"})
			c.Abort()
			return
		}

		if !result.Data.VerifyToken.Success {
			_ = c.Error(&errorHandling.CustomError{Code: http.StatusUnauthorized, Message: "Invalid token"})
			c.Abort()
			return
		}

		// If successful, set user info and continue
		c.Set("userId", result.Data.VerifyToken.User.UserID)
		c.Set("role", result.Data.VerifyToken.User.Role)

		c.Next()
	}
}