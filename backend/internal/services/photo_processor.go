package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/nfnt/resize"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/storage"
)

// PhotoProcessor handles photo downloading, processing, and uploading
type PhotoProcessor struct {
	gcsClient *storage.GCSClient
}

// NewPhotoProcessor creates a new photo processor
func NewPhotoProcessor(gcsClient *storage.GCSClient) *PhotoProcessor {
	return &PhotoProcessor{
		gcsClient: gcsClient,
	}
}

// PhotoSize represents a photo size configuration
type PhotoSize struct {
	Name   string
	Width  uint
	Height uint
}

var (
	// Photo sizes according to Prompt 02
	SizeThumb  = PhotoSize{Name: "thumb", Width: 400, Height: 300}
	SizeMedium = PhotoSize{Name: "medium", Width: 800, Height: 600}
	SizeLarge  = PhotoSize{Name: "large", Width: 1600, Height: 1200}
)

// ProcessPhotoResult contains the result of photo processing
type ProcessPhotoResult struct {
	Photo models.Photo
	Error error
}

// ProcessPhotosAsync processes photos asynchronously
// Returns a channel that will receive results as photos are processed
func (p *PhotoProcessor) ProcessPhotosAsync(ctx context.Context, tenantID, propertyID string, photoURLs []string) <-chan ProcessPhotoResult {
	results := make(chan ProcessPhotoResult, len(photoURLs))

	go func() {
		defer close(results)

		for i, url := range photoURLs {
			if url == "" {
				continue
			}

			photo, err := p.ProcessPhoto(ctx, tenantID, propertyID, url, i)
			results <- ProcessPhotoResult{
				Photo: photo,
				Error: err,
			}
		}
	}()

	return results
}

// ProcessPhoto downloads, converts to WebP, and uploads a single photo
func (p *PhotoProcessor) ProcessPhoto(ctx context.Context, tenantID, propertyID, sourceURL string, order int) (models.Photo, error) {
	photoID := uuid.New().String()

	// 1. Download original image
	log.Printf("📥 Downloading photo %d from %s", order, sourceURL)
	imgData, contentType, err := p.downloadImage(ctx, sourceURL)
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to download image: %w", err)
	}

	// 2. Decode image
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to decode image (format: %s, type: %s): %w", format, contentType, err)
	}

	// 3. Generate JPEG versions in different sizes with optimized quality
	thumbJPEG, err := p.convertToJPEG(img, SizeThumb)
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to convert thumb to JPEG: %w", err)
	}

	mediumJPEG, err := p.convertToJPEG(img, SizeMedium)
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to convert medium to JPEG: %w", err)
	}

	largeJPEG, err := p.convertToJPEG(img, SizeLarge)
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to convert large to JPEG: %w", err)
	}

	// 4. Upload to GCS
	basePath := fmt.Sprintf("tenants/%s/properties/%s/photos", tenantID, propertyID)

	thumbURL, err := p.gcsClient.Upload(ctx, fmt.Sprintf("%s/%s_thumb.jpg", basePath, photoID), thumbJPEG, "image/jpeg")
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to upload thumb: %w", err)
	}

	mediumURL, err := p.gcsClient.Upload(ctx, fmt.Sprintf("%s/%s_medium.jpg", basePath, photoID), mediumJPEG, "image/jpeg")
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to upload medium: %w", err)
	}

	largeURL, err := p.gcsClient.Upload(ctx, fmt.Sprintf("%s/%s_large.jpg", basePath, photoID), largeJPEG, "image/jpeg")
	if err != nil {
		return models.Photo{}, fmt.Errorf("failed to upload large: %w", err)
	}

	log.Printf("✅ Processed photo %d: thumb=%s, medium=%s, large=%s", order, thumbURL, mediumURL, largeURL)

	// 5. Create Photo model
	photo := models.Photo{
		ID:        photoID,
		URL:       largeURL,   // Main URL is the large version
		ThumbURL:  thumbURL,   // 400x300
		MediumURL: mediumURL,  // 800x600
		LargeURL:  largeURL,   // 1600x1200
		Order:     order,
		IsCover:   order == 0, // First photo is cover
	}

	return photo, nil
}

// downloadImage downloads an image from a URL
func (p *PhotoProcessor) downloadImage(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return data, resp.Header.Get("Content-Type"), nil
}

// convertToJPEG converts an image to JPEG format with specified size
func (p *PhotoProcessor) convertToJPEG(img image.Image, size PhotoSize) ([]byte, error) {
	// Resize image maintaining aspect ratio
	resized := resize.Thumbnail(size.Width, size.Height, img, resize.Lanczos3)

	// Encode to JPEG with high quality
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ProcessPhotosBatch processes multiple photos and returns all results
func (p *PhotoProcessor) ProcessPhotosBatch(ctx context.Context, tenantID, propertyID string, photoURLs []string) ([]models.Photo, []error) {
	photos := make([]models.Photo, 0, len(photoURLs))
	errors := make([]error, 0)

	for i, url := range photoURLs {
		if url == "" {
			continue
		}

		photo, err := p.ProcessPhoto(ctx, tenantID, propertyID, url, i)
		if err != nil {
			log.Printf("❌ Failed to process photo %d (%s): %v", i, url, err)
			errors = append(errors, fmt.Errorf("photo %d: %w", i, err))
			continue
		}

		photos = append(photos, photo)
	}

	return photos, errors
}

// GetPhotoBasename extracts basename from URL for GCS path
func GetPhotoBasename(url string) string {
	base := filepath.Base(url)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
