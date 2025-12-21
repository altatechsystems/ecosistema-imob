package models

import "time"

// PropertyBrokerRole represents the relationship between a broker and a property
// Collection: /tenants/{tenantId}/property_broker_roles/{roleId}
// IMPORTANTE: Implementa co-corretagem conforme AI_DEV_DIRECTIVE.md seção 4
//
// REGRAS DE NEGÓCIO:
// 1. Todo Property DEVE ter exatamente 1 originating_broker (captador)
// 2. Todo Listing DEVE criar 1 listing_broker (vendedor)
// 3. Pode haver N co_broker adicionados durante negociação
// 4. Apenas 1 PropertyBrokerRole pode ter is_primary: true (roteamento de leads)
// 5. Comissão é apenas registro, SEM cálculo ou split no MVP
type PropertyBrokerRole struct {
	ID         string `firestore:"-" json:"id"`
	TenantID   string `firestore:"tenant_id" json:"tenant_id"`
	PropertyID string `firestore:"property_id" json:"property_id"` // ref Property
	BrokerID   string `firestore:"broker_id" json:"broker_id"`     // ref Broker

	// Papel do corretor
	// - originating_broker: corretor que captou/originou o imóvel (único por Property)
	// - listing_broker: corretor responsável por um Listing (pode haver múltiplos)
	// - co_broker: corretor adicional na negociação (comum no Brasil)
	Role BrokerPropertyRole `firestore:"role" json:"role"`

	// Comissão (apenas registro, SEM processamento no MVP)
	CommissionPercentage float64 `firestore:"commission_percentage,omitempty" json:"commission_percentage,omitempty"`

	// Primary (para roteamento de leads)
	// Apenas 1 PropertyBrokerRole pode ter is_primary: true
	// Define quem recebe leads primeiro
	IsPrimary bool `firestore:"is_primary" json:"is_primary"`

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
