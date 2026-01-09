package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

// PropertyRepository handles Firestore operations for properties
type PropertyRepository struct {
	*BaseRepository
}

// NewPropertyRepository creates a new property repository
func NewPropertyRepository(client *firestore.Client) *PropertyRepository {
	return &PropertyRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getPropertiesCollection returns the collection path for properties
// Properties are stored in root collection with tenant_id field, not subcollection
func (r *PropertyRepository) getPropertiesCollection(tenantID string) string {
	return "properties"
}

// PropertyFilters contains optional filters for property queries
type PropertyFilters struct {
	Status          *models.PropertyStatus
	PropertyType    *models.PropertyType
	TransactionType *models.TransactionType
	Visibility      *models.PropertyVisibility
	OwnerID         string
	City            string
	Neighborhood    string
	MinPrice        *float64
	MaxPrice        *float64
	MinBedrooms     *int
	MinBathrooms    *int
}

// Create creates a new property
func (r *PropertyRepository) Create(ctx context.Context, property *models.Property) error {
	if property.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if property.ID == "" {
		property.ID = r.GenerateID("properties")
	}

	now := time.Now()
	property.CreatedAt = now
	property.UpdatedAt = now

	// Store in root collection with tenant_id field
	if err := r.CreateDocument(ctx, "properties", property.ID, property); err != nil {
		return fmt.Errorf("failed to create property: %w", err)
	}

	return nil
}

// Get retrieves a property by ID
// If tenantID is empty, skips tenant verification (used for public endpoints)
func (r *PropertyRepository) Get(ctx context.Context, tenantID, id string) (*models.Property, error) {
	var property models.Property
	if err := r.GetDocument(ctx, "properties", id, &property); err != nil {
		return nil, err
	}

	// Verify tenant ownership only if tenantID is provided
	if tenantID != "" && property.TenantID != tenantID {
		return nil, ErrNotFound
	}

	property.ID = id
	return &property, nil
}

// GetBySlug retrieves a property by slug
func (r *PropertyRepository) GetBySlug(ctx context.Context, tenantID, slug string) (*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", ErrInvalidInput)
	}

	query := r.Client().Collection("properties").
		Where("tenant_id", "==", tenantID).
		Where("slug", "==", slug).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query property by slug: %w", err)
	}

	var property models.Property
	if err := doc.DataTo(&property); err != nil {
		return nil, fmt.Errorf("failed to decode property: %w", err)
	}

	property.ID = doc.Ref.ID
	return &property, nil
}

// GetBySlugPublic retrieves a PUBLIC property by slug (across all tenants)
// This is used by the public portal agregador
func (r *PropertyRepository) GetBySlugPublic(ctx context.Context, slug string) (*models.Property, error) {
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", ErrInvalidInput)
	}

	query := r.Client().Collection("properties").
		Where("slug", "==", slug).
		Where("visibility", "==", string(models.PropertyVisibilityPublic)).
		Where("status", "==", string(models.PropertyStatusAvailable)).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query public property by slug: %w", err)
	}

	var property models.Property
	if err := doc.DataTo(&property); err != nil {
		return nil, fmt.Errorf("failed to decode public property: %w", err)
	}

	property.ID = doc.Ref.ID
	return &property, nil
}

// GetByExternalID retrieves a property by external source and ID
func (r *PropertyRepository) GetByExternalID(ctx context.Context, tenantID, externalSource, externalID string) (*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if externalSource == "" || externalID == "" {
		return nil, fmt.Errorf("%w: external_source and external_id are required", ErrInvalidInput)
	}

	query := r.Client().Collection("properties").
		Where("tenant_id", "==", tenantID).
		Where("external_source", "==", externalSource).
		Where("external_id", "==", externalID).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query property by external ID: %w", err)
	}

	var property models.Property
	if err := doc.DataTo(&property); err != nil {
		return nil, fmt.Errorf("failed to decode property: %w", err)
	}

	property.ID = doc.Ref.ID
	return &property, nil
}

// Update updates a property
func (r *PropertyRepository) Update(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: property ID is required", ErrInvalidInput)
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

	if err := r.UpdateDocument(ctx, "properties", id, firestoreUpdates); err != nil {
		return fmt.Errorf("failed to update property: %w", err)
	}

	return nil
}

// Delete deletes a property
func (r *PropertyRepository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if err := r.DeleteDocument(ctx, "properties", id); err != nil {
		return fmt.Errorf("failed to delete property: %w", err)
	}
	return nil
}

