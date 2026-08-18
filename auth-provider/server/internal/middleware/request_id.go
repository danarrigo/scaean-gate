package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware(c *gin.Context) {
	reqID := c.GetHeader("X-Request-ID")
	if reqID == "" {
		reqID = uuid.New().String()
	}

	c.Set("requestId", reqID)
	c.Header("X-Request-ID", reqID)
	c.Next()
}
