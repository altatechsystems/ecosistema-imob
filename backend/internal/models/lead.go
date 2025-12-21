package models

import "time"

// Lead represents a potential customer interested in a property
// Collection: /tenants/{tenantId}/leads/{leadId}
// IMPORTANTE: Lead SEMPRE pertence ao Property, NUNCA diretamente ao corretor
type Lead struct {
	ID         string `firestore:"-" json:"id"`
	TenantID   string `firestore:"tenant_id" json:"tenant_id"`
	PropertyID string `firestore:"property_id" json:"property_id"` // ref Property (OBRIGATÓRIO)

	// Dados do interessado (MÍNIMOS)
	Name    string `firestore:"name,omitempty" json:"name,omitempty"`
	Email   string `firestore:"email,omitempty" json:"email,omitempty"`
	Phone   string `firestore:"phone,omitempty" json:"phone,omitempty"`
	Message string `firestore:"message,omitempty" json:"message,omitempty"`

	// Origem
	Channel     LeadChannel `firestore:"channel" json:"channel"` // whatsapp, form, phone, email
	UTMSource   string      `firestore:"utm_source,omitempty" json:"utm_source,omitempty"`
	UTMCampaign string      `firestore:"utm_campaign,omitempty" json:"utm_campaign,omitempty"`
	UTMMedium   string      `firestore:"utm_medium,omitempty" json:"utm_medium,omitempty"`
	Referrer    string      `firestore:"referrer,omitempty" json:"referrer,omitempty"` // URL da página

	// Status
	Status LeadStatus `firestore:"status" json:"status"` // new, contacted, qualified, lost

	// LGPD - Consentimento (AI_DEV_DIRECTIVE Seção 21)
	// OBRIGATÓRIO: consent_given DEVE ser true para criar lead
	ConsentGiven   bool       `firestore:"consent_given" json:"consent_given"`               // OBRIGATÓRIO para criar lead
	ConsentText    string     `firestore:"consent_text" json:"consent_text"`                 // Texto exibido no checkbox
	ConsentDate    time.Time  `firestore:"consent_date" json:"consent_date"`                 // Timestamp do consentimento
	ConsentIP      string     `firestore:"consent_ip,omitempty" json:"consent_ip,omitempty"` // IP do usuário (se disponível)
	ConsentRevoked bool       `firestore:"consent_revoked" json:"consent_revoked"`           // default: false
	RevokedAt      *time.Time `firestore:"revoked_at,omitempty" json:"revoked_at,omitempty"` // Timestamp da revogação

	// LGPD - Anonimização
	IsAnonymized        bool       `firestore:"is_anonymized" json:"is_anonymized"` // default: false
	AnonymizedAt        *time.Time `firestore:"anonymized_at,omitempty" json:"anonymized_at,omitempty"`
	AnonymizationReason string     `firestore:"anonymization_reason,omitempty" json:"anonymization_reason,omitempty"` // retention_policy, user_request

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
