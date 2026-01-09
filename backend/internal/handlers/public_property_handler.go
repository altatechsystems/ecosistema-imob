package handlers

import (
	"fmt"
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PublicPropertyHandler handles public property HTTP requests (cross-tenant)
// This handler is used by the public portal agregador to list properties from all tenants
type PublicPropertyHandler struct {
	propertyService *services.PropertyService
}

// NewPublicPropertyHandler creates a new public property handler
func NewPublicPropertyHandler(propertyService *services.PropertyService) *PublicPropertyHandler {
	return &PublicPropertyHandler{
		propertyService: propertyService,
	}
}

// ListPublicProperties lists all public properties across all tenants
// @Summary List public properties (cross-tenant)
// @Description List public and available properties from all tenants for portal agregador
// @Tags public-properties
// @Produce json
// @Param limit query int false "Maximum number of results" default(50)
// @Param offset query int false "Number of results to skip" default(0)
// @Param property_type query string false "Property type filter"
// @Param transaction_type query string false "Transaction type filter"
// @Param city query string false "City filter"
// @Param neighborhood query string false "Neighborhood filter"
// @Param min_price query float64 false "Minimum price"
// @Param max_price query float64 false "Maximum price"
// @Param min_bedrooms query int false "Minimum bedrooms"
// @Param min_bathrooms query int false "Minimum bathrooms"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/public/properties [get]
func (h *PublicPropertyHandler) ListPublicProperties(c *gin.Context) {
	// Parse pagination options
	opts := parsePaginationOptions(c)

	// Parse filters from query params
	filters := &repositories.PropertyFilters{}

	if propertyType := c.Query("property_type"); propertyType != "" {
		// Type conversion handled at service layer
		filters.PropertyType = (*repositories.PropertyType)(&propertyType)
	}

	if transactionType := c.Query("transaction_type"); transactionType != "" {
		filters.TransactionType = (*repositories.TransactionType)(&transactionType)
	}

	if city := c.Query("city"); city != "" {
		filters.City = city
	}

	if neighborhood := c.Query("neighborhood"); neighborhood != "" {
		filters.Neighborhood = neighborhood
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		var price float64
		if _, err := fmt.Sscanf(minPrice, "%f", &price); err == nil {
			filters.MinPrice = &price
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		var price float64
		if _, err := fmt.Sscanf(maxPrice, "%f", &price); err == nil {
			filters.MaxPrice = &price
		}
	}

	if minBedrooms := c.Query("min_bedrooms"); minBedrooms != "" {
		var bedrooms int
		if _, err := fmt.Sscanf(minBedrooms, "%d", &bedrooms); err == nil {
			filters.MinBedrooms = &bedrooms
		}
	}

	if minBathrooms := c.Query("min_bathrooms"); minBathrooms != "" {
		var bathrooms int
		if _, err := fmt.Sscanf(minBathrooms, "%d", &bathrooms); err == nil {
			filters.MinBathrooms = &bathrooms
		}
	}

	// Get public properties from service (across all tenants)
	properties, err := h.propertyService.ListAllPublicProperties(c.Request.Context(), filters, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list public properties",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    properties,
		"count":   len(properties),
	})
}

// GetPublicProperty retrieves a single public property by ID (cross-tenant)
// @Summary Get public property by ID (cross-tenant)
// @Description Get a single public and available property from any tenant
// @Tags public-properties
// @Produce json
// @Param id path string true "Property ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/public/properties/{id} [get]
func (h *PublicPropertyHandler) GetPublicProperty(c *gin.Context) {
	id := c.Param("id")

	property, err := h.propertyService.GetPublicProperty(c.Request.Context(), id)
	if err != nil {
		if err == repositories.ErrNotFound || err.Error() == "property is not public" || err.Error() == "property is not available" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Public property not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get public property",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    property,
	})
}

// GetPublicPropertyBySlug retrieves a single public property by slug (cross-tenant)
// @Summary Get public property by slug (cross-tenant)
// @Description Get a single public and available property by slug from any tenant
// @Tags public-properties
// @Produce json
// @Param slug path string true "Property slug"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/public/properties/slug/{slug} [get]
func (h *PublicPropertyHandler) GetPublicPropertyBySlug(c *gin.Context) {
	slug := c.Param("slug")

	property, err := h.propertyService.GetPublicPropertyBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Public property not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get public property by slug",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    property,
	})
}
