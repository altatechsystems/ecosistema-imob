package models

import (
	"encoding/json"
	"time"
)

// UserInvitation represents an invitation to join a tenant
// Collection: /tenants/{tenantId}/user_invitations/{invitationId}
type UserInvitation struct {
	ID       string `firestore:"-" json:"id"`
	TenantID string `firestore:"tenant_id" json:"tenant_id"`

	// Invitee information
	Email string `firestore:"email" json:"email"`
	Name  string `firestore:"name" json:"name"`
	Phone string `firestore:"phone,omitempty" json:"phone,omitempty"`

	// Role and permissions
	Role        string   `firestore:"role" json:"role"`                                 // "admin", "manager", "broker", "broker_admin"
	Permissions []string `firestore:"permissions,omitempty" json:"permissions,omitempty"` // Array of permission strings
	CRECI       string   `firestore:"creci,omitempty" json:"creci,omitempty"`           // Required if role includes "broker"

	// Invitation metadata
	InvitedBy string `firestore:"invited_by" json:"invited_by"` // User ID of the admin who sent the invitation
	Token     string `firestore:"token" json:"-"`               // Unique token for accepting invitation (not exposed in JSON)

	// Status and timestamps
	Status     string      `firestore:"status" json:"status"`           // "pending", "accepted", "expired", "cancelled"
	ExpiresAt  interface{} `firestore:"expires_at" json:"expires_at"`   // 7 days from creation
	CreatedAt  interface{} `firestore:"created_at" json:"created_at"`
	AcceptedAt interface{} `firestore:"accepted_at,omitempty" json:"accepted_at,omitempty"`

	// Reference to created user (after acceptance)
	UserID string `firestore:"user_id,omitempty" json:"user_id,omitempty"`
}

// GetExpiresAt returns expires_at as time.Time
func (i *UserInvitation) GetExpiresAt() time.Time {
	return parseFlexibleTimeInvitation(i.ExpiresAt)
}

// GetCreatedAt returns created_at as time.Time
func (i *UserInvitation) GetCreatedAt() time.Time {
	return parseFlexibleTimeInvitation(i.CreatedAt)
}

// GetAcceptedAt returns accepted_at as time.Time
func (i *UserInvitation) GetAcceptedAt() time.Time {
	return parseFlexibleTimeInvitation(i.AcceptedAt)
}

// IsExpired checks if the invitation has expired
func (i *UserInvitation) IsExpired() bool {
	return time.Now().After(i.GetExpiresAt())
}

// IsPending checks if the invitation is pending
func (i *UserInvitation) IsPending() bool {
	return i.Status == "pending" && !i.IsExpired()
}

// parseFlexibleTimeInvitation converts interface{} to time.Time
func parseFlexibleTimeInvitation(val interface{}) time.Time {
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
func (i UserInvitation) MarshalJSON() ([]byte, error) {
	type Alias UserInvitation
	return json.Marshal(&struct {
		*Alias
		ExpiresAt  time.Time  `json:"expires_at"`
		CreatedAt  time.Time  `json:"created_at"`
		AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	}{
		Alias:     (*Alias)(&i),
		ExpiresAt: i.GetExpiresAt(),
		CreatedAt: i.GetCreatedAt(),
		AcceptedAt: func() *time.Time {
			t := i.GetAcceptedAt()
			if t.IsZero() {
				return nil
			}
			return &t
		}(),
	})
}

// ValidInvitationStatuses returns the list of valid invitation statuses
func ValidInvitationStatuses() []string {
	return []string{"pending", "accepted", "expired", "cancelled"}
}

// IsValidInvitationStatus checks if a status is valid
func IsValidInvitationStatus(status string) bool {
	for _, validStatus := range ValidInvitationStatuses() {
		if status == validStatus {
			return true
		}
	}
	return false
}
