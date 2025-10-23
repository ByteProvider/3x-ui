package middleware

import (
	"strings"

	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"
	"github.com/gin-gonic/gin"
)

// ApiKeyAuth is a middleware that checks for API key authentication
// It looks for the X-API-Key header and validates it against the database
func ApiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If already logged in via session, continue
		if session.IsLogin(c) {
			c.Next()
			return
		}

		// Check for API key in header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Also check for Authorization: Bearer <token>
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey != "" {
			userService := service.UserService{}
			user, err := userService.GetUserByApiKey(apiKey)
			if err == nil && user != nil {
				// Set the user in session for this request
				session.SetLoginUser(c, user)
				c.Next()
				return
			}
		}

		c.Next()
	}
}

