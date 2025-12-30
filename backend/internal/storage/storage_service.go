package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
)

const (
	// MaxFileSize is the maximum file size in bytes (10MB)
	MaxFileSize = 10 * 1024 * 1024

	// SignedURLExpiration is the duration for signed URLs
	SignedURLExpiration = 1 * time.Hour
)

var (
	// AllowedContentTypes defines the allowed MIME types for images
	AllowedContentTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	// ErrFileTooLarge is returned when file size exceeds the limit
	ErrFileTooLarge = fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)

	// ErrInvalidFileType is returned when file type is not allowed
	ErrInvalidFileType = fmt.Errorf("invalid file type, allowed types: image/jpeg, image/png, image/webp")

	// ErrImageNotFound is returned when image is not found
	ErrImageNotFound = fmt.Errorf("image not found")
)

// ImageMetadata represents metadata for an uploaded image
type ImageMetadata struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	PropertyID       string    `json:"property_id"`
	OriginalFilename string    `json:"original_filename"`
	ContentType      string    `json:"content_type"`
	Size             int64     `json:"size"`
	URL              string    `json:"url"`
	UploadedAt       time.Time `json:"uploaded_at"`
}

// StorageService handles Firebase Storage operations for property images
type StorageService struct {
	storageClient   *storage.Client
	bucketName      string
	activityLogRepo *repositories.ActivityLogRepository
}

// NewStorageService creates a new storage service
func NewStorageService(
	ctx context.Context,
	bucketName string,
	activityLogRepo *repositories.ActivityLogRepository,
) (*StorageService, error) {
	// Initialize GCS client directly
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Storage client: %w", err)
	}

	return &StorageService{
		storageClient:   client,
		bucketName:      bucketName,
		activityLogRepo: activityLogRepo,
	}, nil
}

// UploadPropertyImage uploads an image to Firebase Storage
func (s *StorageService) UploadPropertyImage(
	ctx context.Context,
	tenantID, propertyID string,
	file multipart.File,
	header *multipart.FileHeader,
) (*ImageMetadata, error) {
	// Validate inputs
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	// Validate file size
	if header.Size > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !AllowedContentTypes[contentType] {
		return nil, ErrInvalidFileType
	}

	// Generate unique image ID
	imageID := uuid.New().String()

	// Build storage path: /properties/{tenantID}/{propertyID}/{imageID}
	storagePath := fmt.Sprintf("properties/%s/%s/%s", tenantID, propertyID, imageID)

	// Get bucket handle
	bucket := s.storageClient.Bucket(s.bucketName)

	// Create object handle
	obj := bucket.Object(storagePath)

	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	writer.Metadata = map[string]string{
		"tenant_id":         tenantID,
		"property_id":       propertyID,
		"original_filename": header.Filename,
		"uploaded_at":       time.Now().Format(time.RFC3339),
	}

	// Copy file data to storage
	if _, err := io.Copy(writer, file); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize upload: %w", err)
	}

	// Generate signed URL
	url, err := s.generateSignedURL(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %w", err)
	}

	// Create metadata
	metadata := &ImageMetadata{
		ID:               imageID,
		TenantID:         tenantID,
		PropertyID:       propertyID,
		OriginalFilename: header.Filename,
		ContentType:      contentType,
		Size:             header.Size,
		URL:              url,
		UploadedAt:       time.Now(),
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "image_uploaded", models.ActorTypeSystem, "", map[string]interface{}{
		"property_id":       propertyID,
		"image_id":          imageID,
		"original_filename": header.Filename,
		"content_type":      contentType,
		"size":              header.Size,
	})

	return metadata, nil
}

