package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
)

// ScheduledConfirmationHandler handles scheduled confirmation endpoints
type ScheduledConfirmationHandler struct {
	scheduler *services.MonthlyConfirmationScheduler
}

// NewScheduledConfirmationHandler creates a new scheduled confirmation handler
func NewScheduledConfirmationHandler(scheduler *services.MonthlyConfirmationScheduler) *ScheduledConfirmationHandler {
	return &ScheduledConfirmationHandler{
		scheduler: scheduler,
	}
}

// ScheduleMonthlyConfirmationsRequest represents the request body
type ScheduleMonthlyConfirmationsRequest struct {
	ScheduledFor string `json:"scheduled_for"` // ISO 8601 format: "2025-02-01T09:00:00Z"
	DryRun       bool   `json:"dry_run"`       // If true, only returns count
}

// ScheduleMonthlyConfirmations schedules monthly confirmations for all properties
// POST /api/v1/admin/:tenant_id/scheduled-confirmations/schedule
func (h *ScheduledConfirmationHandler) ScheduleMonthlyConfirmations(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	var req ScheduleMonthlyConfirmationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Parse scheduled_for date if provided
	var scheduledFor time.Time
	if req.ScheduledFor != "" {
		var err error
		scheduledFor, err = time.Parse(time.RFC3339, req.ScheduledFor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid scheduled_for format. Use ISO 8601 (e.g., 2025-02-01T09:00:00Z)",
			})
			return
		}
	}

	// Call scheduler service
	result, err := h.scheduler.ScheduleMonthlyConfirmations(c.Request.Context(), services.ScheduleMonthlyConfirmationsRequest{
		TenantID:     tenantID,
		ScheduledFor: scheduledFor,
		DryRun:       req.DryRun,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ProcessPendingConfirmations processes pending confirmations for today
// POST /api/v1/admin/:tenant_id/scheduled-confirmations/process
func (h *ScheduledConfirmationHandler) ProcessPendingConfirmations(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	err := h.scheduler.ProcessPendingConfirmations(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pending confirmations processed successfully",
	})
}

// GetScheduledConfirmations gets all scheduled confirmations for a tenant
// GET /api/v1/admin/:tenant_id/scheduled-confirmations
func (h *ScheduledConfirmationHandler) GetScheduledConfirmations(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	statusParam := c.Query("status")

	var status *models.ScheduledConfirmationStatus
	if statusParam != "" {
		s := models.ScheduledConfirmationStatus(statusParam)
		status = &s
	}

	confirmations, err := h.scheduler.GetScheduledConfirmationsForTenant(c.Request.Context(), tenantID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    confirmations,
	})
}

// GetBrokerScheduledConfirmations gets scheduled confirmations for a broker
// GET /api/v1/admin/:tenant_id/scheduled-confirmations/broker/:broker_id
func (h *ScheduledConfirmationHandler) GetBrokerScheduledConfirmations(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	brokerID := c.Param("broker_id")
	statusParam := c.Query("status")

	var status *models.ScheduledConfirmationStatus
	if statusParam != "" {
		s := models.ScheduledConfirmationStatus(statusParam)
		status = &s
	}

	confirmations, err := h.scheduler.GetScheduledConfirmationsForBroker(c.Request.Context(), tenantID, brokerID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    confirmations,
	})
}

// GetConfirmationMetrics gets metrics for scheduled confirmations
// GET /api/v1/admin/:tenant_id/scheduled-confirmations/metrics
func (h *ScheduledConfirmationHandler) GetConfirmationMetrics(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	metrics, err := h.scheduler.GetConfirmationMetrics(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}
