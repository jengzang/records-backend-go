package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"
)

// Logger middleware logs HTTP requests using structured logging
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get method
		method := c.Request.Method

		// Build query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Prepare log fields
		fields := logrus.Fields{
			"method":      method,
			"path":        path,
			"status":      statusCode,
			"duration_ms": latency.Milliseconds(),
			"client_ip":   clientIP,
		}

		// Add error information if present
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("HTTP Request", nil, fields)
		} else if statusCode >= 400 {
			logger.Warn("HTTP Request", fields)
		} else {
			logger.Info("HTTP Request", fields)
		}
	}
}

