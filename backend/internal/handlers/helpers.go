package handlers

import (
	"strconv"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/gin-gonic/gin"
)

// parsePaginationOptions extracts pagination parameters from query string
func parsePaginationOptions(c *gin.Context) repositories.PaginationOptions {
	opts := repositories.DefaultPaginationOptions()

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			opts.Offset = offset
		}
	}

	if orderBy := c.Query("order_by"); orderBy != "" {
		opts.OrderBy = orderBy
	}

	if startAfter := c.Query("start_after"); startAfter != "" {
		opts.StartAfter = startAfter
	}

	return opts
}

// UpdateStatusRequest is a common request structure for status updates
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// LeadUpdateStatusRequest represents the request body for updating lead status
type LeadUpdateStatusRequest struct {
	Status models.LeadStatus `json:"status" binding:"required"`
}

// PropertyUpdateStatusRequest represents the request body for updating property status
type PropertyUpdateStatusRequest struct {
	Status models.PropertyStatus `json:"status" binding:"required"`
}
