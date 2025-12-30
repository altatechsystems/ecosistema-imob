package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
)

// PropertyService handles business logic for property management
type PropertyService struct {
	propertyRepo    *repositories.PropertyRepository
	listingRepo     *repositories.ListingRepository
	ownerRepo       *repositories.OwnerRepository
	brokerRepo      *repositories.BrokerRepository
	tenantRepo      *repositories.TenantRepository
	activityLogRepo *repositories.ActivityLogRepository
}

// NewPropertyService creates a new property service
func NewPropertyService(
	propertyRepo *repositories.PropertyRepository,
	listingRepo *repositories.ListingRepository,
	ownerRepo *repositories.OwnerRepository,
	brokerRepo *repositories.BrokerRepository,
	tenantRepo *repositories.TenantRepository,
	activityLogRepo *repositories.ActivityLogRepository,
) *PropertyService {
	return &PropertyService{
		propertyRepo:    propertyRepo,
		listingRepo:     listingRepo,
		ownerRepo:       ownerRepo,
		brokerRepo:      brokerRepo,
		tenantRepo:      tenantRepo,
		activityLogRepo: activityLogRepo,
	}
}

// CreateProperty creates a new property with validation and deduplication
func (s *PropertyService) CreateProperty(ctx context.Context, property *models.Property) error {
	// Validate required fields
	if property.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if property.OwnerID == "" {
		return fmt.Errorf("owner_id is required")
	}
	if property.Neighborhood == "" {
		return fmt.Errorf("neighborhood is required")
	}
	if property.City == "" {
		return fmt.Errorf("city is required")
	}
	if property.State == "" {
		return fmt.Errorf("state is required")
	}

	// Validate tenant exists
	if _, err := s.tenantRepo.Get(ctx, property.TenantID); err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Validate owner exists
	if _, err := s.ownerRepo.Get(ctx, property.TenantID, property.OwnerID); err != nil {
		return fmt.Errorf("owner not found: %w", err)
	}

	// Validate property type
	if err := s.validatePropertyType(property.PropertyType); err != nil {
		return err
	}

	// Set defaults
	if property.Country == "" {
		property.Country = "BR"
	}
	if property.PriceCurrency == "" {
		property.PriceCurrency = "BRL"
	}
	if property.Status == "" {
		property.Status = models.PropertyStatusAvailable
	}
	if property.Visibility == "" {
		property.Visibility = models.PropertyVisibilityPrivate
	}

	// Generate slug if not provided
	if property.Slug == "" {
		property.Slug = s.GenerateSlug(property)
	} else {
		property.Slug = s.NormalizeSlug(property.Slug)
	}

	// Generate fingerprint for deduplication
	property.Fingerprint = s.GenerateFingerprint(property)

	// Check for duplicates
	duplicates, err := s.CheckDuplicates(ctx, property)
	if err != nil {
		return fmt.Errorf("failed to check duplicates: %w", err)
	}

	// Mark as possible duplicate if found
	if len(duplicates) > 0 {
		property.PossibleDuplicate = true
	} else {
		property.PossibleDuplicate = false
	}

	// Determine data completeness
	property.DataCompleteness = s.determineDataCompleteness(property)

	// Create property in repository
	if err := s.propertyRepo.Create(ctx, property); err != nil {
		return fmt.Errorf("failed to create property: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, property.TenantID, "property_created", models.ActorTypeSystem, "", map[string]interface{}{
		"property_id":        property.ID,
		"slug":               property.Slug,
		"fingerprint":        property.Fingerprint,
		"possible_duplicate": property.PossibleDuplicate,
		"external_source":    property.ExternalSource,
		"external_id":        property.ExternalID,
	})

	return nil
}

// GetProperty retrieves a property by ID
func (s *PropertyService) GetProperty(ctx context.Context, tenantID, id string) (*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return nil, fmt.Errorf("property ID is required")
	}

	property, err := s.propertyRepo.Get(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	// Populate photos from canonical listing
	s.populatePropertyPhotos(ctx, tenantID, property)

	return property, nil
}

// GetPropertyBySlug retrieves a property by slug
func (s *PropertyService) GetPropertyBySlug(ctx context.Context, tenantID, slug string) (*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	property, err := s.propertyRepo.GetBySlug(ctx, tenantID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get property by slug: %w", err)
	}

	// Populate photos from canonical listing
	s.populatePropertyPhotos(ctx, tenantID, property)

	// Populate broker data
	s.populatePropertyBroker(ctx, tenantID, property)

	return property, nil
}

