package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PropertyHandler handles property-related HTTP requests
type PropertyHandler struct {
	propertyService *services.PropertyService
}

// NewPropertyHandler creates a new property handler
func NewPropertyHandler(propertyService *services.PropertyService) *PropertyHandler {
	return &PropertyHandler{
		propertyService: propertyService,
	}
}

// RegisterRoutes registers property routes (tenant-scoped)
func (h *PropertyHandler) RegisterRoutes(router *gin.RouterGroup) {
	properties := router.Group("/properties")
	{
		properties.POST("", h.CreateProperty)
		properties.GET("/:id", h.GetProperty)
		properties.GET("/slug/:slug", h.GetPropertyBySlug)
		properties.PUT("/:id", h.UpdateProperty)
		properties.DELETE("/:id", h.DeleteProperty)
		properties.GET("", h.ListProperties)
		properties.POST("/:id/status", h.UpdateStatus)
		properties.POST("/:id/visibility", h.UpdateVisibility)
		properties.GET("/:id/duplicates", h.CheckDuplicates)

		// PROMPT 08: Property Status Confirmation
		properties.PATCH("/:id/confirmations", h.ConfirmPropertyStatusPrice)
		properties.POST("/:id/owner-confirmation-link", h.GenerateOwnerConfirmationLink)
	}
}

// CreateProperty creates a new property
// @Summary Create a new property
// @Description Create a new property for a tenant
// @Tags properties
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param property body models.Property true "Property data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties [post]
func (h *PropertyHandler) CreateProperty(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	var property models.Property
	if err := c.ShouldBindJSON(&property); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Set tenant ID from path parameter
	property.TenantID = tenantID

	if err := h.propertyService.CreateProperty(c.Request.Context(), &property); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    property,
	})
}

// GetProperty retrieves a property by ID
// @Summary Get property by ID
// @Description Get property details by ID
// @Tags properties
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id} [get]
func (h *PropertyHandler) GetProperty(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	property, err := h.propertyService.GetProperty(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    property,
	})
}

// GetPropertyBySlug retrieves a property by slug
// @Summary Get property by slug
// @Description Get property details by slug
// @Tags properties
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param slug path string true "Property Slug"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/slug/{slug} [get]
func (h *PropertyHandler) GetPropertyBySlug(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	slug := c.Param("slug")

	property, err := h.propertyService.GetPropertyBySlug(c.Request.Context(), tenantID, slug)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    property,
	})
}

// UpdateProperty updates a property
// @Summary Update property
// @Description Update property information
// @Tags properties
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Param updates body map[string]interface{} true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id} [put]
func (h *PropertyHandler) UpdateProperty(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.propertyService.UpdateProperty(c.Request.Context(), tenantID, id, updates); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "property updated successfully"},
	})
}

// DeleteProperty deletes a property
// @Summary Delete property
// @Description Delete a property
// @Tags properties
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id} [delete]
func (h *PropertyHandler) DeleteProperty(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.propertyService.DeleteProperty(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "property deleted successfully"},
	})
}

// ListProperties lists all properties for a tenant
// @Summary List properties
// @Description List all properties for a tenant with filters and pagination
// @Tags properties
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(created_at)
// @Param property_type query string false "Property type filter"
// @Param status query string false "Status filter"
// @Param visibility query string false "Visibility filter"
// @Param city query string false "City filter"
// @Param neighborhood query string false "Neighborhood filter"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties [get]
func (h *PropertyHandler) ListProperties(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	// Parse filters
	filters := &repositories.PropertyFilters{}

	if propertyType := c.Query("property_type"); propertyType != "" {
		propType := models.PropertyType(propertyType)
		filters.PropertyType = &propType
	}

	if status := c.Query("status"); status != "" {
		propStatus := models.PropertyStatus(status)
		filters.Status = &propStatus
	}

	if visibility := c.Query("visibility"); visibility != "" {
		propVisibility := models.PropertyVisibility(visibility)
		filters.Visibility = &propVisibility
	}

	if city := c.Query("city"); city != "" {
		filters.City = city
	}

	if neighborhood := c.Query("neighborhood"); neighborhood != "" {
		filters.Neighborhood = neighborhood
	}

	if ownerID := c.Query("owner_id"); ownerID != "" {
		filters.OwnerID = ownerID
	}

	properties, err := h.propertyService.ListProperties(c.Request.Context(), tenantID, filters, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get total count (without pagination)
	total, err := h.propertyService.CountProperties(c.Request.Context(), tenantID, filters)
	if err != nil {
		// Log error but don't fail the request - count is optional
		total = len(properties)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    properties,
		"count":   len(properties),
		"total":   total,
	})
}


