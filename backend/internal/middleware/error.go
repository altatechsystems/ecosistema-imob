package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorRecovery returns a middleware that recovers from panics
func ErrorRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic
				logPanic(c, err, stack)

				// Return 500 error
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "internal server error",
					"request_id": GetRequestID(c),
				})

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}

// logPanic logs a panic with stack trace
func logPanic(c *gin.Context, err interface{}, stack []byte) {
	requestID := GetRequestID(c)
	path := c.Request.URL.Path
	method := c.Request.Method

	// Build error message
	errMsg := fmt.Sprintf("PANIC RECOVERED\n")
	errMsg += fmt.Sprintf("Request ID: %s\n", requestID)
	errMsg += fmt.Sprintf("Method: %s\n", method)
	errMsg += fmt.Sprintf("Path: %s\n", path)
	errMsg += fmt.Sprintf("Error: %v\n", err)
	errMsg += fmt.Sprintf("Stack Trace:\n%s\n", string(stack))

	// Write to error log
	gin.DefaultErrorWriter.Write([]byte(errMsg))
}

// ErrorHandler returns a middleware that handles application errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Determine status code based on error type
			statusCode := http.StatusInternalServerError

			// Check if status code was already set
			if c.Writer.Status() != http.StatusOK {
				statusCode = c.Writer.Status()
			}

			// Return error response
			c.JSON(statusCode, gin.H{
				"success": false,
				"error":   err.Error(),
				"request_id": GetRequestID(c),
			})
		}
	}
}
