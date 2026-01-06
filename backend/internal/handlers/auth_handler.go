package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"cloud.google.com/go/firestore"
)

type AuthHandler struct {
	firebaseAuth *auth.Client
	firestoreDB  *firestore.Client
}

func NewAuthHandler(firebaseAuth *auth.Client, firestoreDB *firestore.Client) *AuthHandler {
	return &AuthHandler{
		firebaseAuth: firebaseAuth,
		firestoreDB:  firestoreDB,
	}
}

// SignupRequest represents the signup payload
type SignupRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Name       string `json:"name" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	TenantName string `json:"tenant_name" binding:"required"`

	// PROMPT 10: Differentiate between brokers and administrative users
	IsBroker bool   `json:"is_broker"` // true = create as broker, false = create as admin user
	CRECI    string `json:"creci"`     // Required if is_broker=true
}

// SignupResponse represents the signup response
type SignupResponse struct {
	TenantID      string                 `json:"tenant_id"`
	BrokerID      string                 `json:"broker_id"`
	FirebaseToken string                 `json:"firebase_token"`
	User          map[string]interface{} `json:"user"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	FirebaseToken   string                 `json:"firebase_token"`
	TenantID        string                 `json:"tenant_id"`
	Broker          map[string]interface{} `json:"broker"`
	IsPlatformAdmin bool                   `json:"is_platform_admin"`
}