// UpdateStatus updates the status of a property
// @Summary Update property status
// @Description Update the status of a property
// @Tags properties
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Param body body PropertyUpdateStatusRequest true "Status update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id}/status [post]
func (h *PropertyHandler) UpdateStatus(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	var req PropertyUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.propertyService.UpdateStatus(c.Request.Context(), tenantID, id, req.Status); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "property status updated successfully"},
	})
}

// UpdateVisibilityRequest represents the request body for updating property visibility
type UpdateVisibilityRequest struct {
	Visibility models.PropertyVisibility `json:"visibility" binding:"required"`
}

// UpdateVisibility updates the visibility of a property
// @Summary Update property visibility
// @Description Update the visibility of a property
// @Tags properties
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Param body body UpdateVisibilityRequest true "Visibility update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id}/visibility [post]
func (h *PropertyHandler) UpdateVisibility(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	var req UpdateVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.propertyService.UpdateVisibility(c.Request.Context(), tenantID, id, req.Visibility); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "property visibility updated successfully"},
	})
}

// CheckDuplicates checks for duplicate properties
// @Summary Check for duplicate properties
// @Description Check for duplicate properties by fingerprint
// @Tags properties
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id}/duplicates [get]
func (h *PropertyHandler) CheckDuplicates(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	// Get the property first
	property, err := h.propertyService.GetProperty(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Check for duplicates
	duplicates, err := h.propertyService.CheckDuplicates(c.Request.Context(), property)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       duplicates,
		"count":      len(duplicates),
		"has_dupes":  len(duplicates) > 0,
	})
}

// ========== PROMPT 08: Property Status Confirmation ==========

// ConfirmPropertyStatusPriceRequest represents the request to confirm status/price
type ConfirmPropertyStatusPriceRequest struct {
	ConfirmStatus      *models.PropertyStatus `json:"confirm_status,omitempty"`       // available, unavailable
	ConfirmPriceAmount *float64               `json:"confirm_price_amount,omitempty"` // new price
	Note               string                 `json:"note,omitempty"`                 // internal note
	Reason             string                 `json:"reason,omitempty"`               // operator_reported, owner_reported, stale_refresh
}

// ConfirmPropertyStatusPrice handles operator confirmation of property status/price
// PATCH /api/v1/properties/{propertyId}/confirmations
// @Summary Confirm property status and/or price
// @Description Operator confirms property availability status and/or price (PROMPT 08)
// @Tags properties
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Param body body ConfirmPropertyStatusPriceRequest true "Confirmation data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id}/confirmations [patch]
func (h *PropertyHandler) ConfirmPropertyStatusPrice(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("id")

	var req ConfirmPropertyStatusPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get user/actor from context (set by auth middleware)
	actorID, exists := c.Get("user_id")
	if !exists {
		actorID = "system"
	}

	// Call service to confirm
	property, err := h.propertyService.ConfirmPropertyStatusPrice(
		c.Request.Context(),
		tenantID,
		propertyID,
		actorID.(string),
		req.ConfirmStatus,
		req.ConfirmPriceAmount,
		req.Note,
		req.Reason,
	)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    property,
	})
}

// GenerateOwnerConfirmationLinkRequest represents the request to generate owner link
type GenerateOwnerConfirmationLinkRequest struct {
	DeliveryHint string  `json:"delivery_hint,omitempty"` // whatsapp, sms, email (optional, apenas registro)
	OwnerID      *string `json:"owner_id,omitempty"`      // opcional
}

// GenerateOwnerConfirmationLinkResponse represents the response
type GenerateOwnerConfirmationLinkResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
	ExpiresAt       string `json:"expires_at"`
	TokenID         string `json:"token_id"`
}

// GenerateOwnerConfirmationLink generates a secure link for passive owner confirmation
// POST /api/v1/properties/{propertyId}/owner-confirmation-link
// @Summary Generate owner confirmation link
// @Description Generate a secure, time-limited link for property owner to confirm status/price without login (PROMPT 08)
// @Tags properties
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Property ID"
// @Param body body GenerateOwnerConfirmationLinkRequest true "Link generation params"
// @Success 200 {object} GenerateOwnerConfirmationLinkResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{id}/owner-confirmation-link [post]
func (h *PropertyHandler) GenerateOwnerConfirmationLink(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("id")

	var req GenerateOwnerConfirmationLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get user/actor from context
	actorID, exists := c.Get("user_id")
	if !exists {
		actorID = "system"
	}

	// Call service to generate link
	confirmationURL, tokenID, expiresAt, err := h.propertyService.GenerateOwnerConfirmationLink(
		c.Request.Context(),
		tenantID,
		propertyID,
		actorID.(string),
		req.OwnerID,
		req.DeliveryHint,
	)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "property not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": GenerateOwnerConfirmationLinkResponse{
			ConfirmationURL: confirmationURL,
			ExpiresAt:       expiresAt.Format("2006-01-02T15:04:05Z07:00"),
			TokenID:         tokenID,
		},
	})
}

