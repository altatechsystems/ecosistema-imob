package services

import (
	"context"
	"fmt"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/utils"
)

// LeadService handles business logic for lead management with LGPD compliance and routing
type LeadService struct {
	leadRepo        *repositories.LeadRepository
	propertyRepo    *repositories.PropertyRepository
	roleRepo        *repositories.PropertyBrokerRoleRepository
	tenantRepo      *repositories.TenantRepository
	activityLogRepo *repositories.ActivityLogRepository
}

// NewLeadService creates a new lead service
func NewLeadService(
	leadRepo *repositories.LeadRepository,
	propertyRepo *repositories.PropertyRepository,
	roleRepo *repositories.PropertyBrokerRoleRepository,
	tenantRepo *repositories.TenantRepository,
	activityLogRepo *repositories.ActivityLogRepository,
) *LeadService {
	return &LeadService{
		leadRepo:        leadRepo,
		propertyRepo:    propertyRepo,
		roleRepo:        roleRepo,
		tenantRepo:      tenantRepo,
		activityLogRepo: activityLogRepo,
	}
}

// CreateLead creates a new lead with validation, LGPD compliance, and broker routing
func (s *LeadService) CreateLead(ctx context.Context, lead *models.Lead) error {
	// Validate required fields
	if lead.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if lead.PropertyID == "" {
		return fmt.Errorf("property_id is required")
	}

	// LGPD: Consent is MANDATORY
	if !lead.ConsentGiven {
		return fmt.Errorf("consent must be given to create a lead (LGPD compliance)")
	}

	// Validate at least one contact method
	if lead.Email == "" && lead.Phone == "" {
		return fmt.Errorf("at least one contact method (email or phone) is required")
	}

	// Validate tenant exists
	if _, err := s.tenantRepo.Get(ctx, lead.TenantID); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Validate property exists
	if _, err := s.propertyRepo.Get(ctx, lead.TenantID, lead.PropertyID); err != nil {
		return fmt.Errorf("property not found: %w", err)
	}

	// Validate email if provided
	if lead.Email != "" {
		if err := utils.ValidateEmail(lead.Email); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}
		lead.Email = utils.NormalizeEmail(lead.Email)
	}

	// Validate phone if provided
	if lead.Phone != "" {
		if err := utils.ValidatePhoneBR(lead.Phone); err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}
		lead.Phone = utils.NormalizePhoneBR(lead.Phone)
	}

	// Validate channel
	if err := s.validateChannel(lead.Channel); err != nil {
		return err
	}

	// Set default status
	if lead.Status == "" {
		lead.Status = models.LeadStatusNew
	}

	// LGPD: Set consent date if not provided
	if lead.ConsentDate.IsZero() {
		lead.ConsentDate = time.Now()
	}

	// LGPD: Default consent text if not provided
	if lead.ConsentText == "" {
		lead.ConsentText = "Autorizo o uso dos meus dados pessoais para contato sobre este imóvel, conforme a Lei Geral de Proteção de Dados (LGPD)."
	}

	// LGPD: Initialize anonymization flags
	lead.ConsentRevoked = false
	lead.IsAnonymized = false

	// Create lead in repository
	if err := s.leadRepo.Create(ctx, lead); err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}

	// Log activity based on channel
	eventType := fmt.Sprintf("lead_created_%s", lead.Channel)
	_ = s.logActivity(ctx, lead.TenantID, eventType, models.ActorTypeSystem, "", map[string]interface{}{
		"lead_id":       lead.ID,
		"property_id":   lead.PropertyID,
		"channel":       lead.Channel,
		"consent_given": lead.ConsentGiven,
		"consent_ip":    lead.ConsentIP,
	})

	return nil
}

// GetLead retrieves a lead by ID
func (s *LeadService) GetLead(ctx context.Context, tenantID, id string) (*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return nil, fmt.Errorf("lead ID is required")
	}

	lead, err := s.leadRepo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	// LGPD: Check if lead is anonymized
	if lead.IsAnonymized {
		return nil, fmt.Errorf("lead data has been anonymized")
	}

	return lead, nil
}

// UpdateLead updates a lead with validation
func (s *LeadService) UpdateLead(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("lead ID is required")
	}

	// Validate lead exists and not anonymized
	existing, err := s.leadRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	if existing.IsAnonymized {
		return fmt.Errorf("cannot update anonymized lead")
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

	// Validate status if being updated
	if status, ok := updates["status"].(models.LeadStatus); ok {
		if err := s.validateStatus(status); err != nil {
			return err
		}
	}

	// Prevent updating LGPD fields directly
	delete(updates, "consent_given")
	delete(updates, "consent_date")
	delete(updates, "consent_revoked")
	delete(updates, "revoked_at")
	delete(updates, "is_anonymized")
	delete(updates, "anonymized_at")
	delete(updates, "anonymization_reason")

	// Prevent updating tenant_id and property_id
	delete(updates, "tenant_id")
	delete(updates, "property_id")

	// Update lead in repository
	if err := s.leadRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update lead: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "lead_updated", models.ActorTypeSystem, "", map[string]interface{}{
		"lead_id":     id,
		"property_id": existing.PropertyID,
		"updates":     updates,
	})

	return nil
}

// DeleteLead deletes a lead (should be rare - prefer anonymization)
func (s *LeadService) DeleteLead(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("lead ID is required")
	}

	// Validate lead exists
	lead, err := s.leadRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	// Delete lead from repository
	if err := s.leadRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "lead_deleted", models.ActorTypeSystem, "", map[string]interface{}{
		"lead_id":     id,
		"property_id": lead.PropertyID,
	})

	return nil
}

