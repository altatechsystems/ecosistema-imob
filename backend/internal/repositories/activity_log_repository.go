package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

// ActivityLogRepository handles Firestore operations for activity logs
type ActivityLogRepository struct {
	*BaseRepository
}

// NewActivityLogRepository creates a new activity log repository
func NewActivityLogRepository(client *firestore.Client) *ActivityLogRepository {
	return &ActivityLogRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getActivityLogsCollection returns the collection path for activity logs within a tenant
func (r *ActivityLogRepository) getActivityLogsCollection(tenantID string) string {
	return fmt.Sprintf("tenants/%s/activity_logs", tenantID)
}

// ActivityLogFilters contains optional filters for activity log queries
type ActivityLogFilters struct {
	EventType string
	ActorType *models.ActorType
	ActorID   string
	StartDate *time.Time
	EndDate   *time.Time
}

// Create creates a new activity log entry
func (r *ActivityLogRepository) Create(ctx context.Context, log *models.ActivityLog) error {
	if log.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if log.EventType == "" {
		return fmt.Errorf("%w: event_type is required", ErrInvalidInput)
	}

	if log.ID == "" {
		log.ID = r.GenerateID(r.getActivityLogsCollection(log.TenantID))
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	collectionPath := r.getActivityLogsCollection(log.TenantID)

	// Use Set instead of Create to allow idempotent operations
	// This is important for event deduplication based on event_id
	if err := r.SetDocument(ctx, collectionPath, log.ID, log); err != nil {
		return fmt.Errorf("failed to create activity log: %w", err)
	}

	return nil
}

// Get retrieves an activity log by ID
func (r *ActivityLogRepository) Get(ctx context.Context, tenantID, id string) (*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	var log models.ActivityLog
	collectionPath := r.getActivityLogsCollection(tenantID)
	if err := r.GetDocument(ctx, collectionPath, id, &log); err != nil {
		return nil, err
	}

	log.ID = id
	return &log, nil
}

// GetByEventID retrieves an activity log by event ID (for deduplication)
func (r *ActivityLogRepository) GetByEventID(ctx context.Context, tenantID, eventID string) (*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if eventID == "" {
		return nil, fmt.Errorf("%w: event_id is required", ErrInvalidInput)
	}

	collectionPath := r.getActivityLogsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("event_id", "==", eventID).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query activity log by event_id: %w", err)
	}

	var log models.ActivityLog
	if err := doc.DataTo(&log); err != nil {
		return nil, fmt.Errorf("failed to decode activity log: %w", err)
	}

	log.ID = doc.Ref.ID
	return &log, nil
}

// GetByRequestID retrieves activity logs by request ID
func (r *ActivityLogRepository) GetByRequestID(ctx context.Context, tenantID, requestID string) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if requestID == "" {
		return nil, fmt.Errorf("%w: request_id is required", ErrInvalidInput)
	}

	collectionPath := r.getActivityLogsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("request_id", "==", requestID).
		OrderBy("timestamp", firestore.Desc)

	iter := query.Documents(ctx)
	defer iter.Stop()

	logs := make([]*models.ActivityLog, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate activity logs by request_id: %w", err)
		}

		var log models.ActivityLog
		if err := doc.DataTo(&log); err != nil {
			return nil, fmt.Errorf("failed to decode activity log: %w", err)
		}

		log.ID = doc.Ref.ID
		logs = append(logs, &log)
	}

	return logs, nil
}

// List retrieves activity logs for a tenant with optional filters and pagination
func (r *ActivityLogRepository) List(ctx context.Context, tenantID string, filters *ActivityLogFilters, opts PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getActivityLogsCollection(tenantID)
	query := r.Client().Collection(collectionPath).Query

	// Apply filters if provided
	if filters != nil {
		if filters.EventType != "" {
			query = query.Where("event_type", "==", filters.EventType)
		}
		if filters.ActorType != nil {
			query = query.Where("actor_type", "==", string(*filters.ActorType))
		}
		if filters.ActorID != "" {
			query = query.Where("actor_id", "==", filters.ActorID)
		}
		if filters.StartDate != nil {
			query = query.Where("timestamp", ">=", *filters.StartDate)
		}
		if filters.EndDate != nil {
			query = query.Where("timestamp", "<=", *filters.EndDate)
		}
	}

	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	logs := make([]*models.ActivityLog, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate activity logs: %w", err)
		}

		var log models.ActivityLog
		if err := doc.DataTo(&log); err != nil {
			return nil, fmt.Errorf("failed to decode activity log: %w", err)
		}

		log.ID = doc.Ref.ID
		logs = append(logs, &log)
	}

	return logs, nil
}

