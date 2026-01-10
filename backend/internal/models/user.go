package models

import (
	"encoding/json"
	"time"
)

// User represents a system user (can be admin, manager, broker, or broker_admin)
// Collection: /tenants/{tenantId}/users/{userId}
// Users can have different roles:
// - "admin": Administrator without CRECI
// - "manager": Manager without CRECI
// - "broker": Real estate broker with CRECI
// - "broker_admin": Broker with CRECI who is also tenant administrator
type User struct {
	ID       string `firestore:"-" json:"id"`
	TenantID string `firestore:"tenant_id" json:"tenant_id"`

	// Firebase Auth UID (from Firebase Authentication)
	FirebaseUID string `firestore:"firebase_uid" json:"firebase_uid"`

	// Personal information
	Name  string `firestore:"name" json:"name"`
	Email string `firestore:"email" json:"email"`
	Phone string `firestore:"phone,omitempty" json:"phone,omitempty"`

	// Document information (optional for admin users)
	Document     string `firestore:"document,omitempty" json:"document,omitempty"`           // CPF ou CNPJ
	DocumentType string `firestore:"document_type,omitempty" json:"document_type,omitempty"` // "cpf" ou "cnpj"

	// CRECI (optional - only for users who are also brokers)
	CRECI string `firestore:"creci,omitempty" json:"creci,omitempty"` // Format: "XXXXX-F/UF"

	// Role and status
	// Valid roles: "admin", "manager", "broker", "broker_admin"
	Role     string `firestore:"role,omitempty" json:"role,omitempty"` // "admin", "manager", "broker", "broker_admin"
	IsActive bool   `firestore:"is_active" json:"is_active"`

	// Permissions (array of permission strings)
	// Examples: "properties.view", "properties.edit", "brokers.manage", "settings.edit"
	Permissions []string `firestore:"permissions,omitempty" json:"permissions,omitempty"`

	// Profile
	PhotoURL string `firestore:"photo_url,omitempty" json:"photo_url,omitempty"`

	// Metadata - using interface{} to handle both time.Time and string from Firestore
	CreatedAt interface{} `firestore:"created_at" json:"created_at"`
	UpdatedAt interface{} `firestore:"updated_at" json:"updated_at"`
}

// GetCreatedAt returns created_at as time.Time
func (u *User) GetCreatedAt() time.Time {
	return parseFlexibleTimeUser(u.CreatedAt)
}

// GetUpdatedAt returns updated_at as time.Time
func (u *User) GetUpdatedAt() time.Time {
	return parseFlexibleTimeUser(u.UpdatedAt)
}

// parseFlexibleTimeUser converts interface{} to time.Time
func parseFlexibleTimeUser(val interface{}) time.Time {
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
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		*Alias
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		Alias:     (*Alias)(&u),
		CreatedAt: u.GetCreatedAt(),
		UpdatedAt: u.GetUpdatedAt(),
	})
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission string) bool {
	// Admin role has all permissions
	if u.Role == "admin" {
		return true
	}

	// Check in permissions array
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// AddPermission adds a permission to the user
func (u *User) AddPermission(permission string) {
	if !u.HasPermission(permission) {
		u.Permissions = append(u.Permissions, permission)
	}
}

// RemovePermission removes a permission from the user
func (u *User) RemovePermission(permission string) {
	for i, p := range u.Permissions {
		if p == permission {
			u.Permissions = append(u.Permissions[:i], u.Permissions[i+1:]...)
			break
		}
	}
}

// ValidRoles returns the list of valid roles for users
func ValidUserRoles() []string {
	return []string{"admin", "manager", "broker", "broker_admin"}
}

// IsValidRole checks if a role is valid for admin users
func IsValidUserRole(role string) bool {
	for _, validRole := range ValidUserRoles() {
		if role == validRole {
			return true
		}
	}
	return false
}
