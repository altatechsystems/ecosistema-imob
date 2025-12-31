package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
)

// OwnerConfirmationService handles owner confirmation token logic
type OwnerConfirmationService struct {
	tokenRepo       *repositories.OwnerConfirmationTokenRepository
	propertyRepo    *repositories.PropertyRepository
	ownerRepo       *repositories.OwnerRepository
	activityLogRepo *repositories.ActivityLogRepository
}

// NewOwnerConfirmationService creates a new owner confirmation service
func NewOwnerConfirmationService(
	tokenRepo *repositories.OwnerConfirmationTokenRepository,
	propertyRepo *repositories.PropertyRepository,
	ownerRepo *repositories.OwnerRepository,
	activityLogRepo *repositories.ActivityLogRepository,
) *OwnerConfirmationService {
	return &OwnerConfirmationService{
		tokenRepo:       tokenRepo,
		propertyRepo:    propertyRepo,
		ownerRepo:       ownerRepo,
		activityLogRepo: activityLogRepo,
	}
}

// GenerateOwnerConfirmationLink generates a secure confirmation link for property owner
// Returns: confirmationURL, tokenID, expiresAt, error
func (s *OwnerConfirmationService) GenerateOwnerConfirmationLink(
	ctx context.Context,
	tenantID string,
	propertyID string,
	actorID string,
	ownerID *string,
	deliveryHint string,
) (string, string, time.Time, error) {
	// Verify property exists
	property, err := s.propertyRepo.Get(ctx, tenantID, propertyID)
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Generate secure random token (32 bytes = 256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate random token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create SHA-256 hash of the token (store hash, not plain token)
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	// Set expiration (7 days from now)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Get owner snapshot if owner exists and is not incomplete
	var ownerSnapshot *models.OwnerSnapshotMinimal
	if ownerID != nil && *ownerID != "" {
		owner, err := s.ownerRepo.Get(ctx, tenantID, *ownerID)
		if err == nil && owner.OwnerStatus != models.OwnerStatusIncomplete {
			ownerSnapshot = &models.OwnerSnapshotMinimal{
				Name:  maskName(owner.Name),
				Phone: maskPhone(owner.Phone),
				Email: maskEmail(owner.Email),
			}
		}
	} else if property.OwnerID != "" {
		// Try to get owner from property
		owner, err := s.ownerRepo.Get(ctx, tenantID, property.OwnerID)
		if err == nil && owner.OwnerStatus != models.OwnerStatusIncomplete {
			ownerSnapshot = &models.OwnerSnapshotMinimal{
				Name:  maskName(owner.Name),
				Phone: maskPhone(owner.Phone),
				Email: maskEmail(owner.Email),
			}
			ownerIDCopy := property.OwnerID
			ownerID = &ownerIDCopy
		}
	}

	// Create token record
	confirmationToken := &models.OwnerConfirmationToken{
		TenantID:           tenantID,
		PropertyID:         propertyID,
		OwnerID:            ownerID,
		TokenHash:          tokenHash,
		ExpiresAt:          expiresAt,
		CreatedByActorID:   actorID,
		CreatedByActorType: models.ActorTypeUser,
		DeliveryHint:       deliveryHint,
		OwnerSnapshot:      ownerSnapshot,
	}

	if err := s.tokenRepo.Create(ctx, confirmationToken); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to create confirmation token: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "owner_confirmation_link_created", models.ActorTypeUser, actorID, map[string]interface{}{
		"property_id":    propertyID,
		"token_id":       confirmationToken.ID,
		"expires_at":     expiresAt,
		"delivery_hint":  deliveryHint,
		"owner_id":       ownerID,
		"owner_complete": ownerSnapshot != nil,
	})

	// Build confirmation URL
	// TODO: Get base URL from config/env
	baseURL := "http://localhost:3000" // Change this to actual domain
	confirmationURL := fmt.Sprintf("%s/confirmar/%s", baseURL, token)

	return confirmationURL, confirmationToken.ID, expiresAt, nil
}

// GetConfirmationPageResponse represents the token validation response
type GetConfirmationPageResponse struct {
	Valid         bool    `json:"valid"`
	PropertyID    string  `json:"property_id,omitempty"`
	PropertyType  string  `json:"property_type,omitempty"`
	Neighborhood  string  `json:"neighborhood,omitempty"`
	City          string  `json:"city,omitempty"`
	Reference     string  `json:"reference,omitempty"`
	CurrentStatus string  `json:"current_status,omitempty"`
	CurrentPrice  float64 `json:"current_price,omitempty"`
	ExpiresAt     string  `json:"expires_at,omitempty"`
	Error         string  `json:"error,omitempty"`
}