// List retrieves properties for a tenant with optional filters and pagination
func (r *PropertyRepository) List(ctx context.Context, tenantID string, filters *PropertyFilters, opts PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getPropertiesCollection(tenantID)

	// WORKAROUND: If filtering by owner_id, use single Where to avoid composite index requirement
	// Then filter other fields in memory
	var query firestore.Query
	var needsMemoryFilter bool

	if filters != nil && filters.OwnerID != "" {
		// Query by owner_id only to avoid composite index requirement
		query = r.Client().Collection(collectionPath).Where("owner_id", "==", filters.OwnerID)
		needsMemoryFilter = true
	} else {
		// Normal query with tenant_id
		query = r.Client().Collection(collectionPath).Where("tenant_id", "==", tenantID)

		// Apply other filters if provided
		if filters != nil {
			if filters.Status != nil {
				query = query.Where("status", "==", string(*filters.Status))
			}
			if filters.PropertyType != nil {
				query = query.Where("property_type", "==", string(*filters.PropertyType))
			}
			if filters.TransactionType != nil {
				query = query.Where("transaction_type", "==", string(*filters.TransactionType))
			}
			if filters.Visibility != nil {
				query = query.Where("visibility", "==", string(*filters.Visibility))
			}
			if filters.City != "" {
				query = query.Where("city", "==", filters.City)
			}
			if filters.Neighborhood != "" {
				query = query.Where("neighborhood", "==", filters.Neighborhood)
			}
			if filters.MinPrice != nil {
				query = query.Where("price_amount", ">=", *filters.MinPrice)
			}
			if filters.MaxPrice != nil {
				query = query.Where("price_amount", "<=", *filters.MaxPrice)
			}
			if filters.MinBedrooms != nil {
				query = query.Where("bedrooms", ">=", *filters.MinBedrooms)
			}
			if filters.MinBathrooms != nil {
				query = query.Where("bathrooms", ">=", *filters.MinBathrooms)
			}
		}
	}

	// Apply pagination limit only (skip ordering to avoid index requirement)
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	// TODO: Re-enable ordering after creating composite indexes
	// query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	properties := make([]*models.Property, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate properties: %w", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			return nil, fmt.Errorf("failed to decode property: %w", err)
		}

		property.ID = doc.Ref.ID

		// If we're filtering by owner_id, also filter by tenant_id in memory
		if needsMemoryFilter {
			if property.TenantID != tenantID {
				continue // Skip properties from other tenants
			}
		}

		properties = append(properties, &property)
	}

	return properties, nil
}

// ListAllPublic retrieves PUBLIC properties across ALL tenants with optional filters and pagination
// This is used by the public portal agregador to list properties from all tenants
func (r *PropertyRepository) ListAllPublic(ctx context.Context, filters *PropertyFilters, opts PaginationOptions) ([]*models.Property, error) {
	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	// CRITICAL: Always filter by visibility=public and status=available for security
	query := r.Client().Collection("properties").
		Where("visibility", "==", string(models.PropertyVisibilityPublic)).
		Where("status", "==", string(models.PropertyStatusAvailable))

	// Apply optional filters if provided
	if filters != nil {
		if filters.PropertyType != nil {
			query = query.Where("property_type", "==", string(*filters.PropertyType))
		}
		if filters.TransactionType != nil {
			query = query.Where("transaction_type", "==", string(*filters.TransactionType))
		}
		if filters.City != "" {
			query = query.Where("city", "==", filters.City)
		}
		if filters.Neighborhood != "" {
			query = query.Where("neighborhood", "==", filters.Neighborhood)
		}
		if filters.MinPrice != nil {
			query = query.Where("price_amount", ">=", *filters.MinPrice)
		}
		if filters.MaxPrice != nil {
			query = query.Where("price_amount", "<=", *filters.MaxPrice)
		}
		if filters.MinBedrooms != nil {
			query = query.Where("bedrooms", ">=", *filters.MinBedrooms)
		}
		if filters.MinBathrooms != nil {
			query = query.Where("bathrooms", ">=", *filters.MinBathrooms)
		}
	}

	// Apply pagination limit
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	properties := make([]*models.Property, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate public properties: %w", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			return nil, fmt.Errorf("failed to decode public property: %w", err)
		}

		property.ID = doc.Ref.ID
		properties = append(properties, &property)
	}

	return properties, nil
}

