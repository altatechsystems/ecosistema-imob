package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/google/uuid"
)

// ActivityLogService handles business logic for activity logging
type ActivityLogService struct {
	activityLogRepo *repositories.ActivityLogRepository
	tenantRepo      *repositories.TenantRepository
}

// NewActivityLogService creates a new activity log service
func NewActivityLogService(
	activityLogRepo *repositories.ActivityLogRepository,
	tenantRepo *repositories.TenantRepository,
) *ActivityLogService {
	return &ActivityLogService{
		activityLogRepo: activityLogRepo,
		tenantRepo:      tenantRepo,
	}
}

// LogActivity creates a general activity log entry
func (s *ActivityLogService) LogActivity(ctx context.Context, log *models.ActivityLog) error {
	// Validate required fields
	if log.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if log.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	// Validate tenant exists
	if _, err := s.tenantRepo.Get(ctx, log.TenantID); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// Generate event_id for deduplication (5-minute bucket)
	if log.EventID == "" {
		log.EventID = s.generateEventID(log)
	}

	// Generate event_hash for payload verification
	if log.EventHash == "" {
		log.EventHash = s.generateEventHash(log)
	}

	// Generate request_id if not provided
	if log.RequestID == "" {
		log.RequestID = uuid.New().String()
	}

	// Create activity log in repository
	if err := s.activityLogRepo.Create(ctx, log); err != nil {
		// Ignore duplicate errors (idempotent logging)
		if err != repositories.ErrAlreadyExists {
			return fmt.Errorf("failed to log activity: %w", err)
		}
	}

	return nil
}