// populatePropertyPhotos populates cover_image_url and images from canonical listing
func (s *PropertyService) populatePropertyPhotos(ctx context.Context, tenantID string, property *models.Property) {
	if property == nil || property.CanonicalListingID == "" {
		return
	}

	// Get canonical listing to fetch photos
	listing, err := s.listingRepo.Get(ctx, tenantID, property.CanonicalListingID)
	if err != nil || listing == nil || len(listing.Photos) == 0 {
		return
	}

	// Sort photos: cover first, then by order
	photos := make([]models.Photo, len(listing.Photos))
	copy(photos, listing.Photos)
	sort.SliceStable(photos, func(i, j int) bool {
		// Cover photo always comes first
		if photos[i].IsCover != photos[j].IsCover {
			return photos[i].IsCover
		}
		// Then sort by Order field
		return photos[i].Order < photos[j].Order
	})

	// Set cover image URL (first photo after sorting should be cover)
	property.CoverImageURL = photos[0].ThumbURL

	// Populate images array for detail page (sorted photos)
	property.Images = photos
}

// populatePropertyBroker populates broker data (captador) for public display
func (s *PropertyService) populatePropertyBroker(ctx context.Context, tenantID string, property *models.Property) {
	if property == nil || property.CaptadorID == "" {
		return
	}

	// Get broker data
	broker, err := s.brokerRepo.Get(ctx, tenantID, property.CaptadorID)
	if err != nil || broker == nil {
		return
	}

	// Create a sanitized broker object for public display (omit sensitive data)
	property.Captador = &models.BrokerPublic{
		ID:           broker.ID,
		Name:         broker.Name,
		Email:        broker.Email,
		Phone:        broker.Phone,
		CRECI:        broker.CRECI,
		PhotoURL:     broker.PhotoURL,
		Bio:          broker.Bio,
		Specialties:  broker.Specialties,
		Languages:    broker.Languages,
		Experience:   broker.Experience,
		Company:      broker.Company,
		Website:      broker.Website,
		TotalSales:   broker.TotalSales,
		TotalListings: broker.TotalListings,
		AveragePrice: broker.AveragePrice,
		Rating:       broker.Rating,
		ReviewCount:  broker.ReviewCount,
	}
}

// UpdateProperty updates a property with validation
func (s *PropertyService) UpdateProperty(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("property ID is required")
	}

	// Validate property exists
	existing, err := s.propertyRepo.Get(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("property not found: %w", err)
	}

	// Validate property type if being updated
	if propertyType, ok := updates["property_type"].(models.PropertyType); ok {
		if err := s.validatePropertyType(propertyType); err != nil {
			return err
		}
	}

	// Validate status if being updated
	if status, ok := updates["status"].(models.PropertyStatus); ok {
		if err := s.validatePropertyStatus(status); err != nil {
			return err
		}
	}

	// Validate visibility if being updated
	if visibility, ok := updates["visibility"].(models.PropertyVisibility); ok {
		if err := s.validatePropertyVisibility(visibility); err != nil {
			return err
		}
	}

	// Normalize slug if being updated
	if slug, ok := updates["slug"].(string); ok && slug != "" {
		updates["slug"] = s.NormalizeSlug(slug)
	}

	// Regenerate fingerprint if key fields are updated
	shouldRegenerateFingerprint := false
	keyFields := []string{"street", "number", "city", "property_type", "total_area"}
	for _, field := range keyFields {
		if _, ok := updates[field]; ok {
			shouldRegenerateFingerprint = true
			break
		}
	}

	if shouldRegenerateFingerprint {
		// Create a copy of the property with updates applied
		updatedProperty := *existing
		for key, value := range updates {
			switch key {
			case "street":
				updatedProperty.Street = value.(string)
			case "number":
				updatedProperty.Number = value.(string)
			case "city":
				updatedProperty.City = value.(string)
			case "property_type":
				updatedProperty.PropertyType = value.(models.PropertyType)
			case "total_area":
				updatedProperty.TotalArea = value.(float64)
			}
		}
		updates["fingerprint"] = s.GenerateFingerprint(&updatedProperty)
	}

	// Prevent updating tenant_id
	delete(updates, "tenant_id")

	// Update property in repository
	if err := s.propertyRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update property: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "property_updated", models.ActorTypeSystem, "", map[string]interface{}{
		"property_id": id,
		"updates":     updates,
	})

	return nil
}

// DeleteProperty deletes a property
func (s *PropertyService) DeleteProperty(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("property ID is required")
	}

	// Validate property exists
	if _, err := s.propertyRepo.Get(ctx, tenantID, id); err != nil {
		return fmt.Errorf("property not found: %w", err)
	}

	// Delete property from repository
	if err := s.propertyRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete property: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "property_deleted", models.ActorTypeSystem, "", map[string]interface{}{
		"property_id": id,
	})

	return nil
}

