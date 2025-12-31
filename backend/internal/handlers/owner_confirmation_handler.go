package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// OwnerConfirmationHandler handles public owner confirmation requests
type OwnerConfirmationHandler struct {
	ownerConfirmationService *services.OwnerConfirmationService
}

// NewOwnerConfirmationHandler creates a new owner confirmation handler
func NewOwnerConfirmationHandler(ownerConfirmationService *services.OwnerConfirmationService) *OwnerConfirmationHandler {
	return &OwnerConfirmationHandler{
		ownerConfirmationService: ownerConfirmationService,
	}
}

// RegisterPublicRoutes registers PUBLIC routes (no auth required)
// These routes should be registered at the root level, not under /api/{tenant_id}
func (h *OwnerConfirmationHandler) RegisterPublicRoutes(router *gin.Engine) {
	// Public confirmation page - GET /confirmar/{token}
	router.GET("/confirmar/:token", h.GetConfirmationPage)

	// Public confirmation submission - POST /api/v1/owner-confirmations/{token}/submit
	router.POST("/api/v1/owner-confirmations/:token/submit", h.SubmitConfirmation)
}

// GetConfirmationPage validates token and returns minimal property info
// GET /confirmar/{token}
// @Summary Get owner confirmation page data
// @Description Validates token and returns minimal property information for owner confirmation (PROMPT 08)
// @Tags owner-confirmation
// @Produce json
// @Param token path string true "Confirmation Token"
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} services.GetConfirmationPageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /confirmar/{token} [get]
func (h *OwnerConfirmationHandler) GetConfirmationPage(c *gin.Context) {
	token := c.Param("token")
	tenantID := c.Query("tenant_id")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "token is required",
		})
		return
	}

	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "tenant_id is required",
		})
		return
	}

	// Validate token and get property info
	response, err := h.ownerConfirmationService.ValidateTokenAndGetPropertyInfo(c.Request.Context(), tenantID, token)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "token not found or expired",
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
		"data":    response,
	})
}

// SubmitConfirmationRequest represents the owner's confirmation action
type SubmitConfirmationRequest struct {
	Action      models.ConfirmationAction `json:"action" binding:"required"`      // confirm_available, confirm_unavailable, confirm_price
	PriceAmount *float64                  `json:"price_amount,omitempty"`         // required if action=confirm_price
}

// SubmitConfirmation handles the owner's confirmation submission
// POST /api/v1/owner-confirmations/{token}/submit
// @Summary Submit owner confirmation
// @Description Owner submits confirmation of property status or price (PROMPT 08)
// @Tags owner-confirmation
// @Accept json
// @Produce json
// @Param token path string true "Confirmation Token"
// @Param tenant_id query string true "Tenant ID"
// @Param body body SubmitConfirmationRequest true "Confirmation action"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/owner-confirmations/{token}/submit [post]
func (h *OwnerConfirmationHandler) SubmitConfirmation(c *gin.Context) {
	token := c.Param("token")
	tenantID := c.Query("tenant_id")

	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "tenant_id is required",
		})
		return
	}

	var req SubmitConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate price_amount is provided if action is confirm_price
	if req.Action == models.ConfirmationActionPrice && req.PriceAmount == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "price_amount is required when action is confirm_price",
		})
		return
	}

	// Submit confirmation
	err := h.ownerConfirmationService.SubmitOwnerConfirmation(
		c.Request.Context(),
		tenantID,
		token,
		req.Action,
		req.PriceAmount,
	)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "token not found, expired, or already used",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Obrigado! Informação atualizada com sucesso.",
	})
}
