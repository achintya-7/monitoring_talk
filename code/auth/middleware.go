package main

import "github.com/gin-gonic/gin"

func CorrelationIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// attach correlation ID from request headers to context
		correlationId := c.GetHeader("x-correlation-id")
		if correlationId == "" {
			correlationId = "unknown"
		}

		c.Set(CORRELATION_ID, correlationId)
		c.Next()
	}
}