// ListProperties lists properties with filters and pagination
func (s *PropertyService) ListProperties(ctx context.Context, tenantID string, filters *repositories.PropertyFilters, opts repositories.PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	properties, err := s.propertyRepo.List(ctx, tenantID, filters, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list properties: %w", err)
	}

	// Populate cover image URL from canonical listing
	for _, property := range properties {
		if property.CanonicalListingID != "" {
			// Get listing to fetch first photo
			listing, err := s.listingRepo.Get(ctx, tenantID, property.CanonicalListingID)
			if err != nil {
				// Log error but don't fail the whole request
				continue
			}
			if listing != nil && len(listing.Photos) > 0 {
				property.CoverImageURL = listing.Photos[0].ThumbURL
			}
		}
	}

	return properties, nil
}

// SearchProperties searches properties by location
func (s *PropertyService) SearchProperties(ctx context.Context, tenantID, city, neighborhood string, opts repositories.PaginationOptions) ([]*models.Property, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	properties, err := s.propertyRepo.SearchByLocation(ctx, tenantID, city, neighborhood, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search properties: %w", err)
	}

	return properties, nil
}

// UpdateStatus updates the status of a property
func (s *PropertyService) UpdateStatus(ctx context.Context, tenantID, id string, status models.PropertyStatus) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("property ID is required")
	}

	if err := s.validatePropertyStatus(status); err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":               status,
		"status_confirmed_at":  now,
	}

	if err := s.propertyRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update property status: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "property_status_changed", models.ActorTypeSystem, "", map[string]interface{}{
		"property_id": id,
		"status":      status,
	})

	return nil
}

// UpdateVisibility updates the visibility of a property
func (s *PropertyService) UpdateVisibility(ctx context.Context, tenantID, id string, visibility models.PropertyVisibility) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if id == "" {
		return fmt.Errorf("property ID is required")
	}

	if err := s.validatePropertyVisibility(visibility); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"visibility": visibility,
	}

	if err := s.propertyRepo.Update(ctx, tenantID, id, updates); err != nil {
		return fmt.Errorf("failed to update property visibility: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "property_visibility_changed", models.ActorTypeSystem, "", map[string]interface{}{
		"property_id": id,
		"visibility":  visibility,
	})

	return nil
}

// GenerateFingerprint generates a fingerprint for deduplication
// Fingerprint format: hash(street+number+city+property_type+area)
func (s *PropertyService) GenerateFingerprint(property *models.Property) string {
	// Normalize components
	street := strings.ToLower(strings.TrimSpace(property.Street))
	number := strings.ToLower(strings.TrimSpace(property.Number))
	city := strings.ToLower(strings.TrimSpace(property.City))
	propertyType := strings.ToLower(string(property.PropertyType))
	area := fmt.Sprintf("%.2f", property.TotalArea)

	// Combine components
	combined := fmt.Sprintf("%s|%s|%s|%s|%s", street, number, city, propertyType, area)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// CheckDuplicates checks for duplicate properties by fingerprint
func (s *PropertyService) CheckDuplicates(ctx context.Context, property *models.Property) ([]*models.Property, error) {
	if property.Fingerprint == "" {
		property.Fingerprint = s.GenerateFingerprint(property)
	}

	duplicates, err := s.propertyRepo.ListByFingerprint(ctx, property.TenantID, property.Fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicates: %w", err)
	}

	// Exclude the current property if it already exists
	result := make([]*models.Property, 0)
	for _, dup := range duplicates {
		if dup.ID != property.ID {
			result = append(result, dup)
		}
	}

	return result, nil
}

// LinkDuplicates links duplicate properties (for future merge/resolution)
func (s *PropertyService) LinkDuplicates(ctx context.Context, tenantID, primaryID string, duplicateIDs []string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if primaryID == "" {
		return fmt.Errorf("primary property ID is required")
	}

	// Validate primary property exists
	primary, err := s.propertyRepo.Get(ctx, tenantID, primaryID)
	if err != nil {
		return fmt.Errorf("primary property not found: %w", err)
	}

	// Mark duplicates
	for _, dupID := range duplicateIDs {
		if dupID == primaryID {
			continue // Skip the primary
		}

		updates := map[string]interface{}{
			"possible_duplicate": true,
		}

		if err := s.propertyRepo.Update(ctx, tenantID, dupID, updates); err != nil {
			return fmt.Errorf("failed to mark duplicate property %s: %w", dupID, err)
		}
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "property_duplicates_linked", models.ActorTypeSystem, "", map[string]interface{}{
		"primary_property_id": primaryID,
		"duplicate_ids":       duplicateIDs,
	})

	_ = primary // silence unused warning

	return nil
}

// GenerateSlug generates a URL-friendly slug from property data
func (s *PropertyService) GenerateSlug(property *models.Property) string {
	// Build slug from: property_type-city-neighborhood-street-number
	parts := []string{
		string(property.PropertyType),
		property.City,
		property.Neighborhood,
	}

	if property.Street != "" {
		parts = append(parts, property.Street)
	}
	if property.Number != "" {
		parts = append(parts, property.Number)
	}

	slug := strings.Join(parts, "-")
	return s.NormalizeSlug(slug)
}

// NormalizeSlug normalizes a slug
func (s *PropertyService) NormalizeSlug(slug string) string {
	// Convert to lowercase
	slug = strings.ToLower(slug)

	// Remove accents
	slug = s.removeAccents(slug)

	// Replace invalid characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Replace multiple consecutive hyphens with single hyphen
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
	}

	return slug
}