// GetPropertyImageURL retrieves a signed URL for an image
func (s *StorageService) GetPropertyImageURL(
	ctx context.Context,
	tenantID, propertyID, imageID string,
) (string, error) {
	// Validate inputs
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return "", fmt.Errorf("property_id is required")
	}
	if imageID == "" {
		return "", fmt.Errorf("image_id is required")
	}

	// Build storage path
	storagePath := fmt.Sprintf("properties/%s/%s/%s", tenantID, propertyID, imageID)

	// Check if object exists
	bucket := s.storageClient.Bucket(s.bucketName)
	obj := bucket.Object(storagePath)

	if _, err := obj.Attrs(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return "", ErrImageNotFound
		}
		return "", fmt.Errorf("failed to check image: %w", err)
	}

	// Generate signed URL
	url, err := s.generateSignedURL(ctx, storagePath)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// DeletePropertyImage deletes an image from Firebase Storage
func (s *StorageService) DeletePropertyImage(
	ctx context.Context,
	tenantID, propertyID, imageID string,
) error {
	// Validate inputs
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return fmt.Errorf("property_id is required")
	}
	if imageID == "" {
		return fmt.Errorf("image_id is required")
	}

	// Build storage path
	storagePath := fmt.Sprintf("properties/%s/%s/%s", tenantID, propertyID, imageID)

	// Get bucket handle
	bucket := s.storageClient.Bucket(s.bucketName)
	obj := bucket.Object(storagePath)

	// Delete the object
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return ErrImageNotFound
		}
		return fmt.Errorf("failed to delete image: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "image_deleted", models.ActorTypeSystem, "", map[string]interface{}{
		"property_id": propertyID,
		"image_id":    imageID,
	})

	return nil
}

// ListPropertyImages lists all images for a property
func (s *StorageService) ListPropertyImages(
	ctx context.Context,
	tenantID, propertyID string,
) ([]*ImageMetadata, error) {
	// Validate inputs
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property_id is required")
	}

	// Build prefix for listing
	prefix := fmt.Sprintf("properties/%s/%s/", tenantID, propertyID)

	// Get bucket handle
	bucket := s.storageClient.Bucket(s.bucketName)

	// List objects with prefix
	query := &storage.Query{
		Prefix: prefix,
	}

	it := bucket.Objects(ctx, query)

	images := []*ImageMetadata{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list images: %w", err)
		}

		// Extract image ID from path
		// Path format: properties/{tenantID}/{propertyID}/{imageID}
		parts := strings.Split(attrs.Name, "/")
		if len(parts) != 4 {
			continue
		}
		imageID := parts[3]

		// Generate signed URL
		url, err := s.generateSignedURL(ctx, attrs.Name)
		if err != nil {
			// Skip this image if we can't generate URL
			continue
		}

		// Get metadata
		originalFilename := attrs.Metadata["original_filename"]
		if originalFilename == "" {
			originalFilename = imageID
		}

		uploadedAt := attrs.Created
		if uploadedAtStr, ok := attrs.Metadata["uploaded_at"]; ok {
			if t, err := time.Parse(time.RFC3339, uploadedAtStr); err == nil {
				uploadedAt = t
			}
		}

		images = append(images, &ImageMetadata{
			ID:               imageID,
			TenantID:         tenantID,
			PropertyID:       propertyID,
			OriginalFilename: originalFilename,
			ContentType:      attrs.ContentType,
			Size:             attrs.Size,
			URL:              url,
			UploadedAt:       uploadedAt,
		})
	}

	return images, nil
}

// generateSignedURL generates a signed URL for accessing an object
func (s *StorageService) generateSignedURL(ctx context.Context, objectPath string) (string, error) {
	// For Firebase Storage, use the Firebase Storage download URL format
	// This works if the bucket has public access or Firebase Storage rules allow it

	// URL encode the object path for use in URL
	encodedPath := url.PathEscape(objectPath)

	// Firebase Storage public URL format
	// Format: https://firebasestorage.googleapis.com/v0/b/{bucket}/o/{encodedPath}?alt=media
	downloadURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media",
		s.bucketName,
		encodedPath)

	return downloadURL, nil

	// TODO: Implement proper signed URLs in production for private buckets
	// This requires:
	// 1. Service account with Storage Object Admin role
	// 2. storage.SignedURL() with GoogleAccessID and PrivateKey
	// Example:
	// opts := &storage.SignedURLOptions{
	//     GoogleAccessID: "service-account@project.iam.gserviceaccount.com",
	//     PrivateKey:     []byte("-----BEGIN PRIVATE KEY-----\n..."),
	//     Method:         "GET",
	//     Expires:        time.Now().Add(SignedURLExpiration),
	// }
	// return storage.SignedURL(s.bucketName, objectPath, opts)
}

