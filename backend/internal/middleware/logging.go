package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey is the context key for request ID
const RequestIDKey = "request_id"

// RequestLogger returns a middleware that logs HTTP requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get status code
		statusCode := c.Writer.Status()

		// Get method
		method := c.Request.Method

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Get error message if any
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Log the request
		logFields := map[string]interface{}{
			"request_id":    requestID,
			"client_ip":     clientIP,
			"method":        method,
			"path":          path,
			"status_code":   statusCode,
			"latency_ms":    latency.Milliseconds(),
			"latency":       latency.String(),
			"user_agent":    c.Request.UserAgent(),
		}

		// Add user ID if authenticated
		if userID := GetUserID(c); userID != "" {
			logFields["user_id"] = userID
		}

		// Add error message if present
		if errorMessage != "" {
			logFields["error"] = errorMessage
		}

		// Log based on status code
		if statusCode >= 500 {
			logError(logFields)
		} else if statusCode >= 400 {
			logWarn(logFields)
		} else {
			logInfo(logFields)
		}
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// logInfo logs an info message
// This is a simple implementation. In production, you would use a proper logging library
// like logrus or zap for structured logging
func logInfo(fields map[string]interface{}) {
	// For now, use Gin's default logger
	// In production, replace with proper structured logging
	gin.DefaultWriter.Write([]byte(formatLogMessage("INFO", fields)))
}

// logWarn logs a warning message
func logWarn(fields map[string]interface{}) {
	gin.DefaultWriter.Write([]byte(formatLogMessage("WARN", fields)))
}

// logError logs an error message
func logError(fields map[string]interface{}) {
	gin.DefaultErrorWriter.Write([]byte(formatLogMessage("ERROR", fields)))
}

// formatLogMessage formats a log message with fields
func formatLogMessage(level string, fields map[string]interface{}) string {
	msg := level + " "
	for key, value := range fields {
		msg += key + "=" + toString(value) + " "
	}
	return msg + "\n"
}

// toString converts a value to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val))
	case int64:
		return string(rune(val))
	case float64:
		return string(rune(int(val)))
	default:
		return ""
	}
}
