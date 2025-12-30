package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/utils"
)

// BrokerService handles business logic for broker management
type BrokerService struct {
	brokerRepo             *repositories.BrokerRepository
	tenantRepo             *repositories.TenantRepository
	activityLogRepo        *repositories.ActivityLogRepository
	propertyBrokerRoleRepo *repositories.PropertyBrokerRoleRepository
}

// NewBrokerService creates a new broker service
func NewBrokerService(
	brokerRepo *repositories.BrokerRepository,
	tenantRepo *repositories.TenantRepository,
	activityLogRepo *repositories.ActivityLogRepository,
	propertyBrokerRoleRepo *repositories.PropertyBrokerRoleRepository,
) *BrokerService {
	return &BrokerService{
		brokerRepo:             brokerRepo,
		tenantRepo:             tenantRepo,
		activityLogRepo:        activityLogRepo,
		propertyBrokerRoleRepo: propertyBrokerRoleRepo,
	}
}

// CreateBroker creates a new broker with validation
func (s *BrokerService) CreateBroker(ctx context.Context, broker *models.Broker) error {
	// Validate required fields
	if broker.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if broker.Name == "" {
		return fmt.Errorf("broker name is required")
	}
	if broker.Email == "" {
		return fmt.Errorf("email is required")
	}
	if broker.CRECI == "" {
		return fmt.Errorf("CRECI is required")
	}
	if broker.FirebaseUID == "" {
		return fmt.Errorf("firebase_uid is required")
	}

	// Validate tenant exists
	if _, err := s.tenantRepo.Get(ctx, broker.TenantID); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Validate CRECI
	if err := s.ValidateCRECI(broker.CRECI); err != nil {
		return err
	}
	broker.CRECI = utils.NormalizeCRECI(broker.CRECI)

	// Validate email
	if err := utils.ValidateEmail(broker.Email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}
	broker.Email = utils.NormalizeEmail(broker.Email)

	// Check if email is already registered in this tenant
	existing, err := s.brokerRepo.GetByEmail(ctx, broker.TenantID, broker.Email)
	if err != nil && err != repositories.ErrNotFound {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("email '%s' is already registered in this tenant", broker.Email)
	}

	// Check if Firebase UID is already registered in this tenant
	existing, err = s.brokerRepo.GetByFirebaseUID(ctx, broker.TenantID, broker.FirebaseUID)
	if err != nil && err != repositories.ErrNotFound {
		return fmt.Errorf("failed to check firebase_uid uniqueness: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("firebase_uid '%s' is already registered in this tenant", broker.FirebaseUID)
	}

	// Validate phone if provided
	if broker.Phone != "" {
		if err := utils.ValidatePhoneBR(broker.Phone); err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}
		broker.Phone = utils.NormalizePhoneBR(broker.Phone)
	}

	// Validate document (CPF/CNPJ) if provided
	if broker.Document != "" {
		if broker.DocumentType == "cpf" || broker.DocumentType == "" {
			if err := utils.ValidateCPF(broker.Document); err != nil {
				return fmt.Errorf("invalid CPF: %w", err)
			}
			broker.Document = utils.NormalizeCPF(broker.Document)
			broker.DocumentType = "cpf"
		} else if broker.DocumentType == "cnpj" {
			if err := utils.ValidateCNPJ(broker.Document); err != nil {
				return fmt.Errorf("invalid CNPJ: %w", err)
			}
			broker.Document = utils.NormalizeCNPJ(broker.Document)
		} else {
			return fmt.Errorf("invalid document_type: must be 'cpf' or 'cnpj'")
		}
	}

	// Set defaults
	if broker.Role == "" {
		broker.Role = "broker" // default role
	}
	broker.IsActive = true

	// Validate role
	if err := s.validateRole(broker.Role); err != nil {
		return err
	}

	// Create broker in repository
	if err := s.brokerRepo.Create(ctx, broker); err != nil {
		return fmt.Errorf("failed to create broker: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, broker.TenantID, "broker_created", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id":    broker.ID,
		"name":         broker.Name,
		"email":        broker.Email,
		"creci":        broker.CRECI,
		"role":         broker.Role,
		"firebase_uid": broker.FirebaseUID,
	})

	return nil
}

// GetBroker retrieves a broker by ID
func (s *BrokerService) GetBroker(ctx context.Context, tenantID, id string) (*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return nil, fmt.Errorf("broker ID is required")
	}

	broker, err := s.brokerRepo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker: %w", err)
	}

	return broker, nil
}

