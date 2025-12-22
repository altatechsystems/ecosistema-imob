package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

// BrokerRepository handles Firestore operations for brokers
type BrokerRepository struct {
	*BaseRepository
}

// NewBrokerRepository creates a new broker repository
func NewBrokerRepository(client *firestore.Client) *BrokerRepository {
	return &BrokerRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getBrokersCollection returns the collection path for brokers within a tenant
func (r *BrokerRepository) getBrokersCollection(tenantID string) string {
	return fmt.Sprintf("tenants/%s/brokers", tenantID)
}

// Create creates a new broker
func (r *BrokerRepository) Create(ctx context.Context, broker *models.Broker) error {
	if broker.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if broker.ID == "" {
		broker.ID = r.GenerateID(r.getBrokersCollection(broker.TenantID))
	}

	now := time.Now()
	broker.CreatedAt = now
	broker.UpdatedAt = now

	collectionPath := r.getBrokersCollection(broker.TenantID)
	if err := r.CreateDocument(ctx, collectionPath, broker.ID, broker); err != nil {
		return fmt.Errorf("failed to create broker: %w", err)
	}

	return nil
}

// Get retrieves a broker by ID
func (r *BrokerRepository) Get(ctx context.Context, tenantID, id string) (*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	var broker models.Broker
	collectionPath := r.getBrokersCollection(tenantID)
	if err := r.GetDocument(ctx, collectionPath, id, &broker); err != nil {
		return nil, err
	}

	broker.ID = id
	return &broker, nil
}

// GetByFirebaseUID retrieves a broker by Firebase UID
func (r *BrokerRepository) GetByFirebaseUID(ctx context.Context, tenantID, firebaseUID string) (*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if firebaseUID == "" {
		return nil, fmt.Errorf("%w: firebase_uid is required", ErrInvalidInput)
	}

	collectionPath := r.getBrokersCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("firebase_uid", "==", firebaseUID).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query broker by firebase_uid: %w", err)
	}

	var broker models.Broker
	if err := doc.DataTo(&broker); err != nil {
		return nil, fmt.Errorf("failed to decode broker: %w", err)
	}

	broker.ID = doc.Ref.ID
	return &broker, nil
}

// GetByEmail retrieves a broker by email
func (r *BrokerRepository) GetByEmail(ctx context.Context, tenantID, email string) (*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	collectionPath := r.getBrokersCollection(tenantID)
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
		return nil, fmt.Errorf("failed to query broker by email: %w", err)
	}

	var broker models.Broker
	if err := doc.DataTo(&broker); err != nil {
		return nil, fmt.Errorf("failed to decode broker: %w", err)
	}

	broker.ID = doc.Ref.ID
	return &broker, nil
}

// Update updates a broker
func (r *BrokerRepository) Update(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if id == "" {
		return fmt.Errorf("%w: broker ID is required", ErrInvalidInput)
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

	collectionPath := r.getBrokersCollection(tenantID)
	if err := r.UpdateDocument(ctx, collectionPath, id, firestoreUpdates); err != nil {
		return fmt.Errorf("failed to update broker: %w", err)
	}

	return nil
}

// Delete deletes a broker
func (r *BrokerRepository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	collectionPath := r.getBrokersCollection(tenantID)
	if err := r.DeleteDocument(ctx, collectionPath, id); err != nil {
		return fmt.Errorf("failed to delete broker: %w", err)
	}
	return nil
}

// List retrieves all brokers for a tenant with pagination
func (r *BrokerRepository) List(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getBrokersCollection(tenantID)
	query := r.Client().Collection(collectionPath).Query
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	brokers := make([]*models.Broker, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate brokers: %w", err)
		}

		var broker models.Broker
		if err := doc.DataTo(&broker); err != nil {
			return nil, fmt.Errorf("failed to decode broker: %w", err)
		}

		broker.ID = doc.Ref.ID
		brokers = append(brokers, &broker)
	}

	return brokers, nil
}

// ListActive retrieves all active brokers for a tenant
func (r *BrokerRepository) ListActive(ctx context.Context, tenantID string, opts PaginationOptions) ([]*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getBrokersCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("is_active", "==", true)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	brokers := make([]*models.Broker, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate active brokers: %w", err)
		}

		var broker models.Broker
		if err := doc.DataTo(&broker); err != nil {
			return nil, fmt.Errorf("failed to decode broker: %w", err)
		}

		broker.ID = doc.Ref.ID
		brokers = append(brokers, &broker)
	}

	return brokers, nil
}

// ListByRole retrieves brokers by role for a tenant
func (r *BrokerRepository) ListByRole(ctx context.Context, tenantID, role string, opts PaginationOptions) ([]*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if role == "" {
		return nil, fmt.Errorf("%w: role is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getBrokersCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("role", "==", role)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	brokers := make([]*models.Broker, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate brokers by role: %w", err)
		}

		var broker models.Broker
		if err := doc.DataTo(&broker); err != nil {
			return nil, fmt.Errorf("failed to decode broker: %w", err)
		}

		broker.ID = doc.Ref.ID
		brokers = append(brokers, &broker)
	}

	return brokers, nil
}