// Count returns the total number of properties for a tenant with optional filters
func (r *PropertyRepository) Count(ctx context.Context, tenantID string, filters *PropertyFilters) (int, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	collectionPath := r.getPropertiesCollection(tenantID)
	query := r.Client().Collection(collectionPath).Where("tenant_id", "==", tenantID)

	// Apply filters if provided
	if filters != nil {
		if filters.Status != nil {
			query = query.Where("status", "==", string(*filters.Status))
		}
		if filters.PropertyType != nil {
			query = query.Where("property_type", "==", string(*filters.PropertyType))
		}
		if filters.TransactionType != nil {
			query = query.Where("transaction_type", "==", string(*filters.TransactionType))
		}
		if filters.Visibility != nil {
			query = query.Where("visibility", "==", string(*filters.Visibility))
		}
		if filters.City != "" {
			query = query.Where("city", "==", filters.City)
		}
		if filters.Neighborhood != "" {
			query = query.Where("neighborhood", "==", filters.Neighborhood)
		}
		if filters.OwnerID != "" {
			query = query.Where("owner_id", "==", filters.OwnerID)
		}
	}

	// Use Select() to only fetch document IDs for counting (more efficient)
	docs, err := query.Select().Documents(ctx).GetAll()
	if err != nil {
		return 0, fmt.Errorf("failed to count properties: %w", err)
	}

	return len(docs), nil
}

// ListByOwner retrieves all properties for an owner
func (r *PropertyRepository) ListByOwner(ctx context.Context, tenantID, ownerID string, opts PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if ownerID == "" {
		return nil, fmt.Errorf("%w: owner_id is required", ErrInvalidInput)
	}

	filters := &PropertyFilters{OwnerID: ownerID}
	return r.List(ctx, tenantID, filters, opts)
}

// ListByCaptador retrieves all properties for a captador (broker)
func (r *PropertyRepository) ListByCaptador(ctx context.Context, tenantID, captadorID string, opts PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if captadorID == "" {
		return nil, fmt.Errorf("%w: captador_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	// Query properties where captador_id matches
	query := r.Client().Collection("properties").
		Where("tenant_id", "==", tenantID).
		Where("captador_id", "==", captadorID)

	// Apply pagination limit
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	properties := make([]*models.Property, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate properties: %w", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			return nil, fmt.Errorf("failed to decode property: %w", err)
		}

		property.ID = doc.Ref.ID
		properties = append(properties, &property)
	}

	return properties, nil
}

// ListByStatus retrieves properties by status
func (r *PropertyRepository) ListByStatus(ctx context.Context, tenantID string, status models.PropertyStatus, opts PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	filters := &PropertyFilters{Status: &status}
	return r.List(ctx, tenantID, filters, opts)
}

// ListByVisibility retrieves properties by visibility level
func (r *PropertyRepository) ListByVisibility(ctx context.Context, tenantID string, visibility models.PropertyVisibility, opts PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	filters := &PropertyFilters{Visibility: &visibility}
	return r.List(ctx, tenantID, filters, opts)
}

// ListPossibleDuplicates retrieves properties marked as possible duplicates
func (r *PropertyRepository) ListPossibleDuplicates(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getPropertiesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("possible_duplicate", "==", true)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	properties := make([]*models.Property, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate possible duplicates: %w", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			return nil, fmt.Errorf("failed to decode property: %w", err)
		}

		property.ID = doc.Ref.ID
		properties = append(properties, &property)
	}

	return properties, nil
}

// ListByFingerprint retrieves properties by fingerprint (for deduplication)
func (r *PropertyRepository) ListByFingerprint(ctx context.Context, tenantID, fingerprint string) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if fingerprint == "" {
		return nil, fmt.Errorf("%w: fingerprint is required", ErrInvalidInput)
	}

	collectionPath := r.getPropertiesCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("fingerprint", "==", fingerprint)

	iter := query.Documents(ctx)
	defer iter.Stop()

	properties := make([]*models.Property, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate properties by fingerprint: %w", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			return nil, fmt.Errorf("failed to decode property: %w", err)
		}

		property.ID = doc.Ref.ID
		properties = append(properties, &property)
	}

	return properties, nil
}

// SearchByLocation searches properties by city, neighborhood, or both
func (r *PropertyRepository) SearchByLocation(ctx context.Context, tenantID, city, neighborhood string, opts PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if city == "" && neighborhood == "" {
		return nil, fmt.Errorf("%w: at least one of city or neighborhood is required", ErrInvalidInput)
	}

	filters := &PropertyFilters{
		City:         city,
		Neighborhood: neighborhood,
	}

	return r.List(ctx, tenantID, filters, opts)
}