// Signup creates a new tenant and admin broker
func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// 1. Check if email already exists
	_, err := h.firebaseAuth.GetUserByEmail(ctx, req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// 2. Create Firebase Auth user
	userParams := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		DisplayName(req.Name).
		EmailVerified(false)

	userRecord, err := h.firebaseAuth.CreateUser(ctx, userParams)
	if err != nil {
		log.Printf("Error creating Firebase user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 3. Create Tenant
	tenantID := uuid.New().String()
	slug := generateSlug(req.TenantName)

	tenant := models.Tenant{
		ID:        tenantID,
		Name:      req.TenantName,
		Slug:      slug,
		Email:     req.Email,
		Phone:     req.Phone,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = h.firestoreDB.Collection("tenants").Doc(tenantID).Set(ctx, tenant)
	if err != nil {
		log.Printf("Error creating tenant: %v", err)
		// Rollback: delete Firebase user
		h.firebaseAuth.DeleteUser(ctx, userRecord.UID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// 4. PROMPT 10: Create either Broker or User based on is_broker flag
	var entityID string
	var role string

	if req.IsBroker {
		// Create as Broker (requires CRECI)
		if req.CRECI == "" {
			log.Printf("CRECI is required for broker signup")
			// Rollback: delete tenant and Firebase user
			h.firestoreDB.Collection("tenants").Doc(tenantID).Delete(ctx)
			h.firebaseAuth.DeleteUser(ctx, userRecord.UID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI is required for brokers"})
			return
		}

		brokerID := uuid.New().String()
		broker := models.Broker{
			ID:          brokerID,
			TenantID:    tenantID,
			FirebaseUID: userRecord.UID,
			Name:        req.Name,
			Email:       req.Email,
			Phone:       req.Phone,
			CRECI:       req.CRECI,
			Role:        "broker_admin", // First broker is broker_admin
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err = h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("brokers").Doc(brokerID).Set(ctx, broker)
		if err != nil {
			log.Printf("Error creating broker: %v", err)
			// Rollback: delete tenant and Firebase user
			h.firestoreDB.Collection("tenants").Doc(tenantID).Delete(ctx)
			h.firebaseAuth.DeleteUser(ctx, userRecord.UID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create broker"})
			return
		}

		entityID = brokerID
		role = "broker_admin"
		log.Printf("✅ Created broker with CRECI: %s", req.CRECI)
	} else {
		// Create as User (administrative user without CRECI)
		userID := uuid.New().String()
		user := models.User{
			ID:          userID,
			TenantID:    tenantID,
			FirebaseUID: userRecord.UID,
			Name:        req.Name,
			Email:       req.Email,
			Phone:       req.Phone,
			Role:        "admin", // First user is always admin
			IsActive:    true,
			Permissions: []string{
				"properties.view_all",
				"properties.create",
				"properties.edit_all",
				"properties.delete",
				"brokers.view",
				"brokers.create",
				"brokers.edit",
				"users.view",
				"users.create",
				"users.edit",
				"settings.view",
				"settings.edit",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("users").Doc(userID).Set(ctx, user)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			// Rollback: delete tenant and Firebase user
			h.firestoreDB.Collection("tenants").Doc(tenantID).Delete(ctx)
			h.firebaseAuth.DeleteUser(ctx, userRecord.UID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		entityID = userID
		role = "admin"
		log.Printf("✅ Created administrative user (no CRECI)")
	}

	// 5. Set custom claims
	claims := map[string]interface{}{
		"tenant_id": tenantID,
		"role":      role,
		"broker_id": entityID, // For backwards compatibility
		"user_id":   entityID, // For users
	}

	err = h.firebaseAuth.SetCustomUserClaims(ctx, userRecord.UID, claims)
	if err != nil {
		log.Printf("Error setting custom claims: %v", err)
		// Continue anyway, claims can be set later
	}

	// 6. Generate custom token
	token, err := h.firebaseAuth.CustomToken(ctx, userRecord.UID)
	if err != nil {
		log.Printf("Error creating custom token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// 7. Log activity
	go h.logActivity(ctx, tenantID, "tenant_created", map[string]interface{}{
		"tenant_id":   tenantID,
		"tenant_name": req.TenantName,
		"admin_email": req.Email,
		"is_broker":   req.IsBroker,
	})

	if req.IsBroker {
		go h.logActivity(ctx, tenantID, "broker_created", map[string]interface{}{
			"broker_id":    entityID,
			"broker_email": req.Email,
			"role":         role,
			"creci":        req.CRECI,
		})
	} else {
		go h.logActivity(ctx, tenantID, "user_created", map[string]interface{}{
			"user_id":    entityID,
			"user_email": req.Email,
			"role":       role,
		})
	}

	// 8. Return response
	c.JSON(http.StatusCreated, SignupResponse{
		TenantID:      tenantID,
		BrokerID:      entityID, // For backwards compatibility, this contains either broker_id or user_id
		FirebaseToken: token,
		User: map[string]interface{}{
			"uid":   userRecord.UID,
			"email": req.Email,
			"name":  req.Name,
			"role":  role,
		},
	})
}

// Login authenticates a user
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// 1. Get user by email
	userRecord, err := h.firebaseAuth.GetUserByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 2. Try to find in /brokers collection first (real estate agents with CRECI)
	log.Printf("Looking for broker with firebase_uid: %s", userRecord.UID)
	brokersQuery := h.firestoreDB.CollectionGroup("brokers").
		Where("firebase_uid", "==", userRecord.UID).
		Limit(1)

	brokerDocs, err := brokersQuery.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error querying brokers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query brokers"})
		return
	}

	// 3. If not found in /brokers, try /users collection (administrative users without CRECI)
	var tenantID string
	var role string
	var entityID string
	var entityName string
	var entityEmail string
	var isActive bool

	if len(brokerDocs) > 0 {
		// Found as Broker
		brokerDoc := brokerDocs[0]
		var broker models.Broker
		if err := brokerDoc.DataTo(&broker); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load broker data"})
			return
		}
		broker.ID = brokerDoc.Ref.ID

		tenantID = broker.TenantID
		role = broker.Role
		entityID = broker.ID
		entityName = broker.Name
		entityEmail = broker.Email
		isActive = broker.IsActive

		log.Printf("✅ Found as Broker: %s (tenant: %s, role: %s)", broker.ID, broker.TenantID, broker.Role)
	} else {
		// Not found in brokers, search in users
		log.Printf("Not found in brokers, searching in users...")
		usersQuery := h.firestoreDB.CollectionGroup("users").
			Where("firebase_uid", "==", userRecord.UID).
			Limit(1)

		userDocs, err := usersQuery.Documents(ctx).GetAll()
		if err != nil {
			log.Printf("Error querying users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query users"})
			return
		}

		if len(userDocs) == 0 {
			log.Printf("❌ User not found in brokers or users for firebase_uid: %s (email: %s)", userRecord.UID, userRecord.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found. Please contact your administrator."})
			return
		}

		// Found as User
		userDoc := userDocs[0]
		var user models.User
		if err := userDoc.DataTo(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
			return
		}
		user.ID = userDoc.Ref.ID

		tenantID = user.TenantID
		role = user.Role
		entityID = user.ID
		entityName = user.Name
		entityEmail = user.Email
		isActive = user.IsActive

		log.Printf("✅ Found as User: %s (tenant: %s, role: %s)", user.ID, user.TenantID, user.Role)
	}

	// 4. Check if account is active
	if !isActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is inactive"})
		return
	}

	// 5. Check if tenant is active
	tenantDoc, err := h.firestoreDB.Collection("tenants").Doc(tenantID).Get(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tenant data"})
		return
	}

	var tenant models.Tenant
	if err := tenantDoc.DataTo(&tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tenant data"})
		return
	}

	if !tenant.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tenant account is inactive"})
		return
	}

	// 6. Generate custom token with claims
	// Note: We use "broker_id" claim for both brokers and users for backwards compatibility
	// This will be renamed to "entity_id" in a future version
	claims := map[string]interface{}{
		"tenant_id": tenantID,
		"broker_id": entityID, // For backwards compatibility
		"user_id":   entityID, // New field for users
		"role":      role,
	}

	token, err := h.firebaseAuth.CustomTokenWithClaims(ctx, userRecord.UID, claims)
	if err != nil {
		log.Printf("Error creating custom token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	log.Printf("✅ Custom token generated successfully for %s (tenant: %s, role: %s)", entityID, tenantID, role)
	log.Printf("   Token length: %d", len(token))

	// 7. Return response
	c.JSON(http.StatusOK, LoginResponse{
		FirebaseToken:   token,
		TenantID:        tenantID,
		IsPlatformAdmin: tenant.IsPlatformAdmin,
		Broker: map[string]interface{}{
			"id":    entityID,
			"name":  entityName,
			"email": entityEmail,
			"role":  role,
		},
	})
}

// RefreshToken refreshes the user's token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := context.Background()

	// Generate new custom token
	token, err := h.firebaseAuth.CustomToken(ctx, userID.(string))
	if err != nil {
		log.Printf("Error creating custom token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Get broker info
	tenantID, _ := c.Get("tenant_id")
	brokerID, _ := c.Get("broker_id")

	c.JSON(http.StatusOK, gin.H{
		"firebase_token": token,
		"tenant_id":      tenantID,
		"broker_id":      brokerID,
	})
}

// Helper: Generate slug from tenant name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only letters, numbers, hyphens)
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	slug = result.String()

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()
	slug = fmt.Sprintf("%s-%d", slug, timestamp)

	return slug
}

// Helper: Log activity
func (h *AuthHandler) logActivity(ctx context.Context, tenantID, eventType string, metadata map[string]interface{}) {
	activityLog := map[string]interface{}{
		"tenant_id":  tenantID,
		"event_type": eventType,
		"metadata":   metadata,
		"timestamp":  time.Now(),
	}

	_, err := h.firestoreDB.Collection("tenants").Doc(tenantID).Collection("activity_logs").NewDoc().Set(ctx, activityLog)
	if err != nil {
		log.Printf("Error logging activity: %v", err)
	}
}
