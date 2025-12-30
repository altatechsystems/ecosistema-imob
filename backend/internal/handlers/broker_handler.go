package handlers

import (
	"net/http"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/altatech/ecosistema-imob/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

// BrokerHandler handles broker-related HTTP requests
type BrokerHandler struct {
	brokerService  *services.BrokerService
	storageService *storage.StorageService
}

// NewBrokerHandler creates a new broker handler
func NewBrokerHandler(brokerService *services.BrokerService, storageService *storage.StorageService) *BrokerHandler {
	return &BrokerHandler{
		brokerService:  brokerService,
		storageService: storageService,
	}
}

// RegisterRoutes registers broker routes (tenant-scoped)
func (h *BrokerHandler) RegisterRoutes(router *gin.RouterGroup) {
	brokers := router.Group("/brokers")
	{
		brokers.POST("", h.CreateBroker)
		brokers.GET("/:id", h.GetBroker)
		brokers.PUT("/:id", h.UpdateBroker)
		brokers.DELETE("/:id", h.DeleteBroker)
		brokers.GET("", h.ListBrokers)
		brokers.POST("/:id/activate", h.ActivateBroker)
		brokers.POST("/:id/deactivate", h.DeactivateBroker)
		brokers.POST("/:id/photo", h.UploadPhoto)
		brokers.DELETE("/:id/photo", h.DeletePhoto)
	}
}

// CreateBroker creates a new broker
// @Summary Create a new broker
// @Description Create a new broker for a tenant
// @Tags brokers
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param broker body models.Broker true "Broker data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers [post]
func (h *BrokerHandler) CreateBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	var broker models.Broker
	if err := c.ShouldBindJSON(&broker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Set tenant ID from path parameter
	broker.TenantID = tenantID

	if err := h.brokerService.CreateBroker(c.Request.Context(), &broker); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    broker,
	})
}

// GetBroker retrieves a broker by ID
// @Summary Get broker by ID
// @Description Get broker details by ID
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers/{id} [get]
func (h *BrokerHandler) GetBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	broker, err := h.brokerService.GetBroker(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
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
		"data":    broker,
	})
}

// UpdateBroker updates a broker
// @Summary Update broker
// @Description Update broker information
// @Tags brokers
// @Accept json
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Param updates body map[string]interface{} true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers/{id} [put]
func (h *BrokerHandler) UpdateBroker(c *gin.Context) {
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

	if err := h.brokerService.UpdateBroker(c.Request.Context(), tenantID, id, updates); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
			})
			return
		}
		// Log detailed error
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "broker updated successfully"},
	})
}

// DeleteBroker deletes a broker
// @Summary Delete broker
// @Description Delete a broker
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers/{id} [delete]
func (h *BrokerHandler) DeleteBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.brokerService.DeleteBroker(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
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
		"data":    gin.H{"message": "broker deleted successfully"},
	})
}

// ListBrokers lists all brokers for a tenant
// @Summary List brokers
// @Description List all brokers for a tenant with pagination
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param limit query int false "Limit" default(50)
// @Param order_by query string false "Order by field" default(created_at)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers [get]
func (h *BrokerHandler) ListBrokers(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	// Parse pagination options
	opts := parsePaginationOptions(c)

	brokers, err := h.brokerService.ListBrokers(c.Request.Context(), tenantID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    brokers,
		"count":   len(brokers),
	})
}

// ActivateBroker activates a broker
// @Summary Activate broker
// @Description Activate a broker
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers/{id}/activate [post]
func (h *BrokerHandler) ActivateBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.brokerService.ActivateBroker(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
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
		"data":    gin.H{"message": "broker activated successfully"},
	})
}

// DeactivateBroker deactivates a broker
// @Summary Deactivate broker
// @Description Deactivate a broker
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers/{id}/deactivate [post]
func (h *BrokerHandler) DeactivateBroker(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	if err := h.brokerService.DeactivateBroker(c.Request.Context(), tenantID, id); err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
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
		"data":    gin.H{"message": "broker deactivated successfully"},
	})
}

