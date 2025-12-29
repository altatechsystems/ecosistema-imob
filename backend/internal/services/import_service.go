package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/altatech/ecosistema-imob/backend/internal/adapters/union"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/google/uuid"
)

// ImportService orchestrates the complete import process
type ImportService struct {
	db                   *firestore.Client
	deduplicationService *DeduplicationService
	photoProcessor       *PhotoProcessor // Optional - nil if GCS not configured
}

// NewImportService creates a new import service
func NewImportService(db *firestore.Client) *ImportService {
	return &ImportService{
		db:                   db,
		deduplicationService: NewDeduplicationService(db),
		photoProcessor:       nil, // Will be set via SetPhotoProcessor if needed
	}
}

// NewImportServiceWithPhotos creates import service with photo processing
func NewImportServiceWithPhotos(db *firestore.Client, photoProcessor *PhotoProcessor) *ImportService {
	return &ImportService{
		db:                   db,
		deduplicationService: NewDeduplicationService(db),
		photoProcessor:       photoProcessor,
	}
}

// SetPhotoProcessor sets the photo processor (optional)
func (s *ImportService) SetPhotoProcessor(photoProcessor *PhotoProcessor) {
	s.photoProcessor = photoProcessor
}

// ImportBatchRequest represents an import request
type ImportBatchRequest struct {
	TenantID   string
	XMLPath    string
	XLSPath    string // optional
	Source     string // "union"
	CreatedBy  string // broker_id or "system"
}

