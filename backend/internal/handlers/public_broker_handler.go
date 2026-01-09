package handlers

import (
	"net/http"
	"strconv"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PublicBrokerHandler handles public broker HTTP requests (cross-tenant)
// This handler is used by the public portal agregador to access broker info without tenant_id
type PublicBrokerHandler struct {
	brokerService *services.BrokerService
}

// NewPublicBrokerHandler creates a new public broker handler
func NewPublicBrokerHandler(brokerService *services.BrokerService) *PublicBrokerHandler {
	return &PublicBrokerHandler{
		brokerService: brokerService,
	}
}

// GetPublicBrokerProfile retrieves a broker's public profile (cross-tenant)
// @Summary Get broker public profile (cross-tenant)
// @Description Get a broker's public profile from any tenant
// @Tags public-brokers
// @Produce json
// @Param id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/public/brokers/{id}/profile [get]
func (h *PublicBrokerHandler) GetPublicBrokerProfile(c *gin.Context) {
	id := c.Param("id")

	// Find broker across all tenants
	broker, _, err := h.brokerService.GetBrokerFromAnyTenant(c.Request.Context(), id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Broker not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get broker profile",
			"details": err.Error(),
		})
		return
	}

	// Convert to BrokerPublic model (sanitized for public display)
	publicBroker := &models.BrokerPublic{
		ID:             broker.ID,
		Name:           broker.Name,
		Email:          broker.Email,
		Phone:          broker.Phone,
		CRECI:          broker.CRECI,
		PhotoURL:       broker.PhotoURL,
		Bio:            broker.Bio,
		Specialties:    broker.Specialties,
		Languages:      broker.Languages,
		Experience:     broker.Experience,
		Company:        broker.Company,
		Website:        broker.Website,
		SocialMedia:    broker.SocialMedia,
		TotalSales:     broker.TotalSales,
		TotalListings:  broker.TotalListings,
		AveragePrice:   broker.AveragePrice,
		Rating:         broker.Rating,
		ReviewCount:    broker.ReviewCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    publicBroker,
	})
}

// GetPublicBrokerProperties retrieves a broker's public properties (cross-tenant)
// @Summary Get broker's public properties (cross-tenant)
// @Description Get a list of public and available properties for a broker
// @Tags public-brokers
// @Produce json
// @Param id path string true "Broker ID"
// @Param limit query int false "Maximum number of results" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/public/brokers/{id}/properties [get]
func (h *PublicBrokerHandler) GetPublicBrokerProperties(c *gin.Context) {
	id := c.Param("id")

	// Parse limit from query params
	limit := 20 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 { // cap at 100
				limit = 100
			}
		}
	}

	// Get broker's public properties
	properties, err := h.brokerService.GetPublicBrokerProperties(c.Request.Context(), id, limit)
	if err != nil {
		if err.Error() == "failed to find broker: "+repositories.ErrNotFound.Error() {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Broker not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get broker properties",
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
