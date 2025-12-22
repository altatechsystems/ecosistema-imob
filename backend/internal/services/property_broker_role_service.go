package services

import (
	"context"
	"fmt"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
)

// PropertyBrokerRoleService handles business logic for co-brokerage management
type PropertyBrokerRoleService struct {
	roleRepo        *repositories.PropertyBrokerRoleRepository
	propertyRepo    *repositories.PropertyRepository
	brokerRepo      *repositories.BrokerRepository
	tenantRepo      *repositories.TenantRepository
	activityLogRepo *repositories.ActivityLogRepository
}

// NewPropertyBrokerRoleService creates a new property broker role service
func NewPropertyBrokerRoleService(
	roleRepo *repositories.PropertyBrokerRoleRepository,
	propertyRepo *repositories.PropertyRepository,
	brokerRepo *repositories.BrokerRepository,
	tenantRepo *repositories.TenantRepository,
	activityLogRepo *repositories.ActivityLogRepository,
) *PropertyBrokerRoleService {
	return &PropertyBrokerRoleService{
		roleRepo:        roleRepo,
		propertyRepo:    propertyRepo,
		brokerRepo:      brokerRepo,
		tenantRepo:      tenantRepo,
		activityLogRepo: activityLogRepo,
	}
}

// AssignBrokerToProperty assigns a broker to a property with a specific role
func (s *PropertyBrokerRoleService) AssignBrokerToProperty(ctx context.Context, role *models.PropertyBrokerRole) error {
	// Validate required fields
	if role.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if role.PropertyID == "" {
		return fmt.Errorf("property_id is required")
	}
	if role.BrokerID == "" {
		return fmt.Errorf("broker_id is required")
	}
	if role.Role == "" {
		return fmt.Errorf("role is required")
	}

	// Validate tenant exists
	if _, err := s.tenantRepo.Get(ctx, role.TenantID); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Validate property exists
	if _, err := s.propertyRepo.Get(ctx, role.TenantID, role.PropertyID); err != nil {
		return fmt.Errorf("property not found: %w", err)
	}

	// Validate broker exists
	if _, err := s.brokerRepo.Get(ctx, role.TenantID, role.BrokerID); err != nil {
		return fmt.Errorf("broker not found: %w", err)
	}

	// Validate role type
	if err := s.validateRole(role.Role); err != nil {
		return err
	}

	// Business rule: Only one originating_broker per property
	if role.Role == models.BrokerPropertyRoleOriginating {
		existing, err := s.roleRepo.GetOriginatingBroker(ctx, role.TenantID, role.PropertyID)
		if err != nil && err != repositories.ErrNotFound {
			return fmt.Errorf("failed to check existing originating broker: %w", err)
		}
		if existing != nil {
			return fmt.Errorf("property already has an originating broker (ID: %s)", existing.BrokerID)
		}
	}

	// Check if this broker-property-role combination already exists
	existing, err := s.roleRepo.GetByPropertyAndBroker(ctx, role.TenantID, role.PropertyID, role.BrokerID, role.Role)
	if err != nil && err != repositories.ErrNotFound {
		return fmt.Errorf("failed to check existing role: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("broker already has this role for this property")
	}

	// Create role in repository
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return fmt.Errorf("failed to assign broker to property: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, role.TenantID, "co_broker_added", models.ActorTypeSystem, "", map[string]interface{}{
		"role_id":      role.ID,
		"property_id":  role.PropertyID,
		"broker_id":    role.BrokerID,
		"role":         role.Role,
		"is_primary":   role.IsPrimary,
	})

	return nil
}

// RemoveBrokerFromProperty removes a broker's role from a property
func (s *PropertyBrokerRoleService) RemoveBrokerFromProperty(ctx context.Context, tenantID, roleID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Validate role exists
	role, err := s.roleRepo.Get(ctx, tenantID, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Business rule: Cannot remove originating_broker
	if role.Role == models.BrokerPropertyRoleOriginating {
		return fmt.Errorf("cannot remove originating broker from property")
	}

	// Delete role from repository
	if err := s.roleRepo.Delete(ctx, tenantID, roleID); err != nil {
		return fmt.Errorf("failed to remove broker from property: %w", err)
	}

	// If this was the primary, assign another broker as primary
	if role.IsPrimary {
		// Find another role to set as primary
		otherRoles, err := s.roleRepo.ListByProperty(ctx, tenantID, role.PropertyID, repositories.PaginationOptions{Limit: 10})
		if err != nil {
			return fmt.Errorf("failed to find other roles: %w", err)
		}

		// Prefer originating broker, then listing broker
		var newPrimaryID string
		for _, otherRole := range otherRoles {
			if otherRole.Role == models.BrokerPropertyRoleOriginating {
				newPrimaryID = otherRole.ID
				break
			}
		}
		if newPrimaryID == "" {
			for _, otherRole := range otherRoles {
				if otherRole.Role == models.BrokerPropertyRoleListing {
					newPrimaryID = otherRole.ID
					break
				}
			}
		}

		if newPrimaryID != "" {
			if err := s.SetPrimaryBroker(ctx, tenantID, newPrimaryID); err != nil {
				return fmt.Errorf("failed to set new primary broker: %w", err)
			}
		}
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "co_broker_removed", models.ActorTypeSystem, "", map[string]interface{}{
		"role_id":     roleID,
		"property_id": role.PropertyID,
		"broker_id":   role.BrokerID,
		"role":        role.Role,
	})

	return nil
}

// UpdateRole updates a broker's role for a property
func (s *PropertyBrokerRoleService) UpdateRole(ctx context.Context, tenantID, roleID string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Validate role exists
	existing, err := s.roleRepo.Get(ctx, tenantID, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Validate new role type if being updated
	if newRole, ok := updates["role"].(models.BrokerPropertyRole); ok {
		if err := s.validateRole(newRole); err != nil {
			return err
		}

		// Business rule: Cannot change to/from originating_broker
		if existing.Role == models.BrokerPropertyRoleOriginating || newRole == models.BrokerPropertyRoleOriginating {
			return fmt.Errorf("cannot change originating_broker role")
		}
	}

	// Prevent updating tenant_id, property_id, broker_id
	delete(updates, "tenant_id")
	delete(updates, "property_id")
	delete(updates, "broker_id")

	// Update role in repository
	if err := s.roleRepo.Update(ctx, tenantID, roleID, updates); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_role_updated", models.ActorTypeSystem, "", map[string]interface{}{
		"role_id":     roleID,
		"property_id": existing.PropertyID,
		"broker_id":   existing.BrokerID,
		"updates":     updates,
	})

	return nil
}

