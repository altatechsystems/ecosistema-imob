package models

import (
	"encoding/json"
	"time"
)

// Broker represents a real estate broker (corretor de imóveis)
// Collection: /tenants/{tenantId}/brokers/{brokerId}
type Broker struct {
	ID       string `firestore:"-" json:"id"`
	TenantID string `firestore:"tenant_id" json:"tenant_id"`

	// Firebase Auth UID (from Firebase Authentication)
	FirebaseUID string `firestore:"firebase_uid" json:"firebase_uid"`

	// Personal information
	Name  string `firestore:"name" json:"name"`
	Email string `firestore:"email" json:"email"`
	Phone string `firestore:"phone,omitempty" json:"phone,omitempty"`

	// CRECI (Conselho Regional de Corretores de Imóveis) - OBRIGATÓRIO
	// Formato: XXXXX-F/UF ou XXXXX/UF
	// Exemplo: "12345-J/SP" ou "00123/RJ"
	CRECI string `firestore:"creci" json:"creci"`

	// Document information
	Document     string `firestore:"document,omitempty" json:"document,omitempty"`           // CPF ou CNPJ
	DocumentType string `firestore:"document_type,omitempty" json:"document_type,omitempty"` // "cpf" ou "cnpj"

	// Role and status
	Role     string `firestore:"role,omitempty" json:"role,omitempty"` // "admin", "broker", "manager"
	IsActive bool   `firestore:"is_active" json:"is_active"`

	// Profile (Public Profile - similar to Zillow)
	PhotoURL    string `firestore:"photo_url,omitempty" json:"photo_url,omitempty"`
	Bio         string `firestore:"bio,omitempty" json:"bio,omitempty"`                   // Biografia do corretor
	Specialties string `firestore:"specialties,omitempty" json:"specialties,omitempty"`   // Ex: "Buyer's Agent, Listing Agent"
	Languages   string `firestore:"languages,omitempty" json:"languages,omitempty"`       // Ex: "Português, Inglês, Espanhol"
	Experience  int    `firestore:"experience,omitempty" json:"experience,omitempty"`     // Anos de experiência
	Company     string `firestore:"company,omitempty" json:"company,omitempty"`           // Nome da empresa/imobiliária
	Website     string `firestore:"website,omitempty" json:"website,omitempty"`           // Website pessoal
	SocialMedia string `firestore:"social_media,omitempty" json:"social_media,omitempty"` // Links redes sociais (JSON)

	// Statistics (computed/cached for performance)
	TotalSales       int     `firestore:"total_sales,omitempty" json:"total_sales,omitempty"`             // Total de vendas
	TotalListings    int     `firestore:"total_listings,omitempty" json:"total_listings,omitempty"`       // Total de anúncios ativos
	AveragePrice     float64 `firestore:"average_price,omitempty" json:"average_price,omitempty"`         // Preço médio de vendas
	Rating           float64 `firestore:"rating,omitempty" json:"rating,omitempty"`                       // Avaliação média (0-5)
	ReviewCount      int     `firestore:"review_count,omitempty" json:"review_count,omitempty"`           // Número de avaliações
	LastSaleDate     string  `firestore:"last_sale_date,omitempty" json:"last_sale_date,omitempty"`       // Data da última venda
	ServiceAreas     string  `firestore:"service_areas,omitempty" json:"service_areas,omitempty"`         // Áreas de atendimento (JSON array)
	CertificationsAwards string `firestore:"certifications_awards,omitempty" json:"certifications_awards,omitempty"` // Certificações e prêmios

	// Metadata - using interface{} to handle both time.Time and string from Firestore
	CreatedAt interface{} `firestore:"created_at" json:"created_at"`
	UpdatedAt interface{} `firestore:"updated_at" json:"updated_at"`
}

// GetCreatedAt returns created_at as time.Time
func (b *Broker) GetCreatedAt() time.Time {
	return parseFlexibleTime(b.CreatedAt)
}

// GetUpdatedAt returns updated_at as time.Time
func (b *Broker) GetUpdatedAt() time.Time {
	return parseFlexibleTime(b.UpdatedAt)
}

// parseFlexibleTime converts interface{} to time.Time
func parseFlexibleTime(val interface{}) time.Time {
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		// Try RFC3339
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		// Try Go time.Time string format
		if t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", v); err == nil {
			return t
		}
		// Try without timezone
		if t, err := time.Parse("2006-01-02T15:04:05.999999", v); err == nil {
			return t
		}
	}
	return time.Time{}
}

// MarshalJSON implements custom JSON marshaling
func (b Broker) MarshalJSON() ([]byte, error) {
	type Alias Broker
	return json.Marshal(&struct {
		*Alias
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		Alias:     (*Alias)(&b),
		CreatedAt: b.GetCreatedAt(),
		UpdatedAt: b.GetUpdatedAt(),
	})
}
