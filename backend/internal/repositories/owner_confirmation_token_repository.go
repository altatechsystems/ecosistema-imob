package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OwnerConfirmationTokenRepository handles Firestore operations for owner confirmation tokens
type OwnerConfirmationTokenRepository struct {
	*BaseRepository
}

// NewOwnerConfirmationTokenRepository creates a new owner confirmation token repository
func NewOwnerConfirmationTokenRepository(client *firestore.Client) *OwnerConfirmationTokenRepository {
	return &OwnerConfirmationTokenRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// Create creates a new owner confirmation token
func (r *OwnerConfirmationTokenRepository) Create(ctx context.Context, token *models.OwnerConfirmationToken) error {
	if token.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if token.PropertyID == "" {
		return fmt.Errorf("property_id is required")
	}
	if token.TokenHash == "" {
		return fmt.Errorf("token_hash is required")
	}

	// Set created_at
	token.CreatedAt = time.Now()

	// Create document
	ref := r.client.Collection(fmt.Sprintf("tenants/%s/owner_confirmation_tokens", token.TenantID)).NewDoc()
	token.ID = ref.ID

	if _, err := ref.Create(ctx, token); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to create owner confirmation token: %w", err)
	}

	return nil
}

// Get retrieves an owner confirmation token by ID
func (r *OwnerConfirmationTokenRepository) Get(ctx context.Context, tenantID, tokenID string) (*models.OwnerConfirmationToken, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if tokenID == "" {
		return nil, fmt.Errorf("token_id is required")
	}

	doc, err := r.client.Collection(fmt.Sprintf("tenants/%s/owner_confirmation_tokens", tenantID)).Doc(tokenID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get owner confirmation token: %w", err)
	}

	var token models.OwnerConfirmationToken
	if err := doc.DataTo(&token); err != nil {
		return nil, fmt.Errorf("failed to decode owner confirmation token: %w", err)
	}

	token.ID = doc.Ref.ID
	return &token, nil
}

// GetByTokenHash retrieves an owner confirmation token by its hash
// This is the primary lookup method for validating tokens
func (r *OwnerConfirmationTokenRepository) GetByTokenHash(ctx context.Context, tenantID, tokenHash string) (*models.OwnerConfirmationToken, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if tokenHash == "" {
		return nil, fmt.Errorf("token_hash is required")
	}

	iter := r.client.Collection(fmt.Sprintf("tenants/%s/owner_confirmation_tokens", tenantID)).
		Where("token_hash", "==", tokenHash).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query owner confirmation token: %w", err)
	}

	var token models.OwnerConfirmationToken
	if err := doc.DataTo(&token); err != nil {
		return nil, fmt.Errorf("failed to decode owner confirmation token: %w", err)
	}

	token.ID = doc.Ref.ID
	return &token, nil
}

// Update updates an owner confirmation token
func (r *OwnerConfirmationTokenRepository) Update(ctx context.Context, tenantID, tokenID string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if tokenID == "" {
		return fmt.Errorf("token_id is required")
	}

	ref := r.client.Collection(fmt.Sprintf("tenants/%s/owner_confirmation_tokens", tenantID)).Doc(tokenID)

	// Check if document exists
	_, err := ref.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to check owner confirmation token: %w", err)
	}

	// Update document
	if _, err := ref.Update(ctx, mapToFirestoreUpdates(updates)); err != nil {
		return fmt.Errorf("failed to update owner confirmation token: %w", err)
	}

	return nil
}

// ListByProperty lists all confirmation tokens for a specific property
func (r *OwnerConfirmationTokenRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, opts *PaginationOptions) ([]*models.OwnerConfirmationToken, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	if opts == nil {
		defaultOpts := DefaultPaginationOptions()
		opts = &defaultOpts
	}

	query := r.client.Collection(fmt.Sprintf("tenants/%s/owner_confirmation_tokens", tenantID)).
		Where("property_id", "==", propertyID).
		OrderBy("created_at", firestore.Desc).
		Limit(opts.Limit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var tokens []*models.OwnerConfirmationToken
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate owner confirmation tokens: %w", err)
		}

		var token models.OwnerConfirmationToken
		if err := doc.DataTo(&token); err != nil {
			return nil, fmt.Errorf("failed to decode owner confirmation token: %w", err)
		}

		token.ID = doc.Ref.ID
		tokens = append(tokens, &token)
	}

	return tokens, nil
}

// mapToFirestoreUpdates converts a map to Firestore updates
func mapToFirestoreUpdates(updates map[string]interface{}) []firestore.Update {
	firestoreUpdates := make([]firestore.Update, 0, len(updates))
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}
	return firestoreUpdates
}
