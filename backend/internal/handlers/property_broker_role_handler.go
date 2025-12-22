package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PropertyBrokerRoleHandler handles co-brokerage-related HTTP requests
type PropertyBrokerRoleHandler struct {
	roleService *services.PropertyBrokerRoleService
}

// NewPropertyBrokerRoleHandler creates a new property broker role handler
func NewPropertyBrokerRoleHandler(roleService *services.PropertyBrokerRoleService) *PropertyBrokerRoleHandler {
	return &PropertyBrokerRoleHandler{
		roleService: roleService,
	}
}

// RegisterRoutes registers property broker role routes (tenant-scoped)
func (h *PropertyBrokerRoleHandler) RegisterRoutes(router *gin.RouterGroup) {
	propertyBrokers := router.Group("/:tenant_id/properties/:property_id/brokers")
	{
		propertyBrokers.POST("", h.AssignBroker)
		propertyBrokers.DELETE("/:broker_id", h.RemoveBroker)
		propertyBrokers.PUT("/:broker_id", h.UpdateRole)
		propertyBrokers.GET("", h.GetPropertyBrokers)
		propertyBrokers.POST("/:broker_id/set-primary", h.SetPrimaryBroker)
	}
}

// AssignBroker assigns a broker to a property
// @Summary Assign broker to property
// @Description Assign a broker to a property with a specific role
// @Tags co-brokerage
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param property_id path string true "Property ID"
// @Param role body models.PropertyBrokerRole true "Broker role data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{property_id}/brokers [post]
func (h *PropertyBrokerRoleHandler) AssignBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("property_id")

	var role models.PropertyBrokerRole
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Set tenant ID and property ID from path parameters
	role.TenantID = tenantID
	role.PropertyID = propertyID

	if err := h.roleService.AssignBrokerToProperty(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    role,
	})
}

// RemoveBroker removes a broker from a property
// @Summary Remove broker from property
// @Description Remove a broker's role from a property
// @Tags co-brokerage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param property_id path string true "Property ID"
// @Param broker_id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{property_id}/brokers/{broker_id} [delete]
func (h *PropertyBrokerRoleHandler) RemoveBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("property_id")
	brokerID := c.Param("broker_id")

	// First, get all roles for the property to find the matching role ID
	roles, err := h.roleService.GetPropertyBrokers(c.Request.Context(), tenantID, propertyID, repositories.DefaultPaginationOptions())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Find the role ID for this broker
	var roleID string
	for _, role := range roles {
		if role.BrokerID == brokerID {
			roleID = role.ID
			break
		}
	}

	if roleID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "broker not assigned to this property",
		})
		return
	}

	if err := h.roleService.RemoveBrokerFromProperty(c.Request.Context(), tenantID, roleID); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker role not found",
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
		"data":    gin.H{"message": "broker removed from property successfully"},
	})
}

// UpdateRole updates a broker's role for a property
// @Summary Update broker role
// @Description Update a broker's role for a property
// @Tags co-brokerage
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param property_id path string true "Property ID"
// @Param broker_id path string true "Broker ID"
// @Param updates body map[string]interface{} true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{property_id}/brokers/{broker_id} [put]
func (h *PropertyBrokerRoleHandler) UpdateRole(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("property_id")
	brokerID := c.Param("broker_id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// First, get all roles for the property to find the matching role ID
	roles, err := h.roleService.GetPropertyBrokers(c.Request.Context(), tenantID, propertyID, repositories.DefaultPaginationOptions())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Find the role ID for this broker
	var roleID string
	for _, role := range roles {
		if role.BrokerID == brokerID {
			roleID = role.ID
			break
		}
	}

	if roleID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "broker not assigned to this property",
		})
		return
	}

	if err := h.roleService.UpdateRole(c.Request.Context(), tenantID, roleID, updates); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker role not found",
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
		"data":    gin.H{"message": "broker role updated successfully"},
	})
}

// GetPropertyBrokers retrieves all brokers for a property
// @Summary Get property brokers
// @Description Get all brokers assigned to a property with their roles
// @Tags co-brokerage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param property_id path string true "Property ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(created_at)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{property_id}/brokers [get]
func (h *PropertyBrokerRoleHandler) GetPropertyBrokers(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("property_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	roles, err := h.roleService.GetPropertyBrokers(c.Request.Context(), tenantID, propertyID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    roles,
		"count":   len(roles),
	})
}

// SetPrimaryBroker sets a broker as the primary for lead routing
// @Summary Set primary broker
// @Description Set a broker as the primary for lead routing
// @Tags co-brokerage
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param property_id path string true "Property ID"
// @Param broker_id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/properties/{property_id}/brokers/{broker_id}/set-primary [post]
func (h *PropertyBrokerRoleHandler) SetPrimaryBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	propertyID := c.Param("property_id")
	brokerID := c.Param("broker_id")

	// First, get all roles for the property to find the matching role ID
	roles, err := h.roleService.GetPropertyBrokers(c.Request.Context(), tenantID, propertyID, repositories.DefaultPaginationOptions())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Find the role ID for this broker
	var roleID string
	for _, role := range roles {
		if role.BrokerID == brokerID {
			roleID = role.ID
			break
		}
	}

	if roleID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "broker not assigned to this property",
		})
		return
	}

	if err := h.roleService.SetPrimaryBroker(c.Request.Context(), tenantID, roleID); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker role not found",
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
		"data":    gin.H{"message": "primary broker set successfully"},
	})
}