// SetPrimaryBroker sets a broker as the primary for lead routing
func (s *PropertyBrokerRoleService) SetPrimaryBroker(ctx context.Context, tenantID, roleID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Validate role exists
	role, err := s.roleRepo.Get(ctx, tenantID, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if already primary
	if role.IsPrimary {
		return nil // Already primary, nothing to do
	}

	// Set as primary (repository will unset others)
	updates := map[string]interface{}{
		"is_primary": true,
	}

	if err := s.roleRepo.Update(ctx, tenantID, roleID, updates); err != nil {
		return fmt.Errorf("failed to set primary broker: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "primary_broker_changed", models.ActorTypeSystem, "", map[string]interface{}{
		"role_id":     roleID,
		"property_id": role.PropertyID,
		"broker_id":   role.BrokerID,
	})

	return nil
}

// GetPropertyBrokers retrieves all brokers for a property with their roles
func (s *PropertyBrokerRoleService) GetPropertyBrokers(ctx context.Context, tenantID, propertyID string, opts repositories.PaginationOptions) ([]*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	roles, err := s.roleRepo.ListByProperty(ctx, tenantID, propertyID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get property brokers: %w", err)
	}

	return roles, nil
}

// GetBrokerProperties retrieves all properties for a broker with their roles
func (s *PropertyBrokerRoleService) GetBrokerProperties(ctx context.Context, tenantID, brokerID string, opts repositories.PaginationOptions) ([]*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if brokerID == "" {
		return nil, fmt.Errorf("broker_id is required")
	}

	roles, err := s.roleRepo.ListByBroker(ctx, tenantID, brokerID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker properties: %w", err)
	}

	return roles, nil
}

// GetOriginatingBroker retrieves the originating broker for a property
func (s *PropertyBrokerRoleService) GetOriginatingBroker(ctx context.Context, tenantID, propertyID string) (*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	role, err := s.roleRepo.GetOriginatingBroker(ctx, tenantID, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get originating broker: %w", err)
	}

	return role, nil
}

// GetPrimaryBroker retrieves the primary broker for a property (for lead routing)
func (s *PropertyBrokerRoleService) GetPrimaryBroker(ctx context.Context, tenantID, propertyID string) (*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	role, err := s.roleRepo.GetPrimaryBroker(ctx, tenantID, propertyID)
	if err != nil {
		// If no primary broker, return the originating broker as default
		if err == repositories.ErrNotFound {
			return s.roleRepo.GetOriginatingBroker(ctx, tenantID, propertyID)
		}
		return nil, fmt.Errorf("failed to get primary broker: %w", err)
	}

	return role, nil
}

// CalculateCommissionSplit calculates commission split for co-brokerage (MVP+1 preparation)
// In MVP, this just returns the commission percentages without actual calculation
func (s *PropertyBrokerRoleService) CalculateCommissionSplit(ctx context.Context, tenantID, propertyID string) (map[string]float64, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	// Get all roles for the property
	roles, err := s.roleRepo.ListByProperty(ctx, tenantID, propertyID, repositories.PaginationOptions{Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to get property roles: %w", err)
	}

	// Build commission split map
	commissionSplit := make(map[string]float64)
	for _, role := range roles {
		commissionSplit[role.BrokerID] = role.CommissionPercentage
	}

	return commissionSplit, nil
}

// validateRole validates broker property role
func (s *PropertyBrokerRoleService) validateRole(role models.BrokerPropertyRole) error {
	validRoles := map[models.BrokerPropertyRole]bool{
		models.BrokerPropertyRoleOriginating: true,
		models.BrokerPropertyRoleListing:     true,
		models.BrokerPropertyRoleCoBroker:    true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid broker property role")
	}

	return nil
}

// logActivity logs an activity (helper method)
func (s *PropertyBrokerRoleService) logActivity(ctx context.Context, tenantID, eventType string, actorType models.ActorType, actorID string, metadata map[string]interface{}) error {
	log := &models.ActivityLog{
		TenantID:  tenantID,
		EventType: eventType,
		ActorType: actorType,
		ActorID:   actorID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	return s.activityLogRepo.Create(ctx, log)
}
