package services

import (
	"context"
	"fmt"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/utils"
)

// OwnerService handles business logic for owner management with LGPD compliance
type OwnerService struct {
	ownerRepo       *repositories.OwnerRepository
	tenantRepo      *repositories.TenantRepository
	activityLogRepo *repositories.ActivityLogRepository
}

// NewOwnerService creates a new owner service
func NewOwnerService(
	ownerRepo *repositories.OwnerRepository,
	tenantRepo *repositories.TenantRepository,
	activityLogRepo *repositories.ActivityLogRepository,
) *OwnerService {
	return &OwnerService{
		ownerRepo:       ownerRepo,
		tenantRepo:      tenantRepo,
		activityLogRepo: activityLogRepo,
	}
}

// CreateOwner creates a new owner with validation and LGPD compliance
func (s *OwnerService) CreateOwner(ctx context.Context, owner *models.Owner) error {
	// Validate required fields
	if owner.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	// Validate tenant exists
	if _, err := s.tenantRepo.Get(ctx, owner.TenantID); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Validate document (CPF/CNPJ) if provided
	if owner.Document != "" {
		if err := s.ValidateDocument(owner); err != nil {
			return err
		}
	}

	// Validate email if provided
	if owner.Email != "" {
		if err := utils.ValidateEmail(owner.Email); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}
		owner.Email = utils.NormalizeEmail(owner.Email)
	}

	// Validate phone if provided
	if owner.Phone != "" {
		if err := utils.ValidatePhoneBR(owner.Phone); err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}
		owner.Phone = utils.NormalizePhoneBR(owner.Phone)
	}

	// Determine owner status based on data completeness
	owner.OwnerStatus = s.determineOwnerStatus(owner)

	// LGPD: Default consent_given to false for placeholders
	if !owner.ConsentGiven {
		owner.ConsentGiven = false
	}

	// LGPD: Set consent date if consent given
	if owner.ConsentGiven && owner.ConsentDate == nil {
		now := time.Now()
		owner.ConsentDate = &now
	}

	// LGPD: Initialize anonymization flags
	owner.IsAnonymized = false

	// Create owner in repository
	if err := s.ownerRepo.Create(ctx, owner); err != nil {
		return fmt.Errorf("failed to create owner: %w", err)
	}

	// Log activity based on status
	eventType := "owner_placeholder_created"
	if owner.OwnerStatus == models.OwnerStatusPartial || owner.OwnerStatus == models.OwnerStatusVerified {
		eventType = "owner_created"
	}

	_ = s.logActivity(ctx, owner.TenantID, eventType, models.ActorTypeSystem, "", map[string]interface{}{
		"owner_id":      owner.ID,
		"owner_status":  owner.OwnerStatus,
		"consent_given": owner.ConsentGiven,
	})

	return nil
}

// GetOwner retrieves an owner by ID
func (s *OwnerService) GetOwner(ctx context.Context, tenantID, id string) (*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	owner, err := s.ownerRepo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	// LGPD: Check if owner is anonymized
	if owner.IsAnonymized {
		return nil, fmt.Errorf("owner data has been anonymized")
	}

	return owner, nil
}

// UpdateOwner updates an owner with validation
func (s *OwnerService) UpdateOwner(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Validate owner exists and not anonymized
	existing, err := s.ownerRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("owner not found: %w", err)
	}

	if existing.IsAnonymized {
		return fmt.Errorf("cannot update anonymized owner")
	}

	// Validate document if being updated
	if document, ok := updates["document"].(string); ok && document != "" {
		docType := existing.DocumentType
		if dt, ok := updates["document_type"].(string); ok {
			docType = dt
		}

		if docType == "cpf" || docType == "" {
			if err := utils.ValidateCPF(document); err != nil {
				return fmt.Errorf("invalid CPF: %w", err)
			}
			updates["document"] = utils.NormalizeCPF(document)
			updates["document_type"] = "cpf"
		} else if docType == "cnpj" {
			if err := utils.ValidateCNPJ(document); err != nil {
				return fmt.Errorf("invalid CNPJ: %w", err)
			}
			updates["document"] = utils.NormalizeCNPJ(document)
			updates["document_type"] = "cnpj"
		} else {
			return fmt.Errorf("invalid document_type: must be 'cpf' or 'cnpj'")
		}
	}

	// Validate email if being updated
	if email, ok := updates["email"].(string); ok && email != "" {
		if err := utils.ValidateEmail(email); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}
		updates["email"] = utils.NormalizeEmail(email)
	}

	// Validate phone if being updated
	if phone, ok := updates["phone"].(string); ok && phone != "" {
		if err := utils.ValidatePhoneBR(phone); err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}
		updates["phone"] = utils.NormalizePhoneBR(phone)
	}

	// Prevent updating anonymization fields directly
	delete(updates, "is_anonymized")
	delete(updates, "anonymized_at")
	delete(updates, "anonymization_reason")

	// Prevent updating tenant_id
	delete(updates, "tenant_id")

	// Update owner in repository
	if err := s.ownerRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update owner: %w", err)
	}

	// Check if this is an enrichment (placeholder to partial/verified)
	eventType := "owner_updated"
	if existing.OwnerStatus == models.OwnerStatusIncomplete {
		eventType = "owner_enriched"
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, eventType, models.ActorTypeSystem, "", map[string]interface{}{
		"owner_id": id,
		"updates":  updates,
	})

	return nil
}

// DeleteOwner deletes an owner
func (s *OwnerService) DeleteOwner(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Validate owner exists
	if _, err := s.ownerRepo.Get(ctx, tenantID, id); err != nil {
		return fmt.Errorf("owner not found: %w", err)
	}

	// Delete owner from repository
	if err := s.ownerRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete owner: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "owner_deleted", models.ActorTypeSystem, "", map[string]interface{}{
		"owner_id": id,
	})

	return nil
}