// UploadPhoto handles broker photo upload
// @Summary Upload broker photo
// @Description Upload a profile photo for a broker
// @Tags brokers
// @Accept multipart/form-data
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Param file formData file true "Photo file (JPEG, PNG, or WebP, max 5MB)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers/{id}/photo [post]
func (h *BrokerHandler) UploadPhoto(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	// Validate broker exists
	broker, err := h.brokerService.GetBroker(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "file is required",
		})
		return
	}
	defer file.Close()

	// Validate file size (max 5MB for profile photos)
	const maxSize = 5 * 1024 * 1024 // 5MB
	if header.Size > maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"success": false,
			"error":   "file size exceeds 5MB limit",
		})
		return
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if err := h.storageService.ValidateContentType(contentType); err != nil {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Delete old photo if exists
	if broker.PhotoURL != "" {
		_ = h.storageService.DeleteBrokerPhoto(c.Request.Context(), tenantID, id)
	}

	// Upload new photo
	photoURL, err := h.storageService.UploadBrokerPhoto(c.Request.Context(), tenantID, id, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Update broker with new photo URL
	updates := map[string]interface{}{
		"photo_url": photoURL,
	}
	if err := h.brokerService.UpdateBroker(c.Request.Context(), tenantID, id, updates); err != nil {
		// Try to delete uploaded photo if update fails
		_ = h.storageService.DeleteBrokerPhoto(c.Request.Context(), tenantID, id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"photo_url": photoURL,
			"message":   "photo uploaded successfully",
		},
	})
}

// DeletePhoto handles broker photo deletion
// @Summary Delete broker photo
// @Description Delete a broker's profile photo
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/{tenant_id}/brokers/{id}/photo [delete]
func (h *BrokerHandler) DeletePhoto(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	// Validate broker exists
	broker, err := h.brokerService.GetBroker(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Delete photo from storage
	if broker.PhotoURL != "" {
		if err := h.storageService.DeleteBrokerPhoto(c.Request.Context(), tenantID, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
	}

	// Update broker to remove photo URL
	updates := map[string]interface{}{
		"photo_url": "",
	}
	if err := h.brokerService.UpdateBroker(c.Request.Context(), tenantID, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "photo deleted successfully"},
	})
}

// GetBrokerPublicProfile retrieves a broker's public profile
// @Summary Get broker public profile
// @Description Get broker's public profile (sanitized data for public viewing)
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/{tenant_id}/brokers/{id}/public [get]
func (h *BrokerHandler) GetBrokerPublicProfile(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	broker, err := h.brokerService.GetBroker(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Convert to public broker (sanitized data)
	publicBroker := &models.BrokerPublic{
		ID:            broker.ID,
		Name:          broker.Name,
		Email:         broker.Email,
		Phone:         broker.Phone,
		CRECI:         broker.CRECI,
		PhotoURL:      broker.PhotoURL,
		Bio:           broker.Bio,
		Specialties:   broker.Specialties,
		Languages:     broker.Languages,
		Experience:    broker.Experience,
		Company:       broker.Company,
		Website:       broker.Website,
		TotalSales:    broker.TotalSales,
		TotalListings: broker.TotalListings,
		AveragePrice:  broker.AveragePrice,
		Rating:        broker.Rating,
		ReviewCount:   broker.ReviewCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    publicBroker,
	})
}

// GetBrokerProperties retrieves properties associated with a broker
// @Summary Get broker properties
// @Description Get list of properties where the broker is the captador
// @Tags brokers
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Param id path string true "Broker ID"
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/{tenant_id}/brokers/{id}/properties [get]
func (h *BrokerHandler) GetBrokerProperties(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	id := c.Param("id")

	// Verify broker exists
	_, err := h.brokerService.GetBroker(c.Request.Context(), tenantID, id)
	if err != nil {
		if err == repositories.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "broker not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get properties for this broker
	properties, err := h.brokerService.GetBrokerProperties(c.Request.Context(), tenantID, id, parsePaginationOptions(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    properties,
		"count":   len(properties),
	})
}

