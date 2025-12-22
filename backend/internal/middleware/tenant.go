package middleware

import (
	"context"
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/gin-gonic/gin"
)

// TenantContextKey is the context key for tenant ID
type TenantContextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey TenantContextKey = "tenant_id"
)

// TenantMiddleware provides tenant validation middleware
type TenantMiddleware struct {
	tenantRepo *repositories.TenantRepository
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantRepo *repositories.TenantRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo: tenantRepo,
	}
}

// ValidateTenant returns a middleware that validates tenant exists and is active
// It extracts tenant_id from path parameter and sets it in context
func (m *TenantMiddleware) ValidateTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Param("tenant_id")

		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "tenant_id is required",
			})
			c.Abort()
			return
		}

		// Get tenant from repository
		tenant, err := m.tenantRepo.Get(c.Request.Context(), tenantID)
		if err != nil {
			if err == repositories.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   "tenant not found",
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to validate tenant",
			})
			c.Abort()
			return
		}

		// Check if tenant is active
		if !tenant.IsActive {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "tenant is not active",
			})
			c.Abort()
			return
		}

		// Set tenant ID in context
		c.Set(string(TenantIDKey), tenantID)

		// Set in request context as well for use in repositories/services
		ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireTenant returns a middleware that ensures tenant_id is present in context
// This is useful for routes that need tenant context but don't have it in the path
func (m *TenantMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantID(c)

		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "tenant context is required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetTenantID retrieves the tenant ID from the context
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get(string(TenantIDKey)); exists {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}
	return ""
}

// GetTenantIDFromContext retrieves the tenant ID from a standard context
func GetTenantIDFromContext(ctx context.Context) string {
	if tenantID := ctx.Value(TenantIDKey); tenantID != nil {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}
	return ""
}
