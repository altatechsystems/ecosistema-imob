package models

import "time"

// Owner represents a property owner (proprietário)
// Collection: /tenants/{tenantId}/owners/{ownerId}
// IMPORTANTE: Owner é PASSIVO no MVP - pode ser incompleto
type Owner struct {
	ID       string `firestore:"-" json:"id"`
	TenantID string `firestore:"tenant_id" json:"tenant_id"`

	// Dados pessoais (PODEM ser incompletos no MVP)
	Name         string `firestore:"name,omitempty" json:"name,omitempty"`
	Email        string `firestore:"email,omitempty" json:"email,omitempty"`
	Phone        string `firestore:"phone,omitempty" json:"phone,omitempty"`
	Document     string `firestore:"document,omitempty" json:"document,omitempty"`           // CPF/CNPJ
	DocumentType string `firestore:"document_type,omitempty" json:"document_type,omitempty"` // "cpf", "cnpj"

	// Completude dos dados
	OwnerStatus OwnerStatus `firestore:"owner_status" json:"owner_status"` // incomplete, partial, verified

	// LGPD - Consentimento e Origem (AI_DEV_DIRECTIVE Seção 21)
	ConsentGiven  bool       `firestore:"consent_given" json:"consent_given"`                   // default: false para placeholders
	ConsentText   string     `firestore:"consent_text,omitempty" json:"consent_text,omitempty"` // Texto exibido
	ConsentDate   *time.Time `firestore:"consent_date,omitempty" json:"consent_date,omitempty"`
	ConsentOrigin string     `firestore:"consent_origin,omitempty" json:"consent_origin,omitempty"` // broker, self_service, xls_import, manual_entry

	// LGPD - Anonimização
	IsAnonymized        bool       `firestore:"is_anonymized" json:"is_anonymized"` // default: false
	AnonymizedAt        *time.Time `firestore:"anonymized_at,omitempty" json:"anonymized_at,omitempty"`
	AnonymizationReason string     `firestore:"anonymization_reason,omitempty" json:"anonymization_reason,omitempty"` // retention_policy, user_request

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
