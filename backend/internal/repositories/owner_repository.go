package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

// OwnerRepository handles Firestore operations for property owners
type OwnerRepository struct {
	*BaseRepository
}

// NewOwnerRepository creates a new owner repository
func NewOwnerRepository(client *firestore.Client) *OwnerRepository {
	return &OwnerRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getOwnersCollection returns the collection path for owners within a tenant
func (r *OwnerRepository) getOwnersCollection(tenantID string) string {
	return fmt.Sprintf("tenants/%s/owners", tenantID)
}

// Create creates a new owner
func (r *OwnerRepository) Create(ctx context.Context, owner *models.Owner) error {
	if owner.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if owner.ID == "" {
		owner.ID = r.GenerateID(r.getOwnersCollection(owner.TenantID))
	}

	now := time.Now()
	owner.CreatedAt = now
	owner.UpdatedAt = now

	collectionPath := r.getOwnersCollection(owner.TenantID)
	if err := r.CreateDocument(ctx, collectionPath, owner.ID, owner); err != nil {
		return fmt.Errorf("failed to create owner: %w", err)
	}

	return nil
}

// Get retrieves an owner by ID
func (r *OwnerRepository) Get(ctx context.Context, tenantID, id string) (*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	var owner models.Owner
	collectionPath := r.getOwnersCollection(tenantID)
	if err := r.GetDocument(ctx, collectionPath, id, &owner); err != nil {
		return nil, err
	}

	owner.ID = id
	return &owner, nil
}

// GetByEmail retrieves an owner by email
func (r *OwnerRepository) GetByEmail(ctx context.Context, tenantID, email string) (*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	collectionPath := r.getOwnersCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("email", "==", email).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query owner by email: %w", err)
	}

	var owner models.Owner
	if err := doc.DataTo(&owner); err != nil {
		return nil, fmt.Errorf("failed to decode owner: %w", err)
	}

	owner.ID = doc.Ref.ID
	return &owner, nil
}

// GetByDocument retrieves an owner by document (CPF/CNPJ)
func (r *OwnerRepository) GetByDocument(ctx context.Context, tenantID, document string) (*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if document == "" {
		return nil, fmt.Errorf("%w: document is required", ErrInvalidInput)
	}

	collectionPath := r.getOwnersCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("document", "==", document).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query owner by document: %w", err)
	}

	var owner models.Owner
	if err := doc.DataTo(&owner); err != nil {
		return nil, fmt.Errorf("failed to decode owner: %w", err)
	}

	owner.ID = doc.Ref.ID
	return &owner, nil
}

// Update updates an owner
func (r *OwnerRepository) Update(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: owner ID is required", ErrInvalidInput)
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

	collectionPath := r.getOwnersCollection(tenantID)
	if err := r.UpdateDocument(ctx, collectionPath, id, firestoreUpdates); err != nil {
		return fmt.Errorf("failed to update owner: %w", err)
	}

	return nil
}

// Delete deletes an owner
func (r *OwnerRepository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	collectionPath := r.getOwnersCollection(tenantID)
	if err := r.DeleteDocument(ctx, collectionPath, id); err != nil {
		return fmt.Errorf("failed to delete owner: %w", err)
	}
	return nil
}

// List retrieves all owners for a tenant with pagination
func (r *OwnerRepository) List(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getOwnersCollection(tenantID)
	query := r.Client().Collection(collectionPath).Query
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	owners := make([]*models.Owner, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate owners: %w", err)
		}

		var owner models.Owner
		if err := doc.DataTo(&owner); err != nil {
			return nil, fmt.Errorf("failed to decode owner: %w", err)
		}

		owner.ID = doc.Ref.ID
		owners = append(owners, &owner)
	}

	return owners, nil
}

// ListByStatus retrieves owners by status
func (r *OwnerRepository) ListByStatus(ctx context.Context, tenantID string, status models.OwnerStatus, opts PaginationOptions) ([]*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getOwnersCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("owner_status", "==", string(status))
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	owners := make([]*models.Owner, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate owners by status: %w", err)
		}

		var owner models.Owner
		if err := doc.DataTo(&owner); err != nil {
			return nil, fmt.Errorf("failed to decode owner: %w", err)
		}

		owner.ID = doc.Ref.ID
		owners = append(owners, &owner)
	}

	return owners, nil
}

// ListWithoutConsent retrieves owners without consent
func (r *OwnerRepository) ListWithoutConsent(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getOwnersCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("consent_given", "==", false).
		Where("is_anonymized", "==", false)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	owners := make([]*models.Owner, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate owners without consent: %w", err)
		}

		var owner models.Owner
		if err := doc.DataTo(&owner); err != nil {
			return nil, fmt.Errorf("failed to decode owner: %w", err)
		}

		owner.ID = doc.Ref.ID
		owners = append(owners, &owner)
	}

	return owners, nil
}
