package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

// ListingRepository handles Firestore operations for listings
type ListingRepository struct {
	*BaseRepository
}

// NewListingRepository creates a new listing repository
func NewListingRepository(client *firestore.Client) *ListingRepository {
	return &ListingRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getListingsCollection returns the collection path for listings within a tenant
func (r *ListingRepository) getListingsCollection(tenantID string) string {
	return fmt.Sprintf("tenants/%s/listings", tenantID)
}

// Create creates a new listing
func (r *ListingRepository) Create(ctx context.Context, listing *models.Listing) error {
	if listing.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if listing.PropertyID == "" {
		return fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}
	if listing.BrokerID == "" {
		return fmt.Errorf("%w: broker_id is required", ErrInvalidInput)
	}

	if listing.ID == "" {
		listing.ID = r.GenerateID(r.getListingsCollection(listing.TenantID))
	}

	now := time.Now()
	listing.CreatedAt = now
	listing.UpdatedAt = now

	collectionPath := r.getListingsCollection(listing.TenantID)
	if err := r.CreateDocument(ctx, collectionPath, listing.ID, listing); err != nil {
		return fmt.Errorf("failed to create listing: %w", err)
	}

	return nil
}

// Get retrieves a listing by ID
func (r *ListingRepository) Get(ctx context.Context, tenantID, id string) (*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	var listing models.Listing
	collectionPath := r.getListingsCollection(tenantID)
	if err := r.GetDocument(ctx, collectionPath, id, &listing); err != nil {
		return nil, err
	}

	listing.ID = id
	return &listing, nil
}

// Update updates a listing
func (r *ListingRepository) Update(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: listing ID is required", ErrInvalidInput)
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

	collectionPath := r.getListingsCollection(tenantID)
	if err := r.UpdateDocument(ctx, collectionPath, id, firestoreUpdates); err != nil {
		return fmt.Errorf("failed to update listing: %w", err)
	}

	return nil
}

// Delete deletes a listing
func (r *ListingRepository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	collectionPath := r.getListingsCollection(tenantID)
	if err := r.DeleteDocument(ctx, collectionPath, id); err != nil {
		return fmt.Errorf("failed to delete listing: %w", err)
	}
	return nil
}

// List retrieves all listings for a tenant with pagination
func (r *ListingRepository) List(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getListingsCollection(tenantID)
	query := r.Client().Collection(collectionPath).Query
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	listings := make([]*models.Listing, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate listings: %w", err)
		}

		var listing models.Listing
		if err := doc.DataTo(&listing); err != nil {
			return nil, fmt.Errorf("failed to decode listing: %w", err)
		}

		listing.ID = doc.Ref.ID
		listings = append(listings, &listing)
	}

	return listings, nil
}

// ListByProperty retrieves all listings for a property
func (r *ListingRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, opts PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getListingsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	listings := make([]*models.Listing, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate listings by property: %w", err)
		}

		var listing models.Listing
		if err := doc.DataTo(&listing); err != nil {
			return nil, fmt.Errorf("failed to decode listing: %w", err)
		}

		listing.ID = doc.Ref.ID
		listings = append(listings, &listing)
	}

	return listings, nil
}

// ListByBroker retrieves all listings for a broker
func (r *ListingRepository) ListByBroker(ctx context.Context, tenantID, brokerID string, opts PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if brokerID == "" {
		return nil, fmt.Errorf("%w: broker_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getListingsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("broker_id", "==", brokerID)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	listings := make([]*models.Listing, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate listings by broker: %w", err)
		}

		var listing models.Listing
		if err := doc.DataTo(&listing); err != nil {
			return nil, fmt.Errorf("failed to decode listing: %w", err)
		}

		listing.ID = doc.Ref.ID
		listings = append(listings, &listing)
	}

	return listings, nil
}

// ListActive retrieves all active listings
func (r *ListingRepository) ListActive(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getListingsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("is_active", "==", true)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	listings := make([]*models.Listing, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate active listings: %w", err)
		}

		var listing models.Listing
		if err := doc.DataTo(&listing); err != nil {
			return nil, fmt.Errorf("failed to decode listing: %w", err)
		}

		listing.ID = doc.Ref.ID
		listings = append(listings, &listing)
	}

	return listings, nil
}

// GetCanonicalForProperty retrieves the canonical listing for a property
func (r *ListingRepository) GetCanonicalForProperty(ctx context.Context, tenantID, propertyID string) (*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	collectionPath := r.getListingsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID).
		Where("is_canonical", "==", true).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query canonical listing: %w", err)
	}

	var listing models.Listing
	if err := doc.DataTo(&listing); err != nil {
		return nil, fmt.Errorf("failed to decode listing: %w", err)
	}

	listing.ID = doc.Ref.ID
	return &listing, nil
}

// UnsetCanonicalForProperty unsets the canonical flag for all listings of a property
func (r *ListingRepository) UnsetCanonicalForProperty(ctx context.Context, tenantID, propertyID string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	// Get all listings for the property
	listings, err := r.ListByProperty(ctx, tenantID, propertyID, PaginationOptions{Limit: 100})
	if err != nil {
		return fmt.Errorf("failed to list listings for property: %w", err)
	}

	// Update each listing to unset canonical flag
	batch := r.Client().Batch()
	collectionPath := r.getListingsCollection(tenantID)

	for _, listing := range listings {
		if listing.IsCanonical {
			docRef := r.Client().Collection(collectionPath).Doc(listing.ID)
			batch.Update(docRef, []firestore.Update{
				{Path: "is_canonical", Value: false},
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
