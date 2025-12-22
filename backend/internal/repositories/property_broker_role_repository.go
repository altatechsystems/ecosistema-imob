package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

// PropertyBrokerRoleRepository handles Firestore operations for property-broker role assignments
type PropertyBrokerRoleRepository struct {
	*BaseRepository
}

// NewPropertyBrokerRoleRepository creates a new property broker role repository
func NewPropertyBrokerRoleRepository(client *firestore.Client) *PropertyBrokerRoleRepository {
	return &PropertyBrokerRoleRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getRolesCollection returns the collection path for property broker roles within a tenant
func (r *PropertyBrokerRoleRepository) getRolesCollection(tenantID string) string {
	return fmt.Sprintf("tenants/%s/property_broker_roles", tenantID)
}

// Create creates a new property broker role
func (r *PropertyBrokerRoleRepository) Create(ctx context.Context, role *models.PropertyBrokerRole) error {
	if role.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if role.PropertyID == "" {
		return fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}
	if role.BrokerID == "" {
		return fmt.Errorf("%w: broker_id is required", ErrInvalidInput)
	}

	if role.ID == "" {
		role.ID = r.GenerateID(r.getRolesCollection(role.TenantID))
	}

	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	// If setting as primary, unset other primary roles for the property
	if role.IsPrimary {
		if err := r.UnsetPrimaryForProperty(ctx, role.TenantID, role.PropertyID); err != nil {
			return fmt.Errorf("failed to unset existing primary roles: %w", err)
		}
	}

	collectionPath := r.getRolesCollection(role.TenantID)
	if err := r.CreateDocument(ctx, collectionPath, role.ID, role); err != nil {
		return fmt.Errorf("failed to create property broker role: %w", err)
	}

	return nil
}

// Get retrieves a property broker role by ID
func (r *PropertyBrokerRoleRepository) Get(ctx context.Context, tenantID, id string) (*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	var role models.PropertyBrokerRole
	collectionPath := r.getRolesCollection(tenantID)
	if err := r.GetDocument(ctx, collectionPath, id, &role); err != nil {
		return nil, err
	}

	role.ID = id
	return &role, nil
}

// GetByPropertyAndBroker retrieves a specific role for a property-broker pair
func (r *PropertyBrokerRoleRepository) GetByPropertyAndBroker(ctx context.Context, tenantID, propertyID, brokerID string, roleType models.BrokerPropertyRole) (*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}
	if brokerID == "" {
		return nil, fmt.Errorf("%w: broker_id is required", ErrInvalidInput)
	}

	collectionPath := r.getRolesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID).
		Where("broker_id", "==", brokerID).
		Where("role", "==", string(roleType)).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query property broker role: %w", err)
	}

	var role models.PropertyBrokerRole
	if err := doc.DataTo(&role); err != nil {
		return nil, fmt.Errorf("failed to decode property broker role: %w", err)
	}

	role.ID = doc.Ref.ID
	return &role, nil
}

// Update updates a property broker role
func (r *PropertyBrokerRoleRepository) Update(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: role ID is required", ErrInvalidInput)
	}

	// If setting as primary, get the current role to know which property
	if isPrimary, ok := updates["is_primary"].(bool); ok && isPrimary {
		currentRole, err := r.Get(ctx, tenantID, id)
		if err != nil {
			return fmt.Errorf("failed to get current role: %w", err)
		}

		if err := r.UnsetPrimaryForProperty(ctx, tenantID, currentRole.PropertyID); err != nil {
			return fmt.Errorf("failed to unset existing primary roles: %w", err)
		}
	}

	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	// Convert map to firestore updates
	firestoreUpdates := make([]firestore.Update, 0, len(updates))
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	collectionPath := r.getRolesCollection(tenantID)
	if err := r.UpdateDocument(ctx, collectionPath, id, firestoreUpdates); err != nil {
		return fmt.Errorf("failed to update property broker role: %w", err)
	}

	return nil
}

// Delete deletes a property broker role
func (r *PropertyBrokerRoleRepository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	collectionPath := r.getRolesCollection(tenantID)
	if err := r.DeleteDocument(ctx, collectionPath, id); err != nil {
		return fmt.Errorf("failed to delete property broker role: %w", err)
	}
	return nil
}

