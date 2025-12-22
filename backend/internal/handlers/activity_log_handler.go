package handlers

import (
	"net/http"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ActivityLogHandler handles activity log-related HTTP requests
type ActivityLogHandler struct {
	activityLogService *services.ActivityLogService
}

// NewActivityLogHandler creates a new activity log handler
func NewActivityLogHandler(activityLogService *services.ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{
		activityLogService: activityLogService,
	}
}

// RegisterRoutes registers activity log routes (tenant-scoped)
func (h *ActivityLogHandler) RegisterRoutes(router *gin.RouterGroup) {
	activityLogs := router.Group("/:tenant_id/activity-logs")
	{
		activityLogs.GET("", h.GetActivityLogs)
		activityLogs.GET("/:id", h.GetActivityLog)
	}

	// Timeline endpoints
	properties := router.Group("/:tenant_id/properties/:property_id")
	{
		properties.GET("/timeline", h.GetPropertyTimeline)
	}

	leads := router.Group("/:tenant_id/leads/:lead_id")
	{
		leads.GET("/timeline", h.GetLeadTimeline)
	}
}

// GetActivityLogs retrieves activity logs with filters
// @Summary Get activity logs
// @Description Get activity logs for a tenant with filters and pagination
// @Tags activity-logs
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(timestamp)
// @Param event_type query string false "Event type filter"
// @Param actor_type query string false "Actor type filter"
// @Param actor_id query string false "Actor ID filter"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/activity-logs [get]
func (h *ActivityLogHandler) GetActivityLogs(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	// Parse filters
	filters := &repositories.ActivityLogFilters{}

	if eventType := c.Query("event_type"); eventType != "" {
		filters.EventType = eventType
	}

	if actorType := c.Query("actor_type"); actorType != "" {
		actorTypeVal := models.ActorType(actorType)
		filters.ActorType = &actorTypeVal
	}

	if actorID := c.Query("actor_id"); actorID != "" {
		filters.ActorID = actorID
	}

	// Parse date range
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid start_date format, use RFC3339",
			})
			return
		}
		filters.StartDate = &startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid end_date format, use RFC3339",
			})
			return
		}
		filters.EndDate = &endDate
	}

	logs, err := h.activityLogService.GetActivityLogs(c.Request.Context(), tenantID, filters, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"count":   len(logs),
	})
}

// GetActivityLog retrieves a single activity log by ID
// @Summary Get activity log by ID
// @Description Get activity log details by ID
// @Tags activity-logs
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Activity Log ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/activity-logs/{id} [get]
func (h *ActivityLogHandler) GetActivityLog(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	// For now, we'll get all logs and filter by ID
	// In a production system, you'd have a dedicated GetByID method
	filters := &repositories.ActivityLogFilters{}
	opts := repositories.PaginationOptions{
		Limit:     1,
		OrderBy:   "timestamp",
		Direction: 2, // Desc
	}

	logs, err := h.activityLogService.GetActivityLogs(c.Request.Context(), tenantID, filters, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Find the log with matching ID
	var foundLog interface{}
	for _, log := range logs {
		if log.ID == id {
			foundLog = log
			break
		}
	}

	if foundLog == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "activity log not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    foundLog,
	})
}

// GetPropertyTimeline retrieves the activity timeline for a property
// @Summary Get property timeline
// @Description Get activity timeline for a property
// @Tags activity-logs
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param property_id path string true "Property ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(timestamp)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{property_id}/timeline [get]
func (h *ActivityLogHandler) GetPropertyTimeline(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("property_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	logs, err := h.activityLogService.GetPropertyTimeline(c.Request.Context(), tenantID, propertyID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"count":   len(logs),
	})
}

// GetLeadTimeline retrieves the activity timeline for a lead
// @Summary Get lead timeline
// @Description Get activity timeline for a lead
// @Tags activity-logs
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param lead_id path string true "Lead ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(timestamp)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/leads/{lead_id}/timeline [get]
func (h *ActivityLogHandler) GetLeadTimeline(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	leadID := c.Param("lead_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	logs, err := h.activityLogService.GetLeadTimeline(c.Request.Context(), tenantID, leadID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"count":   len(logs),
	})
}