// ListByEventType retrieves activity logs by event type
func (r *ActivityLogRepository) ListByEventType(ctx context.Context, tenantID, eventType string, opts PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if eventType == "" {
		return nil, fmt.Errorf("%w: event_type is required", ErrInvalidInput)
	}

	filters := &ActivityLogFilters{EventType: eventType}
	return r.List(ctx, tenantID, filters, opts)
}

// ListByActor retrieves activity logs by actor
func (r *ActivityLogRepository) ListByActor(ctx context.Context, tenantID string, actorType models.ActorType, actorID string, opts PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	filters := &ActivityLogFilters{
		ActorType: &actorType,
		ActorID:   actorID,
	}
	return r.List(ctx, tenantID, filters, opts)
}

// ListByDateRange retrieves activity logs within a date range
func (r *ActivityLogRepository) ListByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, opts PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	filters := &ActivityLogFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	return r.List(ctx, tenantID, filters, opts)
}

// ListForEntity retrieves activity logs for a specific entity
func (r *ActivityLogRepository) ListForEntity(ctx context.Context, tenantID, entityID string, opts PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if entityID == "" {
		return nil, fmt.Errorf("%w: entity_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	collectionPath := r.getActivityLogsCollection(tenantID)

	// Query for logs where the entity_id is in the metadata
	// Note: This requires a composite index on (metadata.property_id, timestamp) etc.
	// For MVP, we might query all logs and filter in memory, or use specific event types
	query := r.Client().Collection(collectionPath).Query
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	logs := make([]*models.ActivityLog, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate activity logs for entity: %w", err)
		}

		var log models.ActivityLog
		if err := doc.DataTo(&log); err != nil {
			return nil, fmt.Errorf("failed to decode activity log: %w", err)
		}

		// Check if entity_id is in metadata
		if log.Metadata != nil {
			for _, value := range log.Metadata {
				if strValue, ok := value.(string); ok && strValue == entityID {
					log.ID = doc.Ref.ID
					logs = append(logs, &log)
					break
				}
			}
		}
	}

	return logs, nil
}

// ListPropertyLogs retrieves activity logs for a specific property
func (r *ActivityLogRepository) ListPropertyLogs(ctx context.Context, tenantID, propertyID string, opts PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if propertyID == "" {
		return nil, fmt.Errorf("%w: property_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	// Query logs where metadata.property_id == propertyID
	// This requires a composite index
	collectionPath := r.getActivityLogsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("metadata.property_id", "==", propertyID)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	logs := make([]*models.ActivityLog, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate property logs: %w", err)
		}

		var log models.ActivityLog
		if err := doc.DataTo(&log); err != nil {
			return nil, fmt.Errorf("failed to decode activity log: %w", err)
		}

		log.ID = doc.Ref.ID
		logs = append(logs, &log)
	}

	return logs, nil
}

// ListLeadLogs retrieves activity logs for a specific lead
func (r *ActivityLogRepository) ListLeadLogs(ctx context.Context, tenantID, leadID string, opts PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}
	if leadID == "" {
		return nil, fmt.Errorf("%w: lead_id is required", ErrInvalidInput)
	}

	if opts.Limit == 0 {
		opts = DefaultPaginationOptions()
	}

	// Query logs where metadata.lead_id == leadID
	// This requires a composite index
	collectionPath := r.getActivityLogsCollection(tenantID)
	query := r.Client().Collection(collectionPath).
		Where("metadata.lead_id", "==", leadID)
	query = r.ApplyPagination(query, opts)

	iter := query.Documents(ctx)
	defer iter.Stop()

	logs := make([]*models.ActivityLog, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate lead logs: %w", err)
		}

		var log models.ActivityLog
		if err := doc.DataTo(&log); err != nil {
			return nil, fmt.Errorf("failed to decode activity log: %w", err)
		}

		log.ID = doc.Ref.ID
		logs = append(logs, &log)
	}

	return logs, nil
}

// Delete deletes an activity log (should be rare - logs are typically immutable)
func (r *ActivityLogRepository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidInput)
	}

	collectionPath := r.getActivityLogsCollection(tenantID)
	if err := r.DeleteDocument(ctx, collectionPath, id); err != nil {
		return fmt.Errorf("failed to delete activity log: %w", err)
	}
	return nil
}