// GetBrokerByFirebaseUID retrieves a broker by Firebase UID
func (s *BrokerService) GetBrokerByFirebaseUID(ctx context.Context, tenantID, firebaseUID string) (*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if firebaseUID == "" {
		return nil, fmt.Errorf("firebase_uid is required")
	}

	broker, err := s.brokerRepo.GetByFirebaseUID(ctx, tenantID, firebaseUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker by firebase_uid: %w", err)
	}

	return broker, nil
}

// GetBrokerByEmail retrieves a broker by email
func (s *BrokerService) GetBrokerByEmail(ctx context.Context, tenantID, email string) (*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	email = utils.NormalizeEmail(email)

	broker, err := s.brokerRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker by email: %w", err)
	}

	return broker, nil
}

// UpdateBroker updates a broker with validation
func (s *BrokerService) UpdateBroker(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("broker ID is required")
	}

	// Debug: log updates
	fmt.Printf("DEBUG UpdateBroker - tenantID: %s, id: %s, updates: %+v\n", tenantID, id, updates)

	// Validate broker exists
	existing, err := s.brokerRepo.Get(ctx, tenantID, id)
	if err != nil {
		fmt.Printf("DEBUG UpdateBroker - failed to get existing broker: %v\n", err)
		return fmt.Errorf("broker not found: %w", err)
	}
	fmt.Printf("DEBUG UpdateBroker - existing broker: name=%s, email=%s\n", existing.Name, existing.Email)

	// Validate CRECI if being updated (skip validation if "PENDENTE" or empty)
	if creci, ok := updates["creci"].(string); ok {
		// Only validate if CRECI is not empty and not "PENDENTE" and contains hyphen (formatted)
		if creci != "" && creci != "PENDENTE" {
			// Only validate if it looks like a formatted CRECI (contains hyphen or slash)
			if len(creci) > 5 && (creci[len(creci)-3] == '/' || strings.Contains(creci, "-")) {
				if err := s.ValidateCRECI(creci); err != nil {
					return err
				}
				updates["creci"] = utils.NormalizeCRECI(creci)
			}
			// If it's just numbers, allow it but don't normalize
			// User can update to partial CRECI number and complete later
		}
	}

	// Validate email if being updated (skip validation for @pendente.com.br temporary emails)
	if email, ok := updates["email"].(string); ok && email != "" {
		normalizedEmail := utils.NormalizeEmail(email)

		// Skip strict validation for temporary placeholder emails
		if !isPendingEmail(normalizedEmail) {
			if err := utils.ValidateEmail(email); err != nil {
				return fmt.Errorf("invalid email: %w", err)
			}
		}

		// Check email uniqueness if changed
		if normalizedEmail != existing.Email {
			existingBroker, err := s.brokerRepo.GetByEmail(ctx, tenantID, normalizedEmail)
			if err != nil && err != repositories.ErrNotFound {
				return fmt.Errorf("failed to check email uniqueness: %w", err)
			}
			if existingBroker != nil && existingBroker.ID != id {
				return fmt.Errorf("email '%s' is already registered in this tenant", normalizedEmail)
			}
		}

		updates["email"] = normalizedEmail
	}

	// Validate phone if being updated
	if phone, ok := updates["phone"].(string); ok && phone != "" {
		if err := utils.ValidatePhoneBR(phone); err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}
		updates["phone"] = utils.NormalizePhoneBR(phone)
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
		} else {
			return fmt.Errorf("invalid document_type: must be 'cpf' or 'cnpj'")
		}
	}

	// Validate role if being updated
	if role, ok := updates["role"].(string); ok {
		if err := s.validateRole(role); err != nil {
			return err
		}
	}

	// Prevent updating firebase_uid
	delete(updates, "firebase_uid")

	// Prevent updating tenant_id
	delete(updates, "tenant_id")

	// Update broker in repository
	if err := s.brokerRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update broker: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_updated", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id": id,
		"updates":   updates,
	})

	return nil
}

// DeleteBroker deletes a broker
func (s *BrokerService) DeleteBroker(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("broker ID is required")
	}

	// Validate broker exists
	if _, err := s.brokerRepo.Get(ctx, tenantID, id); err != nil {
		return fmt.Errorf("broker not found: %w", err)
	}

	// Delete broker from repository
	if err := s.brokerRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete broker: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_deleted", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id": id,
	})

	return nil
}

