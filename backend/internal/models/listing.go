package models

import "time"

// Listing represents a broker's advertisement of a property (anúncio)
// Collection: /tenants/{tenantId}/listings/{listingId}
// IMPORTANTE: Listing = visão do corretor sobre o Property
type Listing struct {
	ID         string `firestore:"-" json:"id"`
	TenantID   string `firestore:"tenant_id" json:"tenant_id"`
	PropertyID string `firestore:"property_id" json:"property_id"` // ref Property
	BrokerID   string `firestore:"broker_id" json:"broker_id"`     // ref Broker (listing broker/vendedor)

	// Conteúdo do anúncio
	Title       string `firestore:"title" json:"title"`
	Description string `firestore:"description" json:"description"`

	// Fotos (URLs GCS)
	Photos []Photo `firestore:"photos" json:"photos"`

	// Vídeos (AI_DEV_DIRECTIVE Seção 23)
	Videos []Video `firestore:"videos" json:"videos"`

	// SEO
	MetaTitle       string `firestore:"meta_title,omitempty" json:"meta_title,omitempty"`
	MetaDescription string `firestore:"meta_description,omitempty" json:"meta_description,omitempty"`

	// Status
	IsActive    bool `firestore:"is_active" json:"is_active"`
	IsCanonical bool `firestore:"is_canonical" json:"is_canonical"` // denormalizado para query

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}

// Photo represents a property photo
type Photo struct {
	ID        string `firestore:"id" json:"id"`
	URL       string `firestore:"url" json:"url"`               // GCS URL
	ThumbURL  string `firestore:"thumb_url" json:"thumb_url"`   // 400x300 WebP
	MediumURL string `firestore:"medium_url" json:"medium_url"` // 800x600 WebP
	LargeURL  string `firestore:"large_url" json:"large_url"`   // 1600x1200 WebP
	Order     int    `firestore:"order" json:"order"`
	IsCover   bool   `firestore:"is_cover" json:"is_cover"`

	// Análise de qualidade (AI_DEV_DIRECTIVE Seção 23 - Fase 2)
	RoomType       string  `firestore:"room_type,omitempty" json:"room_type,omitempty"`             // living_room, kitchen, bedroom, bathroom, exterior
	Quality        float64 `firestore:"quality,omitempty" json:"quality,omitempty"`                 // 0.0 - 1.0
	SuggestedOrder int     `firestore:"suggested_order,omitempty" json:"suggested_order,omitempty"` // Ordem sugerida pela IA
}

// Video represents a property video (AI_DEV_DIRECTIVE Seção 23)
type Video struct {
	ID           string    `firestore:"id" json:"id"`
	URL          string    `firestore:"url" json:"url"`                                   // GCS URL
	ThumbnailURL string    `firestore:"thumbnail_url" json:"thumbnail_url"`               // Frame do meio (gerado por ffmpeg)
	Duration     int       `firestore:"duration" json:"duration"`                         // Duração em segundos
	Source       string    `firestore:"source" json:"source"`                             // "upload", "youtube", "instagram"
	SourceURL    string    `firestore:"source_url,omitempty" json:"source_url,omitempty"` // URL original (se externo)
	Order        int       `firestore:"order" json:"order"`
	CreatedAt    time.Time `firestore:"created_at" json:"created_at"`
}
