package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

const (
	tenantsCollection = "tenants"
)

// TenantRepository handles Firestore operations for tenants
type TenantRepository struct {
	*BaseRepository
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(client *firestore.Client) *TenantRepository {
	return &TenantRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	if tenant.ID == "" {
		tenant.ID = r.GenerateID(tenantsCollection)
	}

	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	if err := r.CreateDocument(ctx, tenantsCollection, tenant.ID, tenant); err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// Get retrieves a tenant by ID
func (r *TenantRepository) Get(ctx context.Context, id string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.GetDocument(ctx, tenantsCollection, id, &tenant); err != nil {
		return nil, err
	}

	tenant.ID = id
	return &tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", ErrInvalidInput)
	}

	query := r.Client().Collection(tenantsCollection).
		Where("slug", "==", slug).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query tenant by slug: %w", err)
	}

	var tenant models.Tenant
	if err := doc.DataTo(&tenant); err != nil {
		return nil, fmt.Errorf("failed to decode tenant: %w", err)
	}

	tenant.ID = doc.Ref.ID
	return &tenant, nil
}

// Update updates a tenant
func (r *TenantRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return fmt.Errorf("%w: tenant ID is required", ErrInvalidInput)
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

	if err := r.UpdateDocument(ctx, tenantsCollection, id, firestoreUpdates); err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

// Delete deletes a tenant
func (r *TenantRepository) Delete(ctx context.Context, id string) error {
	if err := r.DeleteDocument(ctx, tenantsCollection, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

// List retrieves all tenants with pagination
func (r *TenantRepository) List(ctx context.Context, opts PaginationOptions) ([]*models.Tenant, error) {
	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	query := r.Client().Collection(tenantsCollection).Query
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	tenants := make([]*models.Tenant, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate tenants: %w", err)
		}

		var tenant models.Tenant
		if err := doc.DataTo(&tenant); err != nil {
			return nil, fmt.Errorf("failed to decode tenant: %w", err)
		}

		tenant.ID = doc.Ref.ID
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// ListActive retrieves all active tenants
func (r *TenantRepository) ListActive(ctx context.Context, opts PaginationOptions) ([]*models.Tenant, error) {
	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	query := r.Client().Collection(tenantsCollection).
		Where("is_active", "==", true)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	tenants := make([]*models.Tenant, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate active tenants: %w", err)
		}

		var tenant models.Tenant
		if err := doc.DataTo(&tenant); err != nil {
			return nil, fmt.Errorf("failed to decode tenant: %w", err)
		}

		tenant.ID = doc.Ref.ID
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}
