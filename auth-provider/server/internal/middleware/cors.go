package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var allowedOrigins = map[string]bool{
	"http://localhost:4200": true,
	"http://localhost:4201": true,
	"http://localhost:4202": true,
}

func CORSMiddleware(c *gin.Context) {
	origin := c.GetHeader("Origin")

	if allowedOrigins[origin] {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, Accept, Origin")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	}

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
