package services

import (
	"context"
	"fmt"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
)

// ListingService handles business logic for listing management with canonical logic
type ListingService struct {
	listingRepo     *repositories.ListingRepository
	propertyRepo    *repositories.PropertyRepository
	brokerRepo      *repositories.BrokerRepository
	tenantRepo      *repositories.TenantRepository
	activityLogRepo *repositories.ActivityLogRepository
}

// NewListingService creates a new listing service
func NewListingService(
	listingRepo *repositories.ListingRepository,
	propertyRepo *repositories.PropertyRepository,
	brokerRepo *repositories.BrokerRepository,
	tenantRepo *repositories.TenantRepository,
	activityLogRepo *repositories.ActivityLogRepository,
) *ListingService {
	return &ListingService{
		listingRepo:     listingRepo,
		propertyRepo:    propertyRepo,
		brokerRepo:      brokerRepo,
		tenantRepo:      tenantRepo,
		activityLogRepo: activityLogRepo,
	}
}

// CreateListing creates a new listing with validation and canonical logic
func (s *ListingService) CreateListing(ctx context.Context, listing *models.Listing) error {
	// Validate required fields
	if listing.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if listing.PropertyID == "" {
		return fmt.Errorf("property_id is required")
	}
	if listing.BrokerID == "" {
		return fmt.Errorf("broker_id is required")
	}
	if listing.Title == "" {
		return fmt.Errorf("title is required")
	}
	if listing.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Validate tenant exists
	if _, err := s.tenantRepo.Get(ctx, listing.TenantID); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Validate property exists
	property, err := s.propertyRepo.Get(ctx, listing.TenantID, listing.PropertyID)
	if err != nil {
		return fmt.Errorf("property not found: %w", err)
	}

	// Validate broker exists
	if _, err := s.brokerRepo.Get(ctx, listing.TenantID, listing.BrokerID); err != nil {
		return fmt.Errorf("broker not found: %w", err)
	}

	// Set defaults
	listing.IsActive = true

	// Check if this is the first listing for the property
	existingListings, err := s.listingRepo.ListByProperty(ctx, listing.TenantID, listing.PropertyID, repositories.PaginationOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to check existing listings: %w", err)
	}

	// If this is the first listing, make it canonical
	if len(existingListings) == 0 {
		listing.IsCanonical = true
	} else {
		listing.IsCanonical = false
	}

	// Initialize photo and video slices if nil
	if listing.Photos == nil {
		listing.Photos = []models.Photo{}
	}
	if listing.Videos == nil {
		listing.Videos = []models.Video{}
	}

	// Create listing in repository
	if err := s.listingRepo.Create(ctx, listing); err != nil {
		return fmt.Errorf("failed to create listing: %w", err)
	}

	// If this is canonical, update the property's canonical_listing_id
	if listing.IsCanonical {
		propertyUpdates := map[string]interface{}{
			"canonical_listing_id": listing.ID,
		}
		if err := s.propertyRepo.Update(ctx, listing.TenantID, listing.PropertyID, propertyUpdates); err != nil {
			return fmt.Errorf("failed to update property canonical listing: %w", err)
		}

		// Log canonical assignment
		_ = s.logActivity(ctx, listing.TenantID, "canonical_listing_assigned", models.ActorTypeSystem, "", map[string]interface{}{
			"property_id": listing.PropertyID,
			"listing_id":  listing.ID,
			"broker_id":   listing.BrokerID,
		})
	}

	// Log activity
	_ = s.logActivity(ctx, listing.TenantID, "listing_created", models.ActorTypeSystem, "", map[string]interface{}{
		"listing_id":   listing.ID,
		"property_id":  listing.PropertyID,
		"broker_id":    listing.BrokerID,
		"is_canonical": listing.IsCanonical,
	})

	_ = property // silence unused warning

	return nil
}

// GetListing retrieves a listing by ID
func (s *ListingService) GetListing(ctx context.Context, tenantID, id string) (*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return nil, fmt.Errorf("listing ID is required")
	}

	listing, err := s.listingRepo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}

	return listing, nil
}

