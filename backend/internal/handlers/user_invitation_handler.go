package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// UserInvitationHandler handles user invitation operations
type UserInvitationHandler struct {
	firebaseAuth *auth.Client
	firestoreDB  *firestore.Client
	emailService *services.EmailService
}

// NewUserInvitationHandler creates a new user invitation handler
func NewUserInvitationHandler(firebaseAuth *auth.Client, firestoreDB *firestore.Client) *UserInvitationHandler {
	return &UserInvitationHandler{
		firebaseAuth: firebaseAuth,
		firestoreDB:  firestoreDB,
		emailService: services.NewEmailService(),
	}
}

// InviteUserRequest represents the request to invite a new user
type InviteUserRequest struct {
	Email       string   `json:"email" binding:"required,email"`
	Name        string   `json:"name" binding:"required"`
	Phone       string   `json:"phone"`
	Role        string   `json:"role" binding:"required"` // "admin", "manager", "broker", "broker_admin"
	Permissions []string `json:"permissions"`
	CRECI       string   `json:"creci"` // Required if role includes "broker"
}

// InviteUserResponse represents the response after creating an invitation
type InviteUserResponse struct {
	InvitationID string `json:"invitation_id"`
	Message      string `json:"message"`
}

// VerifyInvitationResponse represents the response when verifying an invitation
type VerifyInvitationResponse struct {
	Valid      bool                   `json:"valid"`
	Invitation *models.UserInvitation `json:"invitation,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

// AcceptInvitationRequest represents the request to accept an invitation
type AcceptInvitationRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

// AcceptInvitationResponse represents the response after accepting an invitation
type AcceptInvitationResponse struct {
	UserID        string `json:"user_id"`
	TenantID      string `json:"tenant_id"`
	FirebaseToken string `json:"firebase_token"`
	Message       string `json:"message"`
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// InviteUser creates a new user invitation
// POST /api/v1/admin/{tenant_id}/users/invite
func (h *UserInvitationHandler) InviteUser(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	// Get authenticated user from context (set by auth middleware)
	invitedByUID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if !models.IsValidUserRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid role. Must be one of: %v", models.ValidUserRoles()),
		})
		return
	}

	// Validate CRECI for broker roles
	if req.Role == "broker" || req.Role == "broker_admin" {
		if req.CRECI == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI is required for broker roles"})
			return
		}
		// TODO: Validate CRECI format using validators.ValidateCRECI()
	}

	ctx := context.Background()

	// Check if user with this email already exists in the tenant
	usersRef := h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("users")
	existingUsersIter := usersRef.Where("email", "==", req.Email).Documents(ctx)
	_, err := existingUsersIter.Next()
	if err != iterator.Done {
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists in this tenant"})
			return
		}
		log.Printf("❌ Error checking for existing user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing user"})
		return
	}

	// Check if there's already a pending invitation for this email
	invitationsRef := h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("user_invitations")
	pendingInvitationsIter := invitationsRef.Where("email", "==", req.Email).Where("status", "==", "pending").Documents(ctx)
	_, err = pendingInvitationsIter.Next()
	if err != iterator.Done {
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "There is already a pending invitation for this email"})
			return
		}
		log.Printf("❌ Error checking for pending invitations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for pending invitations"})
		return
	}

	// Get tenant name for email
	tenantDoc, err := h.firestoreDB.Collection("tenants").Doc(tenantID).Get(ctx)
	if err != nil {
		log.Printf("❌ Error fetching tenant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant information"})
		return
	}
	var tenant models.Tenant
	if err := tenantDoc.DataTo(&tenant); err != nil {
		log.Printf("❌ Error parsing tenant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tenant information"})
		return
	}

	// Get inviter name for email
	inviterUID := invitedByUID.(string)
	usersIter := usersRef.Where("firebase_uid", "==", inviterUID).Limit(1).Documents(ctx)
	inviterDoc, err := usersIter.Next()
	inviterName := "Administrator"
	if err == nil {
		var inviter models.User
		if err := inviterDoc.DataTo(&inviter); err == nil {
			inviterName = inviter.Name
		}
	}

	// Generate secure token
	token, err := generateSecureToken()
	if err != nil {
		log.Printf("❌ Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invitation token"})
		return
	}

	// Create invitation
	invitationID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days

	invitation := models.UserInvitation{
		ID:          invitationID,
		TenantID:    tenantID,
		Email:       req.Email,
		Name:        req.Name,
		Phone:       req.Phone,
		Role:        req.Role,
		Permissions: req.Permissions,
		CRECI:       req.CRECI,
		InvitedBy:   inviterUID,
		Token:       token,
		Status:      "pending",
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
	}

	// Save to Firestore
	_, err = invitationsRef.Doc(invitationID).Set(ctx, invitation)
	if err != nil {
		log.Printf("❌ Error saving invitation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}

	log.Printf("✅ Invitation created: %s for %s (%s)", invitationID, req.Email, req.Role)

	// Send invitation email
	err = h.emailService.SendInvitation(
		req.Email,
		req.Name,
		token,
		tenantID,
		tenant.Name,
		inviterName,
		req.Role,
		expiresAt,
	)
	if err != nil {
		log.Printf("⚠️ Warning: Invitation created but email failed to send: %v", err)
		// Don't fail the request - invitation is created, email is best-effort
	}

	c.JSON(http.StatusCreated, InviteUserResponse{
		InvitationID: invitationID,
		Message:      fmt.Sprintf("Invitation sent to %s", req.Email),
	})
}

// VerifyInvitation verifies if an invitation token is valid
// GET /api/v1/invitations/{token}/verify
func (h *UserInvitationHandler) VerifyInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	ctx := context.Background()

	// Search for invitation by token across all tenants
	// Note: In production, you might want to optimize this by creating a top-level invitations collection with tenant_id field
	tenantsRef := h.firestoreDB.Collection("tenants")
	tenantsIter := tenantsRef.Documents(ctx)
	defer tenantsIter.Stop()

	// Search through each tenant's invitations
	for {
		tenantDoc, err := tenantsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error iterating tenants: %v", err)
			continue
		}

		invitationsRef := h.firestoreDB.Collection("tenants").Doc(tenantDoc.Ref.ID).Collection("user_invitations")
		invitationsIter := invitationsRef.Where("token", "==", token).Limit(1).Documents(ctx)
		invitationDoc, err := invitationsIter.Next()

		if err == iterator.Done {
			continue
		}
		if err != nil {
			log.Printf("❌ Error querying invitations: %v", err)
			continue
		}

		// Found the invitation
		var invitation models.UserInvitation
		if err := invitationDoc.DataTo(&invitation); err != nil {
			log.Printf("❌ Error parsing invitation: %v", err)
			continue
		}

		invitation.ID = invitationDoc.Ref.ID

		// Check if expired
		if invitation.IsExpired() {
			c.JSON(http.StatusOK, VerifyInvitationResponse{
				Valid:   false,
				Message: "Invitation has expired",
			})
			return
		}

		// Check if not pending
		if invitation.Status != "pending" {
			c.JSON(http.StatusOK, VerifyInvitationResponse{
				Valid:   false,
				Message: fmt.Sprintf("Invitation is %s", invitation.Status),
			})
			return
		}

		// Valid invitation
		c.JSON(http.StatusOK, VerifyInvitationResponse{
			Valid:      true,
			Invitation: &invitation,
		})
		return
	}

	// Token not found
	c.JSON(http.StatusOK, VerifyInvitationResponse{
		Valid:   false,
		Message: "Invalid invitation token",
	})
}

// AcceptInvitation accepts an invitation and creates the user
// POST /api/v1/invitations/{token}/accept
func (h *UserInvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// Parse request
	var req AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Find invitation by token across all tenants
	tenantsRef := h.firestoreDB.Collection("tenants")
	tenantsIter := tenantsRef.Documents(ctx)
	defer tenantsIter.Stop()

	var invitation models.UserInvitation
	var invitationDocRef *firestore.DocumentRef
	var tenantID string
	foundInvitation := false

	// Search through each tenant's invitations
	for {
		tenantDoc, err := tenantsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error iterating tenants: %v", err)
			continue
		}

		invitationsRef := h.firestoreDB.Collection("tenants").Doc(tenantDoc.Ref.ID).Collection("user_invitations")
		invitationsIter := invitationsRef.Where("token", "==", token).Limit(1).Documents(ctx)
		invitationDoc, err := invitationsIter.Next()

		if err == iterator.Done {
			continue
		}
		if err != nil {
			log.Printf("❌ Error querying invitations: %v", err)
			continue
		}

		// Found the invitation
		if err := invitationDoc.DataTo(&invitation); err != nil {
			log.Printf("❌ Error parsing invitation: %v", err)
			continue
		}

		invitation.ID = invitationDoc.Ref.ID
		invitationDocRef = invitationDoc.Ref
		tenantID = tenantDoc.Ref.ID
		foundInvitation = true
		break
	}

	if !foundInvitation {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
		return
	}

	// Validate invitation
	if invitation.IsExpired() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation has expired"})
		return
	}

	if invitation.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invitation is %s", invitation.Status)})
		return
	}

	// Create Firebase user with email and password
	params := (&auth.UserToCreate{}).
		Email(invitation.Email).
		Password(req.Password).
		DisplayName(invitation.Name).
		EmailVerified(true) // Auto-verify since we sent the invitation

	userRecord, err := h.firebaseAuth.CreateUser(ctx, params)
	if err != nil {
		log.Printf("❌ Error creating Firebase user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user account: " + err.Error()})
		return
	}

	log.Printf("✅ Firebase user created: %s (%s)", userRecord.UID, invitation.Email)

	// Create user in Firestore
	userID := uuid.New().String()
	now := time.Now()

	user := models.User{
		ID:          userID,
		TenantID:    tenantID,
		FirebaseUID: userRecord.UID,
		Name:        invitation.Name,
		Email:       invitation.Email,
		Phone:       invitation.Phone,
		CRECI:       invitation.CRECI,
		Role:        invitation.Role,
		IsActive:    true,
		Permissions: invitation.Permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	usersRef := h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("users")
	_, err = usersRef.Doc(userID).Set(ctx, user)
	if err != nil {
		log.Printf("❌ Error creating user in Firestore: %v", err)

		// Rollback: Delete Firebase user
		if deleteErr := h.firebaseAuth.DeleteUser(ctx, userRecord.UID); deleteErr != nil {
			log.Printf("❌ CRITICAL: Failed to rollback Firebase user %s: %v", userRecord.UID, deleteErr)
		} else {
			log.Printf("✅ Rolled back Firebase user creation")
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Set custom claims for the user
	claims := map[string]interface{}{
		"tenant_id": tenantID,
		"role":      invitation.Role,
		"user_id":   userID,
		"broker_id": userID, // Backward compatibility
	}

	err = h.firebaseAuth.SetCustomUserClaims(ctx, userRecord.UID, claims)
	if err != nil {
		log.Printf("⚠️ Warning: User created but failed to set custom claims: %v", err)
	} else {
		log.Printf("✅ Custom claims set for user %s", userRecord.UID)
	}

	// Update invitation status
	invitation.Status = "accepted"
	invitation.AcceptedAt = now
	invitation.UserID = userID

	_, err = invitationDocRef.Set(ctx, invitation)
	if err != nil {
		log.Printf("⚠️ Warning: User created but failed to update invitation status: %v", err)
	}

	log.Printf("✅ User created from invitation: %s (%s) - %s", userID, invitation.Email, invitation.Role)

	// Generate custom token for auto-login
	customToken, err := h.firebaseAuth.CustomToken(ctx, userRecord.UID)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to generate custom token: %v", err)
		customToken = "" // User can still login normally
	}

	c.JSON(http.StatusCreated, AcceptInvitationResponse{
		UserID:        userID,
		TenantID:      tenantID,
		FirebaseToken: customToken,
		Message:       "User created successfully",
	})
}

// ListInvitations lists all invitations for a tenant
// GET /api/v1/admin/{tenant_id}/users/invitations
func (h *UserInvitationHandler) ListInvitations(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	// Optional filter by status
	status := c.Query("status") // "pending", "accepted", "expired", "cancelled"

	ctx := context.Background()
	invitationsRef := h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("user_invitations")

	var invitationsIter *firestore.DocumentIterator

	if status != "" {
		// Filter by status
		invitationsIter = invitationsRef.Where("status", "==", status).Documents(ctx)
	} else {
		// Get all invitations
		invitationsIter = invitationsRef.Documents(ctx)
	}
	defer invitationsIter.Stop()

	invitations := make([]models.UserInvitation, 0)
	for {
		doc, err := invitationsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("⚠️ Error iterating invitations: %v", err)
			continue
		}

		var invitation models.UserInvitation
		if err := doc.DataTo(&invitation); err != nil {
			log.Printf("⚠️ Error parsing invitation: %v", err)
			continue
		}

		invitation.ID = doc.Ref.ID
		invitations = append(invitations, invitation)
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"count":       len(invitations),
	})
}

// CancelInvitation cancels a pending invitation
// DELETE /api/v1/admin/{tenant_id}/users/invitations/{invitation_id}
func (h *UserInvitationHandler) CancelInvitation(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	invitationID := c.Param("invitation_id")

	if tenantID == "" || invitationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id and invitation_id are required"})
		return
	}

	ctx := context.Background()
	invitationRef := h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("user_invitations").Doc(invitationID)

	// Get invitation
	invitationDoc, err := invitationRef.Get(ctx)
	if err != nil {
		log.Printf("❌ Error fetching invitation: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
		return
	}

	var invitation models.UserInvitation
	if err := invitationDoc.DataTo(&invitation); err != nil {
		log.Printf("❌ Error parsing invitation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse invitation"})
		return
	}

	// Can only cancel pending invitations
	if invitation.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot cancel %s invitation", invitation.Status)})
		return
	}

	// Update status to cancelled
	invitation.Status = "cancelled"
	_, err = invitationRef.Set(ctx, invitation)
	if err != nil {
		log.Printf("❌ Error cancelling invitation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel invitation"})
		return
	}

	log.Printf("✅ Invitation cancelled: %s (%s)", invitationID, invitation.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation cancelled successfully",
	})
}