// ImportProperty imports a single property with all related entities
func (s *ImportService) ImportProperty(ctx context.Context, batch *models.ImportBatch, payload union.PropertyPayload) error {
	// 1. Check for duplicates
	dedupResult, err := s.deduplicationService.CheckDuplicate(ctx, &payload.Property)
	if err != nil {
		return fmt.Errorf("deduplication failed: %w", err)
	}

	if dedupResult.IsDuplicate {
		// Property already exists - update photos if any
		batch.TotalPropertiesMatchedExisting++
		log.Printf("Property %s already exists (matched by %s)", payload.Property.Reference, dedupResult.MatchType)

		// Log activity
		s.logActivity(ctx, batch.TenantID, "property_matched_existing", map[string]interface{}{
			"property_id":          dedupResult.ExistingProperty.ID,
			"reference":            payload.Property.Reference,
			"match_type":           dedupResult.MatchType,
			"batch_id":             batch.ID,
		})

		// Check if canonical listing exists, create if not
		existingPropertyID := dedupResult.ExistingProperty.ID
		log.Printf("🔍 Checking canonical listing for property %s (ref: %s)", existingPropertyID, payload.Property.Reference)
		listingID, err := s.findCanonicalListing(ctx, existingPropertyID)
		log.Printf("🔙 findCanonicalListing returned: listingID='%s', err=%v", listingID, err)

		if err != nil {
			log.Printf("⚠️  Error finding canonical listing for property %s: %v", existingPropertyID, err)
		} else if listingID == "" {
			log.Printf("❌ NO CANONICAL LISTING FOUND - will create one for property %s", payload.Property.Reference)
			// No canonical listing exists - create one
			log.Printf("🆕 Creating canonical listing for existing property %s", payload.Property.Reference)

			// Use the existing property from database
			property := dedupResult.ExistingProperty
			newListingID, err := s.createListing(ctx, batch.TenantID, property, payload.Photos, payload.Title, payload.Description)
			if err != nil {
				log.Printf("❌ Failed to create listing for existing property %s: %v", payload.Property.Reference, err)
			} else {
				listingID = newListingID
				batch.TotalListingsCreated++
				log.Printf("✅ Created listing %s for existing property %s", listingID, payload.Property.Reference)

				// Update property with canonical_listing_id
				_, err := s.db.Collection("properties").Doc(existingPropertyID).Update(ctx, []firestore.Update{
					{Path: "canonical_listing_id", Value: listingID},
					{Path: "updated_at", Value: time.Now()},
				})
				if err != nil {
					log.Printf("⚠️  Failed to update property with canonical_listing_id: %v", err)
				}
			}
		}

		// Process photos for existing property if any
		if len(payload.Photos) > 0 && listingID != "" {
			log.Printf("📸 Updating %d photos for existing property %s (ID: %s, Listing: %s)", len(payload.Photos), payload.Property.Reference, existingPropertyID, listingID)

			if s.photoProcessor != nil {
				// Process photos asynchronously
				go s.processPhotosAsync(ctx, batch, listingID, payload)
			} else {
				log.Printf("⚠️  Photo processor not configured - skipping photo update for existing property %s", payload.Property.Reference)
			}
		}

		return nil // Skip creating duplicate, but listing and photos are being processed
	}

	if dedupResult.PossibleDuplicate {
		// Mark as possible duplicate
		payload.Property.PossibleDuplicate = true
		batch.TotalPossibleDuplicates++

		log.Printf("Property %s is a possible duplicate (fingerprint match)", payload.Property.Reference)
	}

	// 2. Create Owner
	ownerID, enrichedFromXLS, err := s.createOwner(ctx, batch.TenantID, payload.Owner, payload.Property.Reference)
	if err != nil {
		return fmt.Errorf("failed to create owner: %w", err)
	}

	payload.Property.OwnerID = ownerID

	if enrichedFromXLS {
		batch.TotalOwnersEnrichedFromXLS++
		s.logActivity(ctx, batch.TenantID, "owner_enriched_from_xls", map[string]interface{}{
			"owner_id":  ownerID,
			"batch_id":  batch.ID,
		})
	} else if payload.Owner.OwnerStatus == models.OwnerStatusIncomplete {
		batch.TotalOwnersPlaceholders++
		s.logActivity(ctx, batch.TenantID, "owner_placeholder_created", map[string]interface{}{
			"owner_id":  ownerID,
			"batch_id":  batch.ID,
		})
	}

	// 3. Create Property
	if err := s.createProperty(ctx, &payload.Property); err != nil {
		return fmt.Errorf("failed to create property: %w", err)
	}

	batch.TotalPropertiesCreated++

	s.logActivity(ctx, batch.TenantID, "property_created", map[string]interface{}{
		"property_id": payload.Property.ID,
		"reference":   payload.Property.Reference,
		"batch_id":    batch.ID,
	})

	// 4. Create Listing
	listingID, err := s.createListing(ctx, batch.TenantID, &payload.Property, payload.Photos, payload.Title, payload.Description)
	if err != nil {
		return fmt.Errorf("failed to create listing: %w", err)
	}

	batch.TotalListingsCreated++

	s.logActivity(ctx, batch.TenantID, "listing_created", map[string]interface{}{
		"listing_id":  listingID,
		"property_id": payload.Property.ID,
		"batch_id":    batch.ID,
	})

	// 5. Set canonical listing (first listing becomes canonical)
	if err := s.setCanonicalListing(ctx, payload.Property.ID, listingID); err != nil {
		log.Printf("Warning: failed to set canonical listing: %v", err)
	} else {
		s.logActivity(ctx, batch.TenantID, "canonical_listing_assigned", map[string]interface{}{
			"property_id": payload.Property.ID,
			"listing_id":  listingID,
			"batch_id":    batch.ID,
		})
	}

	// 6. Create PropertyBrokerRole (originating_broker)
	if batch.CreatedBy != "" && batch.CreatedBy != "system" {
		if err := s.createPropertyBrokerRole(ctx, batch.TenantID, payload.Property.ID, batch.CreatedBy); err != nil {
			log.Printf("Warning: failed to create broker role: %v", err)
		}
	}

	// 7. Process photos (if photo processor is configured)
	if len(payload.Photos) > 0 {
		if s.photoProcessor != nil {
			// Process photos asynchronously (don't block import)
			go s.processPhotosAsync(ctx, batch, listingID, payload)
		} else {
			// No photo processor - photos stay as original URLs
			batch.TotalPhotosProcessed += len(payload.Photos)
			log.Printf("ℹ️  Photo processor not configured - skipping photo processing for property %s", payload.Property.Reference)
		}
	}

	return nil
}

// processPhotosAsync processes photos in background and updates listing
// Uses a worker pool to limit concurrent photo processing
func (s *ImportService) processPhotosAsync(ctx context.Context, batch *models.ImportBatch, listingID string, payload union.PropertyPayload) {
	const maxConcurrentPhotos = 5 // Limit concurrent photo downloads/processing

	semaphore := make(chan struct{}, maxConcurrentPhotos)
	processedPhotos := make([]models.Photo, 0, len(payload.Photos))
	photosMutex := &sync.Mutex{}
	errorCount := 0

	var wg sync.WaitGroup

	for i, photoURL := range payload.Photos {
		if photoURL == "" {
			continue
		}

		wg.Add(1)
		go func(url string, order int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Process single photo
			photo, err := s.photoProcessor.ProcessPhoto(ctx, batch.TenantID, payload.Property.ID, url, order)
			if err != nil {
				log.Printf("❌ Photo processing error for property %s, photo %d: %v", payload.Property.Reference, order, err)
				photosMutex.Lock()
				errorCount++
				photosMutex.Unlock()
				return
			}

			// Add to results
			photosMutex.Lock()
			processedPhotos = append(processedPhotos, photo)
			photosMutex.Unlock()
		}(photoURL, i)
	}

	// Wait for all photos to complete
	wg.Wait()

	// Update batch stats
	batch.TotalPhotosProcessed += len(processedPhotos)
	batch.TotalErrors += errorCount

	// Update listing with processed photos
	if len(processedPhotos) > 0 {
		if err := s.updateListingPhotos(ctx, listingID, processedPhotos); err != nil {
			log.Printf("❌ Failed to update listing photos for %s: %v", listingID, err)
		} else {
			log.Printf("✅ Updated listing %s with %d processed photos", listingID, len(processedPhotos))
		}
	}
}

