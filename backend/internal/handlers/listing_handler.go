package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ListingHandler handles listing-related HTTP requests
type ListingHandler struct {
	listingService *services.ListingService
}

// NewListingHandler creates a new listing handler
func NewListingHandler(listingService *services.ListingService) *ListingHandler {
	return &ListingHandler{
		listingService: listingService,
	}
}

// RegisterRoutes registers listing routes (tenant-scoped)
func (h *ListingHandler) RegisterRoutes(router *gin.RouterGroup) {
	listings := router.Group("/:tenant_id/listings")
	{
		listings.POST("", h.CreateListing)
		listings.GET("/:id", h.GetListing)
		listings.PUT("/:id", h.UpdateListing)
		listings.DELETE("/:id", h.DeleteListing)
		listings.GET("", h.ListListings)
		listings.POST("/:id/set-canonical", h.SetCanonical)
	}
}

// CreateListing creates a new listing
// @Summary Create a new listing
// @Description Create a new listing for a property
// @Tags listings
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param listing body models.Listing true "Listing data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/listings [post]
func (h *ListingHandler) CreateListing(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	var listing models.Listing
	if err := c.ShouldBindJSON(&listing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Set tenant ID from path parameter
	listing.TenantID = tenantID

	if err := h.listingService.CreateListing(c.Request.Context(), &listing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    listing,
	})
}

// GetListing retrieves a listing by ID
// @Summary Get listing by ID
// @Description Get listing details by ID
// @Tags listings
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Listing ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/listings/{id} [get]
func (h *ListingHandler) GetListing(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	listing, err := h.listingService.GetListing(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "listing not found",
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
		"data":    listing,
	})
}

// UpdateListing updates a listing
// @Summary Update listing
// @Description Update listing information
// @Tags listings
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Listing ID"
// @Param updates body map[string]interface{} true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/listings/{id} [put]
func (h *ListingHandler) UpdateListing(c *gin.Context) {
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

	if err := h.listingService.UpdateListing(c.Request.Context(), tenantID, id, updates); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "listing not found",
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
		"data":    gin.H{"message": "listing updated successfully"},
	})
}

// DeleteListing deletes a listing
// @Summary Delete listing
// @Description Delete a listing
// @Tags listings
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Listing ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/listings/{id} [delete]
func (h *ListingHandler) DeleteListing(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.listingService.DeleteListing(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "listing not found",
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
		"data":    gin.H{"message": "listing deleted successfully"},
	})
}

// ListListings lists all listings for a tenant
// @Summary List listings
// @Description List all listings for a tenant with pagination
// @Tags listings
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(created_at)
// @Param property_id query string false "Property ID filter"
// @Param broker_id query string false "Broker ID filter"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/listings [get]
func (h *ListingHandler) ListListings(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	var listings []*models.Listing
	var err error

	// Check for filters
	if propertyID := c.Query("property_id"); propertyID != "" {
		listings, err = h.listingService.ListListingsByProperty(c.Request.Context(), tenantID, propertyID, opts)
	} else if brokerID := c.Query("broker_id"); brokerID != "" {
		listings, err = h.listingService.ListListingsByBroker(c.Request.Context(), tenantID, brokerID, opts)
	} else {
		listings, err = h.listingService.ListListings(c.Request.Context(), tenantID, opts)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listings,
		"count":   len(listings),
	})
}

// SetCanonical sets a listing as the canonical listing for its property
// @Summary Set canonical listing
// @Description Set a listing as the canonical listing for its property
// @Tags listings
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Listing ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/listings/{id}/set-canonical [post]
func (h *ListingHandler) SetCanonical(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.listingService.SetCanonical(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "listing not found",
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
		"data":    gin.H{"message": "canonical listing set successfully"},
	})
}

