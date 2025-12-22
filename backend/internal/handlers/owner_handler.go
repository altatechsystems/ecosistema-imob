package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// OwnerHandler handles owner-related HTTP requests
type OwnerHandler struct {
	ownerService *services.OwnerService
}

// NewOwnerHandler creates a new owner handler
func NewOwnerHandler(ownerService *services.OwnerService) *OwnerHandler {
	return &OwnerHandler{
		ownerService: ownerService,
	}
}

// RegisterRoutes registers owner routes (tenant-scoped)
func (h *OwnerHandler) RegisterRoutes(router *gin.RouterGroup) {
	owners := router.Group("/:tenant_id/owners")
	{
		owners.POST("", h.CreateOwner)
		owners.GET("/:id", h.GetOwner)
		owners.PUT("/:id", h.UpdateOwner)
		owners.DELETE("/:id", h.DeleteOwner)
		owners.GET("", h.ListOwners)
		owners.POST("/:id/revoke-consent", h.RevokeConsent)
		owners.POST("/:id/anonymize", h.AnonymizeOwner)
	}
}

// CreateOwner creates a new owner
// @Summary Create a new owner
// @Description Create a new property owner for a tenant
// @Tags owners
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param owner body models.Owner true "Owner data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/owners [post]
func (h *OwnerHandler) CreateOwner(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	var owner models.Owner
	if err := c.ShouldBindJSON(&owner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Set tenant ID from path parameter
	owner.TenantID = tenantID

	if err := h.ownerService.CreateOwner(c.Request.Context(), &owner); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    owner,
	})
}

// GetOwner retrieves an owner by ID
// @Summary Get owner by ID
// @Description Get owner details by ID
// @Tags owners
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Owner ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/owners/{id} [get]
func (h *OwnerHandler) GetOwner(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	owner, err := h.ownerService.GetOwner(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "owner not found",
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
		"data":    owner,
	})
}

// UpdateOwner updates an owner
// @Summary Update owner
// @Description Update owner information
// @Tags owners
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Owner ID"
// @Param updates body map[string]interface{} true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/owners/{id} [put]
func (h *OwnerHandler) UpdateOwner(c *gin.Context) {
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

	if err := h.ownerService.UpdateOwner(c.Request.Context(), tenantID, id, updates); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "owner not found",
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
		"data":    gin.H{"message": "owner updated successfully"},
	})
}

// DeleteOwner deletes an owner
// @Summary Delete owner
// @Description Delete an owner
// @Tags owners
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Owner ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/owners/{id} [delete]
func (h *OwnerHandler) DeleteOwner(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.ownerService.DeleteOwner(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "owner not found",
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
		"data":    gin.H{"message": "owner deleted successfully"},
	})
}

// ListOwners lists all owners for a tenant
// @Summary List owners
// @Description List all owners for a tenant with pagination
// @Tags owners
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(created_at)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/owners [get]
func (h *OwnerHandler) ListOwners(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	owners, err := h.ownerService.ListOwners(c.Request.Context(), tenantID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    owners,
		"count":   len(owners),
	})
}

// RevokeConsent revokes owner consent (LGPD)
// @Summary Revoke owner consent
// @Description Revoke LGPD consent for an owner
// @Tags owners
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Owner ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/owners/{id}/revoke-consent [post]
func (h *OwnerHandler) RevokeConsent(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.ownerService.RevokeConsent(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "owner not found",
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

// AnonymizeOwnerRequest represents the request body for anonymizing an owner
type AnonymizeOwnerRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// AnonymizeOwner anonymizes owner data (LGPD)
// @Summary Anonymize owner
// @Description Anonymize owner data for LGPD compliance
// @Tags owners
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Owner ID"
// @Param body body AnonymizeOwnerRequest true "Anonymization reason"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/owners/{id}/anonymize [post]
func (h *OwnerHandler) AnonymizeOwner(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	var req AnonymizeOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.ownerService.AnonymizeOwner(c.Request.Context(), tenantID, id, req.Reason); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "owner not found",
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
		"data":    gin.H{"message": "owner anonymized successfully"},
	})
}