// ListOwners lists all owners for a tenant with pagination
func (s *OwnerService) ListOwners(ctx context.Context, tenantID string, opts repositories.PaginationOptions) ([]*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	owners, err := s.ownerRepo.List(ctx, tenantID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list owners: %w", err)
	}

	return owners, nil
}

// ListOwnersByStatus lists owners by status for a tenant
func (s *OwnerService) ListOwnersByStatus(ctx context.Context, tenantID string, status models.OwnerStatus, opts repositories.PaginationOptions) ([]*models.Owner, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	owners, err := s.ownerRepo.ListByStatus(ctx, tenantID, status, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list owners by status: %w", err)
	}

	return owners, nil
}

// UpdateStatus updates the status of an owner
func (s *OwnerService) UpdateStatus(ctx context.Context, tenantID, id string, status models.OwnerStatus) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Validate status
	if err := s.validateOwnerStatus(status); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"owner_status": status,
	}

	if err := s.ownerRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update owner status: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "owner_status_changed", models.ActorTypeSystem, "", map[string]interface{}{
		"owner_id": id,
		"status":   status,
	})

	return nil
}

// ValidateDocument validates owner document (CPF or CNPJ)
func (s *OwnerService) ValidateDocument(owner *models.Owner) error {
	if owner.Document == "" {
		return nil // Document is optional for owners
	}

	// Determine document type if not set
	if owner.DocumentType == "" {
		// Auto-detect based on length
		cleaned := utils.NormalizeCPF(owner.Document)
		cleaned = utils.NormalizeCNPJ(cleaned)
		if len(cleaned) == 14 { // CNPJ format XX.XXX.XXX/XXXX-XX
			owner.DocumentType = "cnpj"
		} else {
			owner.DocumentType = "cpf"
		}
	}

	if owner.DocumentType == "cpf" {
		if err := utils.ValidateCPF(owner.Document); err != nil {
			return fmt.Errorf("invalid CPF: %w", err)
		}
		owner.Document = utils.NormalizeCPF(owner.Document)
	} else if owner.DocumentType == "cnpj" {
		if err := utils.ValidateCNPJ(owner.Document); err != nil {
			return fmt.Errorf("invalid CNPJ: %w", err)
		}
		owner.Document = utils.NormalizeCNPJ(owner.Document)
	} else {
		return fmt.Errorf("invalid document_type: must be 'cpf' or 'cnpj'")
	}

	return nil
}

// RevokeConsent revokes owner consent (LGPD)
func (s *OwnerService) RevokeConsent(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("owner ID is required")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"consent_revoked": true,
		"revoked_at":      now,
	}

	if err := s.ownerRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to revoke consent: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "owner_consent_revoked", models.ActorTypeOwner, id, map[string]interface{}{
		"owner_id": id,
	})

	return nil
}

// AnonymizeOwner anonymizes owner data (LGPD)
func (s *OwnerService) AnonymizeOwner(ctx context.Context, tenantID, id, reason string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("owner ID is required")
	}
	if reason == "" {
		return fmt.Errorf("anonymization reason is required")
	}

	// Validate reason
	validReasons := map[string]bool{
		"retention_policy": true,
		"user_request":     true,
	}
	if !validReasons[reason] {
		return fmt.Errorf("invalid anonymization reason: must be 'retention_policy' or 'user_request'")
	}

	// Get existing owner to preserve ID for reference
	existing, err := s.ownerRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("owner not found: %w", err)
	}

	if existing.IsAnonymized {
		return fmt.Errorf("owner is already anonymized")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"name":                  "[ANONYMIZED]",
		"email":                 "",
		"phone":                 "",
		"document":              "",
		"document_type":         "",
		"is_anonymized":         true,
		"anonymized_at":         now,
		"anonymization_reason":  reason,
		"consent_revoked":       true,
		"revoked_at":            now,
	}

	if err := s.ownerRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to anonymize owner: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "owner_anonymized", models.ActorTypeSystem, "", map[string]interface{}{
		"owner_id": id,
		"reason":   reason,
	})

	return nil
}

// determineOwnerStatus determines the owner status based on data completeness
func (s *OwnerService) determineOwnerStatus(owner *models.Owner) models.OwnerStatus {
	// Count filled fields
	filledCount := 0
	requiredFields := 5 // name, email, phone, document, consent_given

	if owner.Name != "" {
		filledCount++
	}
	if owner.Email != "" {
		filledCount++
	}
	if owner.Phone != "" {
		filledCount++
	}
	if owner.Document != "" {
		filledCount++
	}
	if owner.ConsentGiven {
		filledCount++
	}

	// Determine status
	if filledCount == 0 {
		return models.OwnerStatusIncomplete
	} else if filledCount >= requiredFields && owner.ConsentGiven {
		return models.OwnerStatusVerified
	} else {
		return models.OwnerStatusPartial
	}
}

// validateOwnerStatus validates owner status
func (s *OwnerService) validateOwnerStatus(status models.OwnerStatus) error {
	validStatuses := map[models.OwnerStatus]bool{
		models.OwnerStatusIncomplete: true,
		models.OwnerStatusPartial:    true,
		models.OwnerStatusVerified:   true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid owner status: must be 'incomplete', 'partial', or 'verified'")
	}

	return nil
}

// logActivity logs an activity (helper method)
func (s *OwnerService) logActivity(ctx context.Context, tenantID, eventType string, actorType models.ActorType, actorID string, metadata map[string]interface{}) error {
	log := &models.ActivityLog{
		TenantID:  tenantID,
		EventType: eventType,
		ActorType: actorType,
		ActorID:   actorID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	return s.activityLogRepo.Create(ctx, log)
}
