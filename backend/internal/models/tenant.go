package models

import "time"

// Tenant represents a real estate business (imobiliária, incorporadora, loteadora, construtora, corretor autônomo) in the multi-tenant system
// Collection: /tenants/{tenantId}
type Tenant struct {
	ID   string `firestore:"-" json:"id"`
	Name string `firestore:"name" json:"name"`
	Slug string `firestore:"slug" json:"slug"` // SEO-friendly identifier

	// Contact information
	Email string `firestore:"email,omitempty" json:"email,omitempty"`
	Phone string `firestore:"phone,omitempty" json:"phone,omitempty"`

	// Business information
	TenantType   string `firestore:"tenant_type,omitempty" json:"tenant_type,omitempty"`     // "pf" or "pj"
	Document     string `firestore:"document,omitempty" json:"document,omitempty"`           // CPF or CNPJ
	DocumentType string `firestore:"document_type,omitempty" json:"document_type,omitempty"` // "cpf" or "cnpj"
	BusinessType string `firestore:"business_type,omitempty" json:"business_type,omitempty"` // imobiliaria, incorporadora, loteadora, construtora, corretor_autonomo
	CRECI        string `firestore:"creci,omitempty" json:"creci,omitempty"`                 // CRECI (Pessoa Física ou Jurídica)

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
	Settings        map[string]interface{} `firestore:"settings,omitempty" json:"settings,omitempty"`
	IsActive        bool                   `firestore:"is_active" json:"is_active"`
	IsPlatformAdmin bool                   `firestore:"is_platform_admin,omitempty" json:"is_platform_admin,omitempty"`

	// Subscription
	SubscriptionPlan      string     `firestore:"subscription_plan,omitempty" json:"subscription_plan,omitempty"`           // "free", "full"
	SubscriptionStatus    string     `firestore:"subscription_status,omitempty" json:"subscription_status,omitempty"`       // "active", "trial", "expired", "cancelled"
	TrialEndsAt           *time.Time `firestore:"trial_ends_at,omitempty" json:"trial_ends_at,omitempty"`                   // Trial end date
	SubscriptionStartedAt *time.Time `firestore:"subscription_started_at,omitempty" json:"subscription_started_at,omitempty"` // Subscription start date

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
