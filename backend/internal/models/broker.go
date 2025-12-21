package models

import "time"

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

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