// updateListingPhotos updates the photos in a listing
func (s *ImportService) updateListingPhotos(ctx context.Context, listingID string, photos []models.Photo) error {
	_, err := s.db.Collection("listings").Doc(listingID).Update(ctx, []firestore.Update{
		{Path: "photos", Value: photos},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

// findCanonicalListing finds the canonical listing ID for a property
func (s *ImportService) findCanonicalListing(ctx context.Context, propertyID string) (string, error) {
	log.Printf("  📋 findCanonicalListing: Looking up property %s", propertyID)

	// Get the property to find its canonical_listing_id
	propertyDoc, err := s.db.Collection("properties").Doc(propertyID).Get(ctx)
	if err != nil {
		log.Printf("  ❌ findCanonicalListing: Failed to get property %s: %v", propertyID, err)
		return "", fmt.Errorf("failed to get property: %w", err)
	}

	var property models.Property
	if err := propertyDoc.DataTo(&property); err != nil {
		log.Printf("  ❌ findCanonicalListing: Failed to parse property %s: %v", propertyID, err)
		return "", fmt.Errorf("failed to parse property: %w", err)
	}

	log.Printf("  📄 findCanonicalListing: Property %s has canonical_listing_id='%s'", propertyID, property.CanonicalListingID)

	if property.CanonicalListingID == "" {
		log.Printf("  ⚠️  findCanonicalListing: No canonical_listing_id set for property %s", propertyID)
		return "", nil // No canonical listing set
	}

	// Verify if the listing actually exists in Firestore
	log.Printf("  🔎 findCanonicalListing: Checking if listing %s exists in Firestore...", property.CanonicalListingID)
	listingDoc, err := s.db.Collection("listings").Doc(property.CanonicalListingID).Get(ctx)
	if err != nil {
		log.Printf("  ❌ findCanonicalListing: Error getting listing %s: %v", property.CanonicalListingID, err)
		log.Printf("  ⚠️  Listing %s referenced by property %s does not exist (error)", property.CanonicalListingID, propertyID)
		return "", nil // Listing doesn't exist, return empty string to trigger creation
	}

	if !listingDoc.Exists() {
		log.Printf("  ❌ findCanonicalListing: Listing %s does NOT exist in Firestore", property.CanonicalListingID)
		log.Printf("  ⚠️  Listing %s referenced by property %s does not exist (Exists()=false)", property.CanonicalListingID, propertyID)
		return "", nil // Listing doesn't exist, return empty string to trigger creation
	}

	log.Printf("  ✅ findCanonicalListing: Listing %s exists!", property.CanonicalListingID)
	return property.CanonicalListingID, nil
}

// createOwner creates or finds existing owner
func (s *ImportService) createOwner(ctx context.Context, tenantID string, ownerPayload union.OwnerPayload, reference string) (string, bool, error) {
	now := time.Now()
	ownerID := uuid.New().String()

	owner := models.Owner{
		ID:       ownerID,
		TenantID: tenantID,
		Name:     ownerPayload.Name,
		Email:    ownerPayload.Email,
		Phone:    ownerPayload.Phone,

		// Owner status tracking (CRITICAL FOR PROMPT 02)
		OwnerStatus: ownerPayload.OwnerStatus,

		// LGPD/Consent (passive owner)
		ConsentGiven:  false,
		ConsentOrigin: "import",

		// Metadata
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.Collection("owners").Doc(ownerID).Set(ctx, owner)
	if err != nil {
		return "", false, fmt.Errorf("failed to save owner: %w", err)
	}

	return ownerID, ownerPayload.EnrichedFromXLS, nil
}

// createProperty saves property to Firestore
func (s *ImportService) createProperty(ctx context.Context, property *models.Property) error {
	_, err := s.db.Collection("properties").Doc(property.ID).Set(ctx, property)
	return err
}

// createListing creates a listing for the property
func (s *ImportService) createListing(ctx context.Context, tenantID string, property *models.Property, photoURLs []string, title string, description string) (string, error) {
	now := time.Now()
	listingID := uuid.New().String()

	// Convert photo URLs to Photo objects
	photos := make([]models.Photo, 0, len(photoURLs))
	for i, url := range photoURLs {
		if url == "" {
			continue
		}

		photo := models.Photo{
			ID:        uuid.New().String(),
			URL:       url, // Original URL from XML - TODO: Process to WebP and upload to GCS
			ThumbURL:  url, // TODO: Generate thumbnail
			MediumURL: url, // TODO: Generate medium size
			LargeURL:  url, // TODO: Keep original or generate large
			Order:     i,
			IsCover:   i == 0,
		}
		photos = append(photos, photo)
	}

	// For imports, create a system listing with content from XML
	// The broker can edit this later to create their own listing
	listing := models.Listing{
		ID:         listingID,
		TenantID:   tenantID,
		PropertyID: property.ID,
		BrokerID:   "system", // System-generated listing from import

		// Content from XML (title and description)
		Title:       title,
		Description: description,

		// Photos
		Photos: photos,
		Videos: []models.Video{}, // Empty for imports

		// Status
		IsActive:    true,
		IsCanonical: false, // Will be set by setCanonicalListing

		// Metadata
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.Collection("listings").Doc(listingID).Set(ctx, listing)
	if err != nil {
		return "", err
	}

	return listingID, nil
}

// setCanonicalListing sets the canonical listing for a property
func (s *ImportService) setCanonicalListing(ctx context.Context, propertyID, listingID string) error {
	_, err := s.db.Collection("properties").Doc(propertyID).Update(ctx, []firestore.Update{
		{Path: "canonical_listing_id", Value: listingID},
	})
	return err
}

// createPropertyBrokerRole creates originating broker role
func (s *ImportService) createPropertyBrokerRole(ctx context.Context, tenantID, propertyID, brokerID string) error {
	now := time.Now()
	roleID := uuid.New().String()

	role := models.PropertyBrokerRole{
		ID:                   roleID,
		TenantID:             tenantID,
		PropertyID:           propertyID,
		BrokerID:             brokerID,
		Role:                 models.BrokerPropertyRoleOriginating, // CAPTADOR
		IsPrimary:            true,
		CommissionPercentage: 0, // To be defined later

		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.Collection("property_broker_roles").Doc(roleID).Set(ctx, role)
	return err
}

// logActivity logs an activity event
func (s *ImportService) logActivity(ctx context.Context, tenantID, eventType string, details map[string]interface{}) {
	// TODO: Implement activity logging
	// For now, just log to console
	log.Printf("[ACTIVITY] %s: %v", eventType, details)
}

// CreateBatch creates a new import batch
func (s *ImportService) CreateBatch(ctx context.Context, tenantID, source, createdBy string) (*models.ImportBatch, error) {
	now := time.Now()
	batchID := uuid.New().String()

	batch := &models.ImportBatch{
		ID:        batchID,
		TenantID:  tenantID,
		Source:    source,
		Status:    "processing",
		StartedAt: now,
		CreatedBy: createdBy,
	}

	_, err := s.db.Collection("import_batches").Doc(batchID).Set(ctx, batch)
	if err != nil {
		return nil, err
	}

	s.logActivity(ctx, tenantID, "import_batch_started", map[string]interface{}{
		"batch_id": batchID,
		"source":   source,
	})

	return batch, nil
}

// CompleteBatch marks batch as completed
func (s *ImportService) CompleteBatch(ctx context.Context, batch *models.ImportBatch) error {
	now := time.Now()
	batch.CompletedAt = &now
	batch.Status = "completed"

	_, err := s.db.Collection("import_batches").Doc(batch.ID).Set(ctx, batch)
	if err != nil {
		return err
	}

	s.logActivity(ctx, batch.TenantID, "import_batch_completed", map[string]interface{}{
		"batch_id":                         batch.ID,
		"total_properties_created":         batch.TotalPropertiesCreated,
		"total_properties_matched_existing": batch.TotalPropertiesMatchedExisting,
		"total_possible_duplicates":        batch.TotalPossibleDuplicates,
		"total_errors":                     batch.TotalErrors,
	})

	return nil
}

// LogError logs an import error
func (s *ImportService) LogError(ctx context.Context, batch *models.ImportBatch, errorType, errorMessage string, recordData map[string]interface{}) error {
	now := time.Now()
	errorID := uuid.New().String()

	importError := models.ImportError{
		ID:           errorID,
		BatchID:      batch.ID,
		TenantID:     batch.TenantID,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		RecordData:   recordData,
		Timestamp:    now,
	}

	_, err := s.db.Collection("import_errors").Doc(errorID).Set(ctx, importError)
	if err != nil {
		return err
	}

	batch.TotalErrors++
	return nil
}

// GetBatch retrieves a batch by ID
func (s *ImportService) GetBatch(ctx context.Context, batchID string) (*models.ImportBatch, error) {
	doc, err := s.db.Collection("import_batches").Doc(batchID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var batch models.ImportBatch
	if err := doc.DataTo(&batch); err != nil {
		return nil, err
	}

	batch.ID = doc.Ref.ID
	return &batch, nil
}
