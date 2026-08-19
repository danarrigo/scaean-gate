package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path

	c.Next()

	latency := time.Since(start)
	status := c.Writer.Status()
	method := c.Request.Method
	reqID := c.GetString("requestId")
	clientIP := c.ClientIP()

	log.Printf("[HTTP] %s %s | Status: %d | Latency: %v | IP: %s | RequestID: %s",
		method, path, status, latency, clientIP, reqID,
	)
}
