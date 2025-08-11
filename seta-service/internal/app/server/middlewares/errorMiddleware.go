package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CustomError represents a custom error structure.
type CustomError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *CustomError) Error() string {
	return e.Message
}

// ErrorHandler is a middleware to handle errors consistently.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // process request

		// This part executes after the handler
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Log the error
			log.Error().Err(err).Msg("An error occurred")

			// Check for our custom error type
			if appErr, ok := err.(*CustomError); ok {
				c.JSON(appErr.Code, gin.H{"error": appErr.Message})
				return
			}

			// Handle other generic errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
		}
	}
}
