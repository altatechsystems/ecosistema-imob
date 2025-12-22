package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// LeadHandler handles lead-related HTTP requests
type LeadHandler struct {
	leadService *services.LeadService
}

// NewLeadHandler creates a new lead handler
func NewLeadHandler(leadService *services.LeadService) *LeadHandler {
	return &LeadHandler{
		leadService: leadService,
	}
}

// RegisterRoutes registers lead routes (tenant-scoped)
func (h *LeadHandler) RegisterRoutes(router *gin.RouterGroup) {
	leads := router.Group("/:tenant_id/leads")
	{
		leads.POST("", h.CreateLead)
		leads.GET("/:id", h.GetLead)
		leads.PUT("/:id", h.UpdateLead)
		leads.DELETE("/:id", h.DeleteLead)
		leads.GET("", h.ListLeads)
		leads.POST("/:id/status", h.UpdateStatus)
		leads.POST("/:id/assign", h.AssignToBroker)
		leads.POST("/:id/revoke-consent", h.RevokeConsent)
		leads.POST("/:id/anonymize", h.AnonymizeLead)
	}
}

// CreateLead creates a new lead
// @Summary Create a new lead
// @Description Create a new lead with LGPD compliance
// @Tags leads
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param lead body models.Lead true "Lead data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads [post]
func (h *LeadHandler) CreateLead(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	var lead models.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Set tenant ID from path parameter
	lead.TenantID = tenantID

	if err := h.leadService.CreateLead(c.Request.Context(), &lead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    lead,
	})
}

// GetLead retrieves a lead by ID
// @Summary Get lead by ID
// @Description Get lead details by ID
// @Tags leads
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Lead ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{id} [get]
func (h *LeadHandler) GetLead(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	lead, err := h.leadService.GetLead(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "lead not found",
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
		"data":    lead,
	})
}

// UpdateLead updates a lead
// @Summary Update lead
// @Description Update lead information
// @Tags leads
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Lead ID"
// @Param updates body map[string]interface{} true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{id} [put]
func (h *LeadHandler) UpdateLead(c *gin.Context) {
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

	if err := h.leadService.UpdateLead(c.Request.Context(), tenantID, id, updates); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "lead not found",
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
		"data":    gin.H{"message": "lead updated successfully"},
	})
}

// DeleteLead deletes a lead
// @Summary Delete lead
// @Description Delete a lead (prefer anonymization)
// @Tags leads
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Lead ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{id} [delete]
func (h *LeadHandler) DeleteLead(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.leadService.DeleteLead(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "lead not found",
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
		"data":    gin.H{"message": "lead deleted successfully"},
	})
}

// ListLeads lists all leads for a tenant
// @Summary List leads
// @Description List all leads for a tenant with filters and pagination
// @Tags leads
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(created_at)
// @Param property_id query string false "Property ID filter"
// @Param status query string false "Status filter"
// @Param channel query string false "Channel filter"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads [get]
func (h *LeadHandler) ListLeads(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	// Parse filters
	filters := &repositories.LeadFilters{}

	if propertyID := c.Query("property_id"); propertyID != "" {
		filters.PropertyID = propertyID
	}

	if status := c.Query("status"); status != "" {
		leadStatus := models.LeadStatus(status)
		filters.Status = &leadStatus
	}

	if channel := c.Query("channel"); channel != "" {
		leadChannel := models.LeadChannel(channel)
		filters.Channel = &leadChannel
	}

	leads, err := h.leadService.ListLeads(c.Request.Context(), tenantID, filters, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    leads,
		"count":   len(leads),
	})
}


// UpdateStatus updates the status of a lead
// @Summary Update lead status
// @Description Update the status of a lead
// @Tags leads
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Lead ID"
// @Param body body LeadUpdateStatusRequest true "Status update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{id}/status [post]
func (h *LeadHandler) UpdateStatus(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	var req LeadUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.leadService.UpdateStatus(c.Request.Context(), tenantID, id, req.Status); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "lead not found",
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
		"data":    gin.H{"message": "lead status updated successfully"},
	})
}

// AssignToBrokerRequest represents the request body for assigning a lead to a broker
type AssignToBrokerRequest struct {
	BrokerID string `json:"broker_id" binding:"required"`
}

// AssignToBroker assigns a lead to a specific broker
// @Summary Assign lead to broker
// @Description Assign a lead to a specific broker for manual routing
// @Tags leads
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Lead ID"
// @Param body body AssignToBrokerRequest true "Broker assignment"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{id}/assign [post]
func (h *LeadHandler) AssignToBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	var req AssignToBrokerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.leadService.AssignToBroker(c.Request.Context(), tenantID, id, req.BrokerID); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "lead not found",
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
		"data":    gin.H{"message": "lead assigned to broker successfully"},
	})
}

// RevokeConsent revokes lead consent (LGPD)
// @Summary Revoke lead consent
// @Description Revoke LGPD consent for a lead
// @Tags leads
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Lead ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{id}/revoke-consent [post]
func (h *LeadHandler) RevokeConsent(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.leadService.RevokeConsent(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "lead not found",
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
		"data":    gin.H{"message": "consent revoked successfully"},
	})
}

// AnonymizeLeadRequest represents the request body for anonymizing a lead
type AnonymizeLeadRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// AnonymizeLead anonymizes lead data (LGPD)
// @Summary Anonymize lead
// @Description Anonymize lead data for LGPD compliance
// @Tags leads
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Lead ID"
// @Param body body AnonymizeLeadRequest true "Anonymization reason"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{id}/anonymize [post]
func (h *LeadHandler) AnonymizeLead(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	var req AnonymizeLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.leadService.AnonymizeLead(c.Request.Context(), tenantID, id, req.Reason); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "lead not found",
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
		"data":    gin.H{"message": "lead anonymized successfully"},
	})
}

