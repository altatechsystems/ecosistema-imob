package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PublicLeadHandler handles public lead HTTP requests (cross-tenant)
// This handler is used by the public portal agregador to create leads without tenant_id in URL
type PublicLeadHandler struct {
	leadService     *services.LeadService
	propertyService *services.PropertyService
}

// NewPublicLeadHandler creates a new public lead handler
func NewPublicLeadHandler(leadService *services.LeadService, propertyService *services.PropertyService) *PublicLeadHandler {
	return &PublicLeadHandler{
		leadService:     leadService,
		propertyService: propertyService,
	}
}

// CreatePublicWhatsAppLead creates a WhatsApp lead for a public property (cross-tenant)
// @Summary Create WhatsApp lead (cross-tenant)
// @Description Create a WhatsApp lead for a public property. Tenant is resolved from property.
// @Tags public-leads
// @Accept json
// @Produce json
// @Param property_id path string true "Property ID"
// @Param lead body CreateWhatsAppLeadRequest true "WhatsApp lead data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/public/properties/{property_id}/leads/whatsapp [post]
func (h *PublicLeadHandler) CreatePublicWhatsAppLead(c *gin.Context) {
	propertyID := c.Param("property_id")

	var req CreateWhatsAppLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get property to resolve tenant_id and verify it's public
	property, err := h.propertyService.GetPublicProperty(c.Request.Context(), propertyID)
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
			"error":   "Failed to get property",
			"details": err.Error(),
		})
		return
	}

	// Get client IP for LGPD consent tracking
	clientIP := c.ClientIP()

	// Use provided name/phone or placeholders
	leadName := req.Name
	if leadName == "" {
		leadName = "Lead via WhatsApp"
	}

	leadPhone := req.Phone
	if leadPhone == "" {
		leadPhone = "WhatsApp" // Placeholder - will be updated when customer responds
	}

	// Create lead with WhatsApp channel
	// Use tenant_id from property
	lead := &models.Lead{
		TenantID:     property.TenantID,
		PropertyID:   propertyID,
		Channel:      models.LeadChannelWhatsApp,
		Name:         leadName,
		Phone:        leadPhone,
		ConsentGiven: true, // Implícito ao clicar no botão WhatsApp
		ConsentText:  "Concordo com a Política de Privacidade e autorizo o uso dos meus dados para contato sobre este imóvel.",
		ConsentIP:    clientIP,
		UTMSource:    req.UTMSource,
		UTMCampaign:  req.UTMCampaign,
		UTMMedium:    req.UTMMedium,
		Referrer:     req.Referrer,
	}

	// Create lead (validates property exists)
	if err := h.leadService.CreateLead(c.Request.Context(), lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Generate WhatsApp URL and message
	whatsappData, err := h.leadService.GenerateWhatsAppURL(c.Request.Context(), property.TenantID, propertyID, lead.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":      true,
		"lead_id":      lead.ID,
		"whatsapp_url": whatsappData.URL,
		"message":      whatsappData.Message,
	})
}

// CreatePublicFormLead creates a form lead for a public property (cross-tenant)
// @Summary Create form lead (cross-tenant)
// @Description Create a form lead for a public property. Tenant is resolved from property. LGPD consent required.
// @Tags public-leads
// @Accept json
// @Produce json
// @Param property_id path string true "Property ID"
// @Param lead body CreateFormLeadRequest true "Form lead data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/public/properties/{property_id}/leads/form [post]
func (h *PublicLeadHandler) CreatePublicFormLead(c *gin.Context) {
	propertyID := c.Param("property_id")

	var req CreateFormLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// LGPD validation: consent is mandatory
	if !req.ConsentGiven {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "consent_given must be true (LGPD compliance)",
		})
		return
	}

	// Get property to resolve tenant_id and verify it's public
	property, err := h.propertyService.GetPublicProperty(c.Request.Context(), propertyID)
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
			"error":   "Failed to get property",
			"details": err.Error(),
		})
		return
	}

	// Get client IP for LGPD consent tracking
	clientIP := c.ClientIP()

	// Create lead with form channel
	// Use tenant_id from property
	lead := &models.Lead{
		TenantID:     property.TenantID,
		PropertyID:   propertyID,
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Message:      req.Message,
		Channel:      models.LeadChannelForm,
		ConsentGiven: req.ConsentGiven,
		ConsentText:  req.ConsentText,
		ConsentIP:    clientIP,
		UTMSource:    req.UTMSource,
		UTMCampaign:  req.UTMCampaign,
		UTMMedium:    req.UTMMedium,
		Referrer:     req.Referrer,
	}

	// Create lead (validates property exists and contact methods)
	if err := h.leadService.CreateLead(c.Request.Context(), lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"lead_id": lead.ID,
		"message": "Lead criado com sucesso. O corretor entrará em contato em breve.",
	})
}
