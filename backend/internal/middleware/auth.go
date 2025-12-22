package middleware

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "user_email"
	// FirebaseTokenKey is the context key for Firebase token
	FirebaseTokenKey ContextKey = "firebase_token"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authClient *auth.Client
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authClient *auth.Client) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
	}
}

// AuthRequired returns a middleware that requires authentication
// It verifies Firebase ID tokens and sets user information in the context
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing authorization header",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		idToken := parts[1]

		// Verify the ID token
		token, err := m.authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract user information from token
		userID := token.UID
		userEmail := ""
		if email, ok := token.Claims["email"].(string); ok {
			userEmail = email
		}

		// Set user information in context
		c.Set(string(UserIDKey), userID)
		c.Set(string(UserEmailKey), userEmail)
		c.Set(string(FirebaseTokenKey), token)

		// Set in request context as well for use in repositories/services
		ctx := context.WithValue(c.Request.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserEmailKey, userEmail)
		ctx = context.WithValue(ctx, FirebaseTokenKey, token)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalAuth returns a middleware that extracts user info if token is present
// but doesn't require authentication (useful for public endpoints that show different data for authenticated users)
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		idToken := parts[1]

		// Try to verify the token
		token, err := m.authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			// If token is invalid, just continue without user info
			c.Next()
			return
		}

		// Extract user information from token
		userID := token.UID
		userEmail := ""
		if email, ok := token.Claims["email"].(string); ok {
			userEmail = email
		}

		// Set user information in context
		c.Set(string(UserIDKey), userID)
		c.Set(string(UserEmailKey), userEmail)
		c.Set(string(FirebaseTokenKey), token)

		// Set in request context as well
		ctx := context.WithValue(c.Request.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserEmailKey, userEmail)
		ctx = context.WithValue(ctx, FirebaseTokenKey, token)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get(string(UserIDKey)); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return ""
}

// GetUserEmail retrieves the user email from the context
func GetUserEmail(c *gin.Context) string {
	if userEmail, exists := c.Get(string(UserEmailKey)); exists {
		if email, ok := userEmail.(string); ok {
			return email
		}
	}
	return ""
}

// GetFirebaseToken retrieves the Firebase token from the context
func GetFirebaseToken(c *gin.Context) *auth.Token {
	if token, exists := c.Get(string(FirebaseTokenKey)); exists {
		if t, ok := token.(*auth.Token); ok {
			return t
		}
	}
	return nil
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	return GetUserID(c) != ""
}
