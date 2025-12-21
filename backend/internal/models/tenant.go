package models

import "time"

// Tenant represents a real estate agency (imobiliária) in the multi-tenant system
// Collection: /tenants/{tenantId}
type Tenant struct {
	ID   string `firestore:"-" json:"id"`
	Name string `firestore:"name" json:"name"`
	Slug string `firestore:"slug" json:"slug"` // SEO-friendly identifier

	// Contact information
	Email string `firestore:"email,omitempty" json:"email,omitempty"`
	Phone string `firestore:"phone,omitempty" json:"phone,omitempty"`

	// Business information
	Document     string `firestore:"document,omitempty" json:"document,omitempty"`           // CNPJ
	DocumentType string `firestore:"document_type,omitempty" json:"document_type,omitempty"` // "cnpj"
	CRECI        string `firestore:"creci,omitempty" json:"creci,omitempty"`                 // CRECI Pessoa Jurídica

	// Address
	Street       string `firestore:"street,omitempty" json:"street,omitempty"`
	Number       string `firestore:"number,omitempty" json:"number,omitempty"`
	Complement   string `firestore:"complement,omitempty" json:"complement,omitempty"`
	Neighborhood string `firestore:"neighborhood,omitempty" json:"neighborhood,omitempty"`
	City         string `firestore:"city,omitempty" json:"city,omitempty"`
	State        string `firestore:"state,omitempty" json:"state,omitempty"` // UF
	ZipCode      string `firestore:"zip_code,omitempty" json:"zip_code,omitempty"`
	Country      string `firestore:"country,omitempty" json:"country,omitempty"` // default "BR"

	// Settings
	IsActive bool `firestore:"is_active" json:"is_active"`

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