// ListByProperty retrieves all broker roles for a property
func (r *PropertyBrokerRoleRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, opts PaginationOptions) ([]*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getRolesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	roles := make([]*models.PropertyBrokerRole, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate property broker roles: %w", err)
		}

		var role models.PropertyBrokerRole
		if err := doc.DataTo(&role); err != nil {
			return nil, fmt.Errorf("failed to decode property broker role: %w", err)
		}

		role.ID = doc.Ref.ID
		roles = append(roles, &role)
	}

	return roles, nil
}

// ListByBroker retrieves all property roles for a broker
func (r *PropertyBrokerRoleRepository) ListByBroker(ctx context.Context, tenantID, brokerID string, opts PaginationOptions) ([]*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if brokerID == "" {
		return nil, fmt.Errorf("%w: broker_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getRolesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("broker_id", "==", brokerID)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	roles := make([]*models.PropertyBrokerRole, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate property broker roles: %w", err)
		}

		var role models.PropertyBrokerRole
		if err := doc.DataTo(&role); err != nil {
			return nil, fmt.Errorf("failed to decode property broker role: %w", err)
		}

		role.ID = doc.Ref.ID
		roles = append(roles, &role)
	}

	return roles, nil
}

// ListByRole retrieves roles by type
func (r *PropertyBrokerRoleRepository) ListByRole(ctx context.Context, tenantID string, roleType models.BrokerPropertyRole, opts PaginationOptions) ([]*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getRolesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("role", "==", string(roleType))
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	roles := make([]*models.PropertyBrokerRole, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate property broker roles by type: %w", err)
		}

		var role models.PropertyBrokerRole
		if err := doc.DataTo(&role); err != nil {
			return nil, fmt.Errorf("failed to decode property broker role: %w", err)
		}

		role.ID = doc.Ref.ID
		roles = append(roles, &role)
	}

	return roles, nil
}

// GetOriginatingBroker retrieves the originating broker for a property
func (r *PropertyBrokerRoleRepository) GetOriginatingBroker(ctx context.Context, tenantID, propertyID string) (*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	collectionPath := r.getRolesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID).
		Where("role", "==", string(models.BrokerPropertyRoleOriginating)).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query originating broker: %w", err)
	}

	var role models.PropertyBrokerRole
	if err := doc.DataTo(&role); err != nil {
		return nil, fmt.Errorf("failed to decode property broker role: %w", err)
	}

	role.ID = doc.Ref.ID
	return &role, nil
}

// GetPrimaryBroker retrieves the primary broker for a property (for lead routing)
func (r *PropertyBrokerRoleRepository) GetPrimaryBroker(ctx context.Context, tenantID, propertyID string) (*models.PropertyBrokerRole, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	collectionPath := r.getRolesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID).
		Where("is_primary", "==", true).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query primary broker: %w", err)
	}

	var role models.PropertyBrokerRole
	if err := doc.DataTo(&role); err != nil {
		return nil, fmt.Errorf("failed to decode property broker role: %w", err)
	}

	role.ID = doc.Ref.ID
	return &role, nil
}

// UnsetPrimaryForProperty unsets the primary flag for all roles of a property
func (r *PropertyBrokerRoleRepository) UnsetPrimaryForProperty(ctx context.Context, tenantID, propertyID string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	// Get all roles for the property
	roles, err := r.ListByProperty(ctx, tenantID, propertyID, PaginationOptions{Limit: 100})
	if err != nil {
		return fmt.Errorf("failed to list roles for property: %w", err)
	}

	// Update each role to unset primary flag
	batch := r.Client().Batch()
	collectionPath := r.getRolesCollection(tenantID)

	for _, role := range roles {
		if role.IsPrimary {
			docRef := r.Client().Collection(collectionPath).Doc(role.ID)
			batch.Update(docRef, []firestore.Update{
				{Path: "is_primary", Value: false},
				{Path: "updated_at", Value: time.Now()},
			})
		}
	}

	_, err = batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit batch update: %w", err)
	}

	return nil
}