// ListBrokers lists all brokers for a tenant with pagination
func (s *BrokerService) ListBrokers(ctx context.Context, tenantID string, opts repositories.PaginationOptions) ([]*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	brokers, err := s.brokerRepo.List(ctx, tenantID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list brokers: %w", err)
	}

	// Enrich brokers with statistics
	if err := s.enrichBrokersWithStats(ctx, brokers); err != nil {
		// Log error but don't fail - return brokers without stats
		fmt.Printf("Warning: failed to enrich brokers with stats: %v\n", err)
	}

	return brokers, nil
}

// ListActiveBrokers lists all active brokers for a tenant
func (s *BrokerService) ListActiveBrokers(ctx context.Context, tenantID string, opts repositories.PaginationOptions) ([]*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	brokers, err := s.brokerRepo.ListActive(ctx, tenantID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list active brokers: %w", err)
	}

	return brokers, nil
}

// ListBrokersByRole lists brokers by role for a tenant
func (s *BrokerService) ListBrokersByRole(ctx context.Context, tenantID, role string, opts repositories.PaginationOptions) ([]*models.Broker, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if role == "" {
		return nil, fmt.Errorf("role is required")
	}

	if err := s.validateRole(role); err != nil {
		return nil, err
	}

	brokers, err := s.brokerRepo.ListByRole(ctx, tenantID, role, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list brokers by role: %w", err)
	}

	return brokers, nil
}

// ActivateBroker activates a broker
func (s *BrokerService) ActivateBroker(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("broker ID is required")
	}

	updates := map[string]interface{}{
		"is_active": true,
	}

	if err := s.brokerRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to activate broker: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_activated", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id": id,
	})

	return nil
}

// DeactivateBroker deactivates a broker
func (s *BrokerService) DeactivateBroker(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("broker ID is required")
	}

	updates := map[string]interface{}{
		"is_active": false,
	}

	if err := s.brokerRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to deactivate broker: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_deactivated", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id": id,
	})

	return nil
}

// AssignRole assigns a role to a broker
func (s *BrokerService) AssignRole(ctx context.Context, tenantID, brokerID, role string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if brokerID == "" {
		return fmt.Errorf("broker ID is required")
	}

	if err := s.validateRole(role); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"role": role,
	}

	if err := s.brokerRepo.Update(ctx, tenantID, brokerID, updates); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_role_assigned", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id": brokerID,
		"role":      role,
	})

	return nil
}

// ValidateCRECI validates CRECI format and returns error if invalid
func (s *BrokerService) ValidateCRECI(creci string) error {
	if err := utils.ValidateCRECI(creci); err != nil {
		return fmt.Errorf("invalid CRECI: %w", err)
	}
	return nil
}

// validateRole validates broker role
func (s *BrokerService) validateRole(role string) error {
	validRoles := map[string]bool{
		"admin":   true,
		"broker":  true,
		"manager": true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid role: must be 'admin', 'broker', or 'manager'")
	}

	return nil
}

// logActivity logs an activity (helper method)
func (s *BrokerService) logActivity(ctx context.Context, tenantID, eventType string, actorType models.ActorType, actorID string, metadata map[string]interface{}) error {
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

// isPendingEmail checks if an email is a temporary placeholder email
func isPendingEmail(email string) bool {
	return len(email) > 16 && email[len(email)-16:] == "@pendente.com.br"
}

// enrichBrokerWithStats enriches a single broker with statistics
func (s *BrokerService) enrichBrokerWithStats(ctx context.Context, broker *models.Broker) error {
	if broker == nil {
		return nil
	}

	// Count properties where broker has any role
	roles, err := s.propertyBrokerRoleRepo.ListByBroker(ctx, broker.TenantID, broker.ID, repositories.PaginationOptions{
		Limit: 1000, // Get all roles for counting
	})
	if err != nil {
		// Log error but don't fail - just return broker without stats
		fmt.Printf("Warning: failed to get property roles for broker %s: %v\n", broker.ID, err)
		return nil
	}

	// Count unique properties
	propertyMap := make(map[string]bool)
	for _, role := range roles {
		propertyMap[role.PropertyID] = true
	}

	broker.TotalListings = len(propertyMap)
	return nil
}

// enrichBrokersWithStats enriches multiple brokers with statistics
func (s *BrokerService) enrichBrokersWithStats(ctx context.Context, brokers []*models.Broker) error {
	for _, broker := range brokers {
		if err := s.enrichBrokerWithStats(ctx, broker); err != nil {
			// Continue enriching other brokers even if one fails
			continue
		}
	}
	return nil
}