// GetCanonicalListingForProperty retrieves the canonical listing for a property
func (s *ListingService) GetCanonicalListingForProperty(ctx context.Context, tenantID, propertyID string) (*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	listing, err := s.listingRepo.GetCanonicalForProperty(ctx, tenantID, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get canonical listing: %w", err)
	}

	return listing, nil
}

// UpdateListing updates a listing with validation
func (s *ListingService) UpdateListing(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("listing ID is required")
	}

	// Validate listing exists
	existing, err := s.listingRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}

	// Prevent updating tenant_id, property_id, and broker_id
	delete(updates, "tenant_id")
	delete(updates, "property_id")
	delete(updates, "broker_id")

	// Prevent directly updating is_canonical (use SetCanonical instead)
	delete(updates, "is_canonical")

	// Update listing in repository
	if err := s.listingRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update listing: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "listing_updated", models.ActorTypeSystem, "", map[string]interface{}{
		"listing_id":  id,
		"property_id": existing.PropertyID,
		"updates":     updates,
	})

	return nil
}

// DeleteListing deletes a listing
func (s *ListingService) DeleteListing(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("listing ID is required")
	}

	// Validate listing exists
	existing, err := s.listingRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}

	// If this is canonical, we need to assign a new canonical listing
	if existing.IsCanonical {
		// Find other listings for the same property
		otherListings, err := s.listingRepo.ListByProperty(ctx, tenantID, existing.PropertyID, repositories.PaginationOptions{Limit: 10})
		if err != nil {
			return fmt.Errorf("failed to find other listings: %w", err)
		}

		// Find the next listing to promote (excluding the one being deleted)
		var newCanonicalID string
		for _, listing := range otherListings {
			if listing.ID != id && listing.IsActive {
				newCanonicalID = listing.ID
				break
			}
		}

		// If found, promote it to canonical
		if newCanonicalID != "" {
			if err := s.SetCanonical(ctx, tenantID, newCanonicalID); err != nil {
				return fmt.Errorf("failed to promote new canonical listing: %w", err)
			}
		} else {
			// No other active listing, clear property's canonical_listing_id
			propertyUpdates := map[string]interface{}{
				"canonical_listing_id": "",
			}
			if err := s.propertyRepo.Update(ctx, tenantID, existing.PropertyID, propertyUpdates); err != nil {
				return fmt.Errorf("failed to clear property canonical listing: %w", err)
			}
		}
	}

	// Delete listing from repository
	if err := s.listingRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete listing: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "listing_deleted", models.ActorTypeSystem, "", map[string]interface{}{
		"listing_id":   id,
		"property_id":  existing.PropertyID,
		"was_canonical": existing.IsCanonical,
	})

	return nil
}

// SetCanonical sets a listing as the canonical listing for its property
// This ensures only one canonical listing exists per property
func (s *ListingService) SetCanonical(ctx context.Context, tenantID, listingID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if listingID == "" {
		return fmt.Errorf("listing ID is required")
	}

	// Validate listing exists
	listing, err := s.listingRepo.Get(ctx, tenantID, listingID)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}

	// Check if already canonical
	if listing.IsCanonical {
		return nil // Already canonical, nothing to do
	}

	// Get current canonical listing for the property
	currentCanonical, err := s.listingRepo.GetCanonicalForProperty(ctx, tenantID, listing.PropertyID)
	if err != nil && err != repositories.ErrNotFound {
		return fmt.Errorf("failed to get current canonical listing: %w", err)
	}

	// Unset all canonical flags for this property
	if err := s.listingRepo.UnsetCanonicalForProperty(ctx, tenantID, listing.PropertyID); err != nil {
		return fmt.Errorf("failed to unset canonical flags: %w", err)
	}

	// Set the new canonical listing
	updates := map[string]interface{}{
		"is_canonical": true,
	}
	if err := s.listingRepo.Update(ctx, tenantID, listingID, updates); err != nil {
		return fmt.Errorf("failed to set canonical flag: %w", err)
	}

	// Update property's canonical_listing_id
	propertyUpdates := map[string]interface{}{
		"canonical_listing_id": listingID,
	}
	if err := s.propertyRepo.Update(ctx, tenantID, listing.PropertyID, propertyUpdates); err != nil {
		return fmt.Errorf("failed to update property canonical listing: %w", err)
	}

	// Log activity
	metadata := map[string]interface{}{
		"property_id":     listing.PropertyID,
		"new_listing_id":  listingID,
	}
	if currentCanonical != nil {
		metadata["old_listing_id"] = currentCanonical.ID
		_ = s.logActivity(ctx, tenantID, "canonical_listing_changed", models.ActorTypeSystem, "", metadata)
	} else {
		_ = s.logActivity(ctx, tenantID, "canonical_listing_assigned", models.ActorTypeSystem, "", metadata)
	}

	return nil
}

