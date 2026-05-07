package middleware

import (
	"time"

	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs HTTP requests
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		log.Info("HTTP Request",
			"method", method,
			"path", path,
			"status", status,
			"latency", latency.String(),
			"ip", clientIP,
		)
	}
}
