package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(origins []string) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowedOrigins[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, allowed := allowedOrigins[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-Request-ID")
			c.Header("Access-Control-Expose-Headers", "X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