// LogPropertyActivity logs a property-specific activity
func (s *ActivityLogService) LogPropertyActivity(ctx context.Context, tenantID, eventType, propertyID string, actorType models.ActorType, actorID string, additionalMetadata map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if propertyID == "" {
		return fmt.Errorf("property_id is required")
	}

	// Build metadata
	metadata := map[string]interface{}{
		"property_id": propertyID,
	}

	// Merge additional metadata
	if additionalMetadata != nil {
		for key, value := range additionalMetadata {
			metadata[key] = value
		}
	}

	// Create log entry
	log := &models.ActivityLog{
		TenantID:  tenantID,
		EventType: eventType,
		ActorType: actorType,
		ActorID:   actorID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	return s.LogActivity(ctx, log)
}

// LogLeadActivity logs a lead-specific activity
func (s *ActivityLogService) LogLeadActivity(ctx context.Context, tenantID, eventType, leadID, propertyID string, actorType models.ActorType, actorID string, additionalMetadata map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if leadID == "" {
		return fmt.Errorf("lead_id is required")
	}
	if propertyID == "" {
		return fmt.Errorf("property_id is required")
	}

	// Build metadata
	metadata := map[string]interface{}{
		"lead_id":     leadID,
		"property_id": propertyID,
	}

	// Merge additional metadata
	if additionalMetadata != nil {
		for key, value := range additionalMetadata {
			metadata[key] = value
		}
	}

	// Create log entry
	log := &models.ActivityLog{
		TenantID:  tenantID,
		EventType: eventType,
		ActorType: actorType,
		ActorID:   actorID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	return s.LogActivity(ctx, log)
}

// LogBrokerActivity logs a broker-specific activity
func (s *ActivityLogService) LogBrokerActivity(ctx context.Context, tenantID, eventType, brokerID string, actorType models.ActorType, actorID string, additionalMetadata map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if brokerID == "" {
		return fmt.Errorf("broker_id is required")
	}

	// Build metadata
	metadata := map[string]interface{}{
		"broker_id": brokerID,
	}

	// Merge additional metadata
	if additionalMetadata != nil {
		for key, value := range additionalMetadata {
			metadata[key] = value
		}
	}

	// Create log entry
	log := &models.ActivityLog{
		TenantID:  tenantID,
		EventType: eventType,
		ActorType: actorType,
		ActorID:   actorID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	return s.LogActivity(ctx, log)
}

// GetActivityLogs retrieves activity logs with filters and pagination
func (s *ActivityLogService) GetActivityLogs(ctx context.Context, tenantID string, filters *repositories.ActivityLogFilters, opts repositories.PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	logs, err := s.activityLogRepo.List(ctx, tenantID, filters, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs: %w", err)
	}

	return logs, nil
}

// GetActivityLogsByEventType retrieves activity logs by event type
func (s *ActivityLogService) GetActivityLogsByEventType(ctx context.Context, tenantID, eventType string, opts repositories.PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if eventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	logs, err := s.activityLogRepo.ListByEventType(ctx, tenantID, eventType, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs by event type: %w", err)
	}

	return logs, nil
}

// GetActivityLogsByActor retrieves activity logs by actor
func (s *ActivityLogService) GetActivityLogsByActor(ctx context.Context, tenantID string, actorType models.ActorType, actorID string, opts repositories.PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	logs, err := s.activityLogRepo.ListByActor(ctx, tenantID, actorType, actorID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs by actor: %w", err)
	}

	return logs, nil
}

// GetActivityLogsByDateRange retrieves activity logs within a date range
func (s *ActivityLogService) GetActivityLogsByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, opts repositories.PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	logs, err := s.activityLogRepo.ListByDateRange(ctx, tenantID, startDate, endDate, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs by date range: %w", err)
	}

	return logs, nil
}

// GetEntityTimeline retrieves the timeline (all logs) for a specific entity
func (s *ActivityLogService) GetEntityTimeline(ctx context.Context, tenantID, entityID string, opts repositories.PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if entityID == "" {
		return nil, fmt.Errorf("entity_id is required")
	}

	logs, err := s.activityLogRepo.ListForEntity(ctx, tenantID, entityID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity timeline: %w", err)
	}

	return logs, nil
}

// GetPropertyTimeline retrieves the timeline for a property
func (s *ActivityLogService) GetPropertyTimeline(ctx context.Context, tenantID, propertyID string, opts repositories.PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	logs, err := s.activityLogRepo.ListPropertyLogs(ctx, tenantID, propertyID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get property timeline: %w", err)
	}

	return logs, nil
}

// GetLeadTimeline retrieves the timeline for a lead
func (s *ActivityLogService) GetLeadTimeline(ctx context.Context, tenantID, leadID string, opts repositories.PaginationOptions) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if leadID == "" {
		return nil, fmt.Errorf("lead_id is required")
	}

	logs, err := s.activityLogRepo.ListLeadLogs(ctx, tenantID, leadID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead timeline: %w", err)
	}

	return logs, nil
}

// GetActivityLogsByRequestID retrieves all logs for a specific request
func (s *ActivityLogService) GetActivityLogsByRequestID(ctx context.Context, tenantID, requestID string) ([]*models.ActivityLog, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}

	logs, err := s.activityLogRepo.GetByRequestID(ctx, tenantID, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs by request_id: %w", err)
	}

	return logs, nil
}

// generateEventID generates a deterministic event ID for deduplication
// Format: hash(entityId + action + timestamp_bucket_5min)
func (s *ActivityLogService) generateEventID(log *models.ActivityLog) string {
	// Round timestamp to 5-minute bucket
	bucket := log.Timestamp.Truncate(5 * time.Minute)

	// Extract entity ID from metadata (if available)
	var entityID string
	if log.Metadata != nil {
		// Try to get property_id, lead_id, broker_id, etc.
		if id, ok := log.Metadata["property_id"].(string); ok {
			entityID = id
		} else if id, ok := log.Metadata["lead_id"].(string); ok {
			entityID = id
		} else if id, ok := log.Metadata["broker_id"].(string); ok {
			entityID = id
		} else if id, ok := log.Metadata["listing_id"].(string); ok {
			entityID = id
		}
	}

	// Combine components
	combined := fmt.Sprintf("%s|%s|%s|%d", log.TenantID, entityID, log.EventType, bucket.Unix())

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)[:32] // Use first 32 chars
}

// generateEventHash generates a hash of the event payload for verification
func (s *ActivityLogService) generateEventHash(log *models.ActivityLog) string {
	// Create a normalized representation of the log
	// Note: This is a simplified version. In production, you'd want to use
	// a more robust serialization method
	combined := fmt.Sprintf("%s|%s|%s|%s|%v",
		log.TenantID,
		log.EventType,
		log.ActorType,
		log.ActorID,
		log.Metadata,
	)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}