// ListLeads lists leads with filters and pagination
func (s *LeadService) ListLeads(ctx context.Context, tenantID string, filters *repositories.LeadFilters, opts repositories.PaginationOptions) ([]*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	leads, err := s.leadRepo.List(ctx, tenantID, filters, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list leads: %w", err)
	}

	return leads, nil
}

// ListLeadsByProperty lists all leads for a property
func (s *LeadService) ListLeadsByProperty(ctx context.Context, tenantID, propertyID string, opts repositories.PaginationOptions) ([]*models.Lead, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	leads, err := s.leadRepo.ListByProperty(ctx, tenantID, propertyID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list leads by property: %w", err)
	}

	return leads, nil
}

// UpdateStatus updates the status of a lead
func (s *LeadService) UpdateStatus(ctx context.Context, tenantID, id string, status models.LeadStatus) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("lead ID is required")
	}

	if err := s.validateStatus(status); err != nil {
		return err
	}

	// Get existing lead for logging
	lead, err := s.leadRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if err := s.leadRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update lead status: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "lead_status_changed", models.ActorTypeSystem, "", map[string]interface{}{
		"lead_id":     id,
		"property_id": lead.PropertyID,
		"old_status":  lead.Status,
		"new_status":  status,
	})

	return nil
}

// AssignToBroker assigns a lead to a specific broker (for manual routing)
func (s *LeadService) AssignToBroker(ctx context.Context, tenantID, leadID, brokerID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if leadID == "" {
		return fmt.Errorf("lead ID is required")
	}
	if brokerID == "" {
		return fmt.Errorf("broker ID is required")
	}

	// Get existing lead for logging
	lead, err := s.leadRepo.Get(ctx, tenantID, leadID)
	if err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	// Note: In MVP, we don't store assigned_broker_id in the Lead model
	// This is for future implementation in MVP+1
	// For now, we just log the assignment

	// Log activity
	_ = s.logActivity(ctx, tenantID, "lead_assigned_to_broker", models.ActorTypeSystem, "", map[string]interface{}{
		"lead_id":     leadID,
		"property_id": lead.PropertyID,
		"broker_id":   brokerID,
	})

	return nil
}

// RouteToAvailableBroker routes a lead to an available broker (automatic routing)
// Uses round-robin or primary broker logic
func (s *LeadService) RouteToAvailableBroker(ctx context.Context, tenantID, leadID string) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if leadID == "" {
		return "", fmt.Errorf("lead ID is required")
	}

	// Get lead
	lead, err := s.leadRepo.Get(ctx, tenantID, leadID)
	if err != nil {
		return "", fmt.Errorf("lead not found: %w", err)
	}

	// Get primary broker for the property
	primaryRole, err := s.roleRepo.GetPrimaryBroker(ctx, tenantID, lead.PropertyID)
	if err != nil {
		// If no primary broker, try to get originating broker
		if err == repositories.ErrNotFound {
			originatingRole, err := s.roleRepo.GetOriginatingBroker(ctx, tenantID, lead.PropertyID)
			if err != nil {
				return "", fmt.Errorf("no broker available for routing: %w", err)
			}
			return originatingRole.BrokerID, nil
		}
		return "", fmt.Errorf("failed to get primary broker: %w", err)
	}

	return primaryRole.BrokerID, nil
}

// RevokeConsent revokes lead consent (LGPD)
func (s *LeadService) RevokeConsent(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("lead ID is required")
	}

	// Get existing lead for logging
	lead, err := s.leadRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	if err := s.leadRepo.RevokeConsent(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to revoke consent: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "lead_consent_revoked", models.ActorTypeSystem, id, map[string]interface{}{
		"lead_id":     id,
		"property_id": lead.PropertyID,
	})

	return nil
}

// AnonymizeLead anonymizes lead data (LGPD)
func (s *LeadService) AnonymizeLead(ctx context.Context, tenantID, id, reason string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("lead ID is required")
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

	// Get existing lead for logging
	lead, err := s.leadRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	if lead.IsAnonymized {
		return fmt.Errorf("lead is already anonymized")
	}

	if err := s.leadRepo.Anonymize(ctx, tenantID, id, reason); err != nil {
		return fmt.Errorf("failed to anonymize lead: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "lead_anonymized", models.ActorTypeSystem, "", map[string]interface{}{
		"lead_id":     id,
		"property_id": lead.PropertyID,
		"reason":      reason,
	})

	return nil
}

// validateChannel validates lead channel
func (s *LeadService) validateChannel(channel models.LeadChannel) error {
	validChannels := map[models.LeadChannel]bool{
		models.LeadChannelWhatsApp: true,
		models.LeadChannelForm:     true,
		models.LeadChannelPhone:    true,
		models.LeadChannelEmail:    true,
	}

	if !validChannels[channel] {
		return fmt.Errorf("invalid lead channel")
	}

	return nil
}

// validateStatus validates lead status
func (s *LeadService) validateStatus(status models.LeadStatus) error {
	validStatuses := map[models.LeadStatus]bool{
		models.LeadStatusNew:       true,
		models.LeadStatusContacted: true,
		models.LeadStatusQualified: true,
		models.LeadStatusLost:      true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid lead status")
	}

	return nil
}

// logActivity logs an activity (helper method)
func (s *LeadService) logActivity(ctx context.Context, tenantID, eventType string, actorType models.ActorType, actorID string, metadata map[string]interface{}) error {
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
