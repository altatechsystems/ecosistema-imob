package models

import "time"

// OwnerConfirmationToken represents a secure token for passive owner confirmation
// Collection: /tenants/{tenantId}/owner_confirmation_tokens/{tokenId}
// IMPORTANTE: Owner pode estar incompleto/ausente (PROMPT 08)
type OwnerConfirmationToken struct {
	ID       string `firestore:"-" json:"id"`
	TenantID string `firestore:"tenant_id" json:"tenant_id"`

	// Relacionamentos
	PropertyID string  `firestore:"property_id" json:"property_id"`           // obrigatório
	OwnerID    *string `firestore:"owner_id,omitempty" json:"owner_id,omitempty"` // opcional (Owner pode estar incompleto)

	// Token (armazenar HASH, não o token puro)
	TokenHash string `firestore:"token_hash" json:"token_hash"` // SHA-256 hash do token

	// Validade
	ExpiresAt time.Time `firestore:"expires_at" json:"expires_at"` // ex: 7 dias após criação
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`

	// Auditoria
	CreatedByActorID   string     `firestore:"created_by_actor_id" json:"created_by_actor_id"`     // operador que gerou
	CreatedByActorType ActorType  `firestore:"created_by_actor_type" json:"created_by_actor_type"` // user
	UsedAt             *time.Time `firestore:"used_at,omitempty" json:"used_at,omitempty"`
	LastAction         string     `firestore:"last_action,omitempty" json:"last_action,omitempty"` // confirm_available, confirm_unavailable, confirm_price

	// Metadados opcionais
	DeliveryHint string `firestore:"delivery_hint,omitempty" json:"delivery_hint,omitempty"` // whatsapp, sms, email (apenas anotação)

	// Snapshot mínimo do Owner (opcional, para auditoria)
	// NÃO obrigatório no MVP - pode estar vazio se Owner incompleto
	OwnerSnapshot *OwnerSnapshotMinimal `firestore:"owner_snapshot,omitempty" json:"owner_snapshot,omitempty"`
}

// OwnerSnapshotMinimal é uma versão mínima/mascarada dos dados do Owner
// para auditoria e exibição segura (sem expor dados sensíveis)
type OwnerSnapshotMinimal struct {
	Name  string `firestore:"name,omitempty" json:"name,omitempty"`   // pode ser mascarado "João S."
	Phone string `firestore:"phone,omitempty" json:"phone,omitempty"` // mascarado "(11) 9****-1234"
	Email string `firestore:"email,omitempty" json:"email,omitempty"` // mascarado "j***@example.com"
}

// ConfirmationAction defines the type of confirmation action
type ConfirmationAction string

const (
	ConfirmationActionAvailable   ConfirmationAction = "confirm_available"
	ConfirmationActionUnavailable ConfirmationAction = "confirm_unavailable"
	ConfirmationActionPrice       ConfirmationAction = "confirm_price"
)