// ValidateTokenAndGetPropertyInfo validates token and returns minimal property info for display
// NOTE: This requires tenant_id to be extracted from the token or passed separately
// For MVP, we'll pass tenant_id from the handler
func (s *OwnerConfirmationService) ValidateTokenAndGetPropertyInfo(
	ctx context.Context,
	tenantID string,
	token string,
) (*GetConfirmationPageResponse, error) {
	// Hash the provided token
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	// Find token by hash
	confirmationToken, err := s.tokenRepo.GetByTokenHash(ctx, tenantID, tokenHash)
	if err != nil {
		return &GetConfirmationPageResponse{
			Valid: false,
			Error: "Token não encontrado ou expirado",
		}, nil
	}

	// Check expiration
	if time.Now().After(confirmationToken.ExpiresAt) {
		return &GetConfirmationPageResponse{
			Valid: false,
			Error: "Link expirado. Solicite um novo ao corretor.",
		}, nil
	}

	// Check if already used
	if confirmationToken.UsedAt != nil {
		return &GetConfirmationPageResponse{
			Valid: false,
			Error: "Este link já foi utilizado.",
		}, nil
	}

	// Get property info
	property, err := s.propertyRepo.Get(ctx, tenantID, confirmationToken.PropertyID)
	if err != nil {
		return &GetConfirmationPageResponse{
			Valid: false,
			Error: "Imóvel não encontrado",
		}, nil
	}

	// Return minimal property info (don't expose sensitive data)
	return &GetConfirmationPageResponse{
		Valid:         true,
		PropertyID:    property.ID,
		PropertyType:  string(property.PropertyType),
		Neighborhood:  property.Neighborhood,
		City:          property.City,
		Reference:     property.Reference,
		CurrentStatus: string(property.Status),
		CurrentPrice:  property.PriceAmount,
		ExpiresAt:     confirmationToken.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// SubmitOwnerConfirmation processes the owner's confirmation action
func (s *OwnerConfirmationService) SubmitOwnerConfirmation(
	ctx context.Context,
	tenantID string,
	token string,
	action models.ConfirmationAction,
	priceAmount *float64,
) error {
	// Hash the provided token
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	// Find token by hash
	confirmationToken, err := s.tokenRepo.GetByTokenHash(ctx, tenantID, tokenHash)
	if err != nil {
		return fmt.Errorf("token not found or expired")
	}

	// Validate token
	if time.Now().After(confirmationToken.ExpiresAt) {
		return fmt.Errorf("token expired")
	}

	if confirmationToken.UsedAt != nil {
		return fmt.Errorf("token already used")
	}

	// Get property
	property, err := s.propertyRepo.Get(ctx, tenantID, confirmationToken.PropertyID)
	if err != nil {
		return fmt.Errorf("property not found")
	}

	// Prepare updates
	updates := make(map[string]interface{})
	now := time.Now()
	metadata := map[string]interface{}{
		"property_id": property.ID,
		"token_id":    confirmationToken.ID,
		"action":      action,
	}

	// Process action
	switch action {
	case models.ConfirmationActionAvailable:
		updates["status"] = models.PropertyStatusAvailable
		updates["status_confirmed_at"] = now
		metadata["status"] = models.PropertyStatusAvailable

		// Log activity
		_ = s.logActivity(ctx, tenantID, "owner_confirmed_status", models.ActorTypeOwner, "", metadata)

	case models.ConfirmationActionUnavailable:
		updates["status"] = models.PropertyStatusUnavailable
		updates["status_confirmed_at"] = now
		updates["visibility"] = models.PropertyVisibilityPrivate // Hide unavailable properties
		metadata["status"] = models.PropertyStatusUnavailable

		// Log activity
		_ = s.logActivity(ctx, tenantID, "owner_confirmed_status", models.ActorTypeOwner, "", metadata)

	case models.ConfirmationActionPrice:
		if priceAmount == nil || *priceAmount <= 0 {
			return fmt.Errorf("valid price_amount is required for price confirmation")
		}
		updates["price_amount"] = *priceAmount
		updates["price_confirmed_at"] = now
		metadata["price_amount"] = *priceAmount

		// Log activity
		_ = s.logActivity(ctx, tenantID, "owner_confirmed_price", models.ActorTypeOwner, "", metadata)

	default:
		return fmt.Errorf("invalid action")
	}

	// Mark token as used
	tokenUpdates := map[string]interface{}{
		"used_at":     now,
		"last_action": string(action),
	}
	if err := s.tokenRepo.Update(ctx, tenantID, confirmationToken.ID, tokenUpdates); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Update property
	if err := s.propertyRepo.Update(ctx, tenantID, property.ID, updates); err != nil {
		return fmt.Errorf("failed to update property: %w", err)
	}

	return nil
}

// logActivity logs an activity (helper method)
func (s *OwnerConfirmationService) logActivity(
	ctx context.Context,
	tenantID string,
	eventType string,
	actorType models.ActorType,
	actorID string,
	metadata map[string]interface{},
) error {
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

// Helper functions for masking sensitive data

func maskName(name string) string {
	if name == "" {
		return ""
	}
	if len(name) <= 2 {
		return name
	}
	// "João Silva" -> "João S."
	parts := []rune(name)
	if len(parts) > 5 {
		return string(parts[0:5]) + "..."
	}
	return name
}

func maskPhone(phone string) string {
	if phone == "" {
		return ""
	}
	// "(11) 98765-4321" -> "(11) 9****-4321"
	if len(phone) > 8 {
		return phone[0:len(phone)-8] + "****" + phone[len(phone)-4:]
	}
	return "****" + phone[len(phone)-4:]
}

func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	// "joao@example.com" -> "j***@example.com"
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}

	if atIndex <= 0 {
		return email
	}

	if atIndex == 1 {
		return email[0:1] + "***" + email[atIndex:]
	}

	return email[0:1] + "***" + email[atIndex:]
}