// removeAccents removes accents from a string
func (s *PropertyService) removeAccents(str string) string {
	replacements := map[rune]string{
		'á': "a", 'à': "a", 'ã': "a", 'â': "a",
		'é': "e", 'è': "e", 'ê': "e",
		'í': "i", 'ì': "i", 'î': "i",
		'ó': "o", 'ò': "o", 'õ': "o", 'ô': "o",
		'ú': "u", 'ù': "u", 'û': "u",
		'ç': "c",
		'ñ': "n",
	}

	result := ""
	for _, char := range str {
		if replacement, ok := replacements[char]; ok {
			result += replacement
		} else {
			result += string(char)
		}
	}

	return result
}

// determineDataCompleteness determines the data completeness of a property
func (s *PropertyService) determineDataCompleteness(property *models.Property) string {
	requiredFields := []bool{
		property.Street != "",
		property.Number != "",
		property.Neighborhood != "",
		property.City != "",
		property.State != "",
		property.PropertyType != "",
		property.PriceAmount > 0,
	}

	optionalFields := []bool{
		property.Bedrooms > 0,
		property.Bathrooms > 0,
		property.TotalArea > 0,
		property.ZipCode != "",
	}

	requiredCount := 0
	for _, filled := range requiredFields {
		if filled {
			requiredCount++
		}
	}

	optionalCount := 0
	for _, filled := range optionalFields {
		if filled {
			optionalCount++
		}
	}

	// Complete: all required + most optional
	if requiredCount == len(requiredFields) && optionalCount >= len(optionalFields)-1 {
		return "complete"
	}

	// Partial: most required fields
	if requiredCount >= len(requiredFields)-2 {
		return "partial"
	}

	// Incomplete: missing many fields
	return "incomplete"
}

// validatePropertyType validates property type
func (s *PropertyService) validatePropertyType(propertyType models.PropertyType) error {
	validTypes := map[models.PropertyType]bool{
		models.PropertyTypeApartment:       true,
		models.PropertyTypeHouse:           true,
		models.PropertyTypeLand:            true,
		models.PropertyTypeCommercial:      true,
		models.PropertyTypeNewDevelopment:  true,
		models.PropertyTypeCondoLot:        true,
		models.PropertyTypeBuildingLot:     true,
	}

	if !validTypes[propertyType] {
		return fmt.Errorf("invalid property type")
	}

	return nil
}

// validatePropertyStatus validates property status
func (s *PropertyService) validatePropertyStatus(status models.PropertyStatus) error {
	validStatuses := map[models.PropertyStatus]bool{
		models.PropertyStatusAvailable:           true,
		models.PropertyStatusUnavailable:         true,
		models.PropertyStatusPendingConfirmation: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid property status")
	}

	return nil
}

// validatePropertyVisibility validates property visibility
func (s *PropertyService) validatePropertyVisibility(visibility models.PropertyVisibility) error {
	validVisibilities := map[models.PropertyVisibility]bool{
		models.PropertyVisibilityPrivate:     true,
		models.PropertyVisibilityNetwork:     true,
		models.PropertyVisibilityMarketplace: true,
		models.PropertyVisibilityPublic:      true,
	}

	if !validVisibilities[visibility] {
		return fmt.Errorf("invalid property visibility")
	}

	return nil
}

// logActivity logs an activity (helper method)
func (s *PropertyService) logActivity(ctx context.Context, tenantID, eventType string, actorType models.ActorType, actorID string, metadata map[string]interface{}) error {
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