// ValidateContentType validates if the content type is allowed
func (s *StorageService) ValidateContentType(contentType string) error {
	if !AllowedContentTypes[contentType] {
		return ErrInvalidFileType
	}
	return nil
}

// ValidateFileSize validates if the file size is within limits
func (s *StorageService) ValidateFileSize(size int64) error {
	if size > MaxFileSize {
		return ErrFileTooLarge
	}
	return nil
}

// UploadBrokerPhoto uploads a broker profile photo to Firebase Storage
func (s *StorageService) UploadBrokerPhoto(
	ctx context.Context,
	tenantID, brokerID string,
	file multipart.File,
	header *multipart.FileHeader,
) (string, error) {
	// Validate inputs
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	if brokerID == "" {
		return "", fmt.Errorf("broker_id is required")
	}

	// Validate file size (broker photos have 5MB limit as set in handler)
	if header.Size > MaxFileSize {
		return "", ErrFileTooLarge
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !AllowedContentTypes[contentType] {
		return "", ErrInvalidFileType
	}

	// Determine file extension
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	}

	// Build storage path: /brokers/{tenantID}/{brokerID}/photo{ext}
	storagePath := fmt.Sprintf("brokers/%s/%s/photo%s", tenantID, brokerID, ext)

	// Get bucket handle
	bucket := s.storageClient.Bucket(s.bucketName)

	// Create object handle
	obj := bucket.Object(storagePath)

	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	writer.Metadata = map[string]string{
		"tenant_id":         tenantID,
		"broker_id":         brokerID,
		"original_filename": header.Filename,
		"uploaded_at":       time.Now().Format(time.RFC3339),
	}

	// Copy file data to storage
	if _, err := io.Copy(writer, file); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to upload broker photo: %w", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize broker photo upload: %w", err)
	}

	// Generate signed URL
	url, err := s.generateSignedURL(ctx, storagePath)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL for broker photo: %w", err)
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_photo_uploaded", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id":         brokerID,
		"original_filename": header.Filename,
		"content_type":      contentType,
		"size":              header.Size,
	})

	return url, nil
}

// DeleteBrokerPhoto deletes a broker's profile photo from Firebase Storage
func (s *StorageService) DeleteBrokerPhoto(
	ctx context.Context,
	tenantID, brokerID string,
) error {
	// Validate inputs
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if brokerID == "" {
		return fmt.Errorf("broker_id is required")
	}

	// Build prefix for broker photos: /brokers/{tenantID}/{brokerID}/
	prefix := fmt.Sprintf("brokers/%s/%s/", tenantID, brokerID)

	// Get bucket handle
	bucket := s.storageClient.Bucket(s.bucketName)

	// List all objects with the prefix (should be only one photo, but handle multiple extensions)
	query := &storage.Query{
		Prefix: prefix,
	}

	it := bucket.Objects(ctx, query)

	// Delete all photos found (usually just one)
	deletedCount := 0
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list broker photos: %w", err)
		}

		// Delete the object
		obj := bucket.Object(attrs.Name)
		if err := obj.Delete(ctx); err != nil {
			if err == storage.ErrObjectNotExist {
				continue
			}
			return fmt.Errorf("failed to delete broker photo: %w", err)
		}
		deletedCount++
	}

	// If no photos were found or deleted, that's okay - not an error
	if deletedCount == 0 {
		return nil
	}

	// Log activity
	_ = s.logActivity(ctx, tenantID, "broker_photo_deleted", models.ActorTypeSystem, "", map[string]interface{}{
		"broker_id": brokerID,
	})

	return nil
}

// logActivity logs an activity
func (s *StorageService) logActivity(
	ctx context.Context,
	tenantID, eventType string,
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
