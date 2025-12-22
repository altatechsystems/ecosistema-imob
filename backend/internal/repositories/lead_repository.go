package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

// LeadRepository handles Firestore operations for leads
type LeadRepository struct {
	*BaseRepository
}

// NewLeadRepository creates a new lead repository
func NewLeadRepository(client *firestore.Client) *LeadRepository {
	return &LeadRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getLeadsCollection returns the collection path for leads within a tenant
func (r *LeadRepository) getLeadsCollection(tenantID string) string {
	return fmt.Sprintf("tenants/%s/leads", tenantID)
}

// LeadFilters contains optional filters for lead queries
type LeadFilters struct {
	PropertyID string
	Status     *models.LeadStatus
	Channel    *models.LeadChannel
}

// Create creates a new lead
func (r *LeadRepository) Create(ctx context.Context, lead *models.Lead) error {
	if lead.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if lead.PropertyID == "" {
		return fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}
	if !lead.ConsentGiven {
		return fmt.Errorf("%w: consent_given must be true to create a lead", ErrInvalidInput)
	}

	if lead.ID == "" {
		lead.ID = r.GenerateID(r.getLeadsCollection(lead.TenantID))
	}

	now := time.Now()
	lead.CreatedAt = now
	lead.UpdatedAt = now

	// Set consent date if not provided
	if lead.ConsentDate.IsZero() {
		lead.ConsentDate = now
	}

	collectionPath := r.getLeadsCollection(lead.TenantID)
	if err := r.CreateDocument(ctx, collectionPath, lead.ID, lead); err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}

	return nil
}

// Get retrieves a lead by ID
func (r *LeadRepository) Get(ctx context.Context, tenantID, id string) (*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	var lead models.Lead
	collectionPath := r.getLeadsCollection(tenantID)
	if err := r.GetDocument(ctx, collectionPath, id, &lead); err != nil {
		return nil, err
	}

	lead.ID = id
	return &lead, nil
}

// Update updates a lead
func (r *LeadRepository) Update(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: lead ID is required", ErrInvalidInput)
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

	collectionPath := r.getLeadsCollection(tenantID)
	if err := r.UpdateDocument(ctx, collectionPath, id, firestoreUpdates); err != nil {
		return fmt.Errorf("failed to update lead: %w", err)
	}

	return nil
}

// Delete deletes a lead (should be rare - prefer anonymization)
func (r *LeadRepository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	collectionPath := r.getLeadsCollection(tenantID)
	if err := r.DeleteDocument(ctx, collectionPath, id); err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}
	return nil
}

// List retrieves leads for a tenant with optional filters and pagination
func (r *LeadRepository) List(ctx context.Context, tenantID string, filters *LeadFilters, opts PaginationOptions) ([]*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getLeadsCollection(tenantID)
	query := r.Client().Collection(collectionPath).Query

	// Apply filters if provided
	if filters != nil {
		if filters.PropertyID != "" {
			query = query.Where("property_id", "==", filters.PropertyID)
		}
		if filters.Status != nil {
			query = query.Where("status", "==", string(*filters.Status))
		}
		if filters.Channel != nil {
			query = query.Where("channel", "==", string(*filters.Channel))
		}
	}

	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	leads := make([]*models.Lead, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate leads: %w", err)
		}

		var lead models.Lead
		if err := doc.DataTo(&lead); err != nil {
			return nil, fmt.Errorf("failed to decode lead: %w", err)
		}

		lead.ID = doc.Ref.ID
		leads = append(leads, &lead)
	}

	return leads, nil
}

// ListByProperty retrieves all leads for a property
func (r *LeadRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, opts PaginationOptions) ([]*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	filters := &LeadFilters{PropertyID: propertyID}
	return r.List(ctx, tenantID, filters, opts)
}

// ListByStatus retrieves leads by status
func (r *LeadRepository) ListByStatus(ctx context.Context, tenantID string, status models.LeadStatus, opts PaginationOptions) ([]*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	filters := &LeadFilters{Status: &status}
	return r.List(ctx, tenantID, filters, opts)
}

// ListByChannel retrieves leads by channel
func (r *LeadRepository) ListByChannel(ctx context.Context, tenantID string, channel models.LeadChannel, opts PaginationOptions) ([]*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	filters := &LeadFilters{Channel: &channel}
	return r.List(ctx, tenantID, filters, opts)
}

// GetByEmail retrieves a lead by email within a property context
func (r *LeadRepository) GetByEmail(ctx context.Context, tenantID, propertyID, email string) (*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	collectionPath := r.getLeadsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID).
		Where("email", "==", email).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query lead by email: %w", err)
	}

	var lead models.Lead
	if err := doc.DataTo(&lead); err != nil {
		return nil, fmt.Errorf("failed to decode lead: %w", err)
	}

	lead.ID = doc.Ref.ID
	return &lead, nil
}

// GetByPhone retrieves a lead by phone within a property context
func (r *LeadRepository) GetByPhone(ctx context.Context, tenantID, propertyID, phone string) (*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}
	if phone == "" {
		return nil, fmt.Errorf("%w: phone is required", ErrInvalidInput)
	}

	collectionPath := r.getLeadsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("property_id", "==", propertyID).
		Where("phone", "==", phone).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query lead by phone: %w", err)
	}

	var lead models.Lead
	if err := doc.DataTo(&lead); err != nil {
		return nil, fmt.Errorf("failed to decode lead: %w", err)
	}

	lead.ID = doc.Ref.ID
	return &lead, nil
}

// ListWithRevokedConsent retrieves leads with revoked consent
func (r *LeadRepository) ListWithRevokedConsent(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getLeadsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("consent_revoked", "==", true).
		Where("is_anonymized", "==", false)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	leads := make([]*models.Lead, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate leads with revoked consent: %w", err)
		}

		var lead models.Lead
		if err := doc.DataTo(&lead); err != nil {
			return nil, fmt.Errorf("failed to decode lead: %w", err)
		}

		lead.ID = doc.Ref.ID
		leads = append(leads, &lead)
	}

	return leads, nil
}

// RevokeConsent marks a lead's consent as revoked
func (r *LeadRepository) RevokeConsent(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: lead ID is required", ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"consent_revoked": true,
		"revoked_at":      now,
		"updated_at":      now,
	}

	return r.Update(ctx, tenantID, id, updates)
}

// Anonymize anonymizes a lead's personal data (LGPD compliance)
func (r *LeadRepository) Anonymize(ctx context.Context, tenantID, id string, reason string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: lead ID is required", ErrInvalidInput)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"name":                  "ANONYMIZED",
		"email":                 "",
		"phone":                 "",
		"message":               "",
		"consent_ip":            "",
		"is_anonymized":         true,
		"anonymized_at":         now,
		"anonymization_reason":  reason,
		"updated_at":            now,
	}

	return r.Update(ctx, tenantID, id, updates)
}
