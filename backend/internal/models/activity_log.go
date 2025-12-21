package models

import "time"

// ActivityLog represents an audit log entry for tracking all important events
// Collection: /tenants/{tenantId}/activity_logs/{logId}
// Ver AI_DEV_DIRECTIVE.md seção 17 para especificação completa
//
// Eventos OBRIGATÓRIOS no MVP:
// - property_created
// - property_updated
// - listing_created
// - listing_updated
// - canonical_listing_assigned
// - canonical_listing_changed
// - owner_placeholder_created
// - owner_enriched (quando dados forem adicionados)
// - lead_created_whatsapp
// - lead_created_form
// - lead_status_changed
// - co_broker_added
// - co_broker_removed
type ActivityLog struct {
	ID       string `firestore:"-" json:"id"`
	TenantID string `firestore:"tenant_id" json:"tenant_id"`

	// Identificação determinística
	EventID   string `firestore:"event_id" json:"event_id"`     // hash(entityId + action + timestamp_bucket_5min)
	EventHash string `firestore:"event_hash" json:"event_hash"` // SHA256 do payload normalizado
	RequestID string `firestore:"request_id" json:"request_id"` // UUID v4 por request HTTP

	// Evento
	EventType string `firestore:"event_type" json:"event_type"` // ex: property_created, lead_created_whatsapp

	// Ator
	ActorType ActorType `firestore:"actor_type" json:"actor_type"` // user, system, owner
	ActorID   string    `firestore:"actor_id,omitempty" json:"actor_id,omitempty"`

	// Dados flexíveis
	// Metadata contém informações específicas do evento
	// Exemplos:
	// - property_created: {"property_id": "...", "broker_id": "...", "external_source": "union"}
	// - lead_created_whatsapp: {"lead_id": "...", "property_id": "...", "consent_given": true, "consent_ip": "..."}
	// - canonical_listing_changed: {"property_id": "...", "old_listing_id": "...", "new_listing_id": "..."}
	Metadata map[string]interface{} `firestore:"metadata" json:"metadata"`

	// Timestamp
	Timestamp time.Time `firestore:"timestamp" json:"timestamp"`
}