// ListListings lists all listings for a tenant
func (s *ListingService) ListListings(ctx context.Context, tenantID string, opts repositories.PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	listings, err := s.listingRepo.List(ctx, tenantID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list listings: %w", err)
	}

	return listings, nil
}

// ListListingsByProperty lists all listings for a property
func (s *ListingService) ListListingsByProperty(ctx context.Context, tenantID, propertyID string, opts repositories.PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	listings, err := s.listingRepo.ListByProperty(ctx, tenantID, propertyID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list listings by property: %w", err)
	}

	return listings, nil
}

// ListListingsByBroker lists all listings for a broker
func (s *ListingService) ListListingsByBroker(ctx context.Context, tenantID, brokerID string, opts repositories.PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if brokerID == "" {
		return nil, fmt.Errorf("broker_id is required")
	}

	listings, err := s.listingRepo.ListByBroker(ctx, tenantID, brokerID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list listings by broker: %w", err)
	}

	return listings, nil
}

// ListActiveListings lists all active listings
func (s *ListingService) ListActiveListings(ctx context.Context, tenantID string, opts repositories.PaginationOptions) ([]*models.Listing, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	listings, err := s.listingRepo.ListActive(ctx, tenantID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list active listings: %w", err)
	}

	return listings, nil
}

// ActivateListing activates a listing
func (s *ListingService) ActivateListing(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("listing ID is required")
	}

	// Validate listing exists
	listing, err := s.listingRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}

	updates := map[string]interface{}{
		"is_active": true,
	}

	if err := s.listingRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to activate listing: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "listing_activated", models.ActorTypeSystem, "", map[string]interface{}{
		"listing_id":  id,
		"property_id": listing.PropertyID,
	})

	return nil
}

// DeactivateListing deactivates a listing
func (s *ListingService) DeactivateListing(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("listing ID is required")
	}

	// Validate listing exists
	listing, err := s.listingRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}

	updates := map[string]interface{}{
		"is_active": false,
	}

	if err := s.listingRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to deactivate listing: %w", err)
	}

	// If this was canonical and being deactivated, promote another active listing
	if listing.IsCanonical {
		otherListings, err := s.listingRepo.ListByProperty(ctx, tenantID, listing.PropertyID, repositories.PaginationOptions{Limit: 10})
		if err != nil {
			return fmt.Errorf("failed to find other listings: %w", err)
		}

		// Find another active listing to promote
		var newCanonicalID string
		for _, otherListing := range otherListings {
			if otherListing.ID != id && otherListing.IsActive {
				newCanonicalID = otherListing.ID
				break
			}
		}

		if newCanonicalID != "" {
			if err := s.SetCanonical(ctx, tenantID, newCanonicalID); err != nil {
				return fmt.Errorf("failed to promote new canonical listing: %w", err)
			}
		}
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "listing_deactivated", models.ActorTypeSystem, "", map[string]interface{}{
		"listing_id":  id,
		"property_id": listing.PropertyID,
	})

	return nil
}

// logActivity logs an activity (helper method)
func (s *ListingService) logActivity(ctx context.Context, tenantID, eventType string, actorType models.ActorType, actorID string, metadata map[string]interface{}) error {
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
