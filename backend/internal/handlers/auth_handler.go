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
	"github.com/altatech/ecosistema-imob/backend/internal/utils"
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
	// User information (administrador)
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`

	// Tenant information
	TenantName string `json:"tenant_name" binding:"required"`
	TenantType string `json:"tenant_type" binding:"required"` // "pf" or "pj"
	Document   string `json:"document" binding:"required"`    // CPF or CNPJ

	// PJ-specific fields
	BusinessType string `json:"business_type,omitempty"` // Required if tenant_type == "pj"
	TenantCRECI  string `json:"tenant_creci,omitempty"`  // Tenant-level CRECI

	// User role selection
	IsUserBroker bool   `json:"is_user_broker"` // Is the admin also a broker?
	UserCRECI    string `json:"user_creci,omitempty"` // User's personal CRECI if is_user_broker=true
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

// Signup creates a new tenant and admin broker or user
func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Normalize phone to E.164 format
	req.Phone = utils.NormalizePhoneE164(req.Phone, "55")
	if err := utils.ValidatePhoneE164(req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefone inválido: " + err.Error()})
		return
	}

	// 1. Validate tenant_type
	if req.TenantType != "pf" && req.TenantType != "pj" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_type must be 'pf' or 'pj'"})
		return
	}

	// 2. FLUXO PF (Pessoa Física - Corretor Autônomo)
	if req.TenantType == "pf" {
		// 2a. Validar CPF
		if err := utils.ValidateCPF(req.Document); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CPF inválido: " + err.Error()})
			return
		}

		// 2b. CRECI-F obrigatório
		if req.TenantCRECI == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI-F é obrigatório para corretor autônomo"})
			return
		}

		// 2c. Validar formato CRECI-F
		if err := utils.ValidateCRECI(req.TenantCRECI); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI inválido: " + err.Error()})
			return
		}
		if err := utils.ValidateCRECIType(req.TenantCRECI, "F"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 2d. Fixar business_type e configurar broker
		req.BusinessType = "corretor_autonomo"
		req.IsUserBroker = true         // PF é sempre broker
		req.UserCRECI = req.TenantCRECI // Mesmo CRECI
	}

	// 3. FLUXO PJ (Pessoa Jurídica)
	if req.TenantType == "pj" {
		// 3a. Validar CNPJ
		if err := utils.ValidateCNPJ(req.Document); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CNPJ inválido: " + err.Error()})
			return
		}

		// 3b. Business type obrigatório
		if req.BusinessType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "business_type é obrigatório para PJ"})
			return
		}

		validBusinessTypes := []string{"imobiliaria", "incorporadora", "construtora", "loteadora"}
		isValid := false
		for _, bt := range validBusinessTypes {
			if req.BusinessType == bt {
				isValid = true
				break
			}
		}
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "business_type inválido"})
			return
		}

		// 3c. CRECI-J obrigatório APENAS para imobiliária
		// IMPORTANTE: Incorporadoras, construtoras e loteadoras NÃO precisam de CRECI
		// porque vendem seus próprios empreendimentos (não fazem intermediação de terceiros)
		if req.BusinessType == "imobiliaria" {
			if req.TenantCRECI == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI-J é obrigatório para imobiliária"})
				return
			}
			if err := utils.ValidateCRECI(req.TenantCRECI); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI inválido: " + err.Error()})
				return
			}
			if err := utils.ValidateCRECIType(req.TenantCRECI, "J"); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Imobiliária requer CRECI-J (Pessoa Jurídica)"})
				return
			}
		}

		// 3d. CRECI opcional para incorporadoras, construtoras e loteadoras
		// (validar formato se fornecido)
		if req.TenantCRECI != "" && req.BusinessType != "imobiliaria" {
			if err := utils.ValidateCRECI(req.TenantCRECI); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI inválido: " + err.Error()})
				return
			}
		}

		// 3e. Validar CRECI do admin se broker
		if req.IsUserBroker {
			if req.UserCRECI == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI é obrigatório para corretores"})
				return
			}
			if err := utils.ValidateCRECI(req.UserCRECI); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "CRECI do admin inválido: " + err.Error()})
				return
			}
			if err := utils.ValidateCRECIType(req.UserCRECI, "F"); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Admin corretor precisa de CRECI-F individual"})
				return
			}
		}
	}

	// 4. Check if email already exists
	_, err := h.firebaseAuth.GetUserByEmail(ctx, req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// 5. Create Firebase Auth user
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

	// 6. Create Tenant
	tenantID := uuid.New().String()
	slug := generateSlug(req.TenantName)

	// Determinar document_type
	documentType := "cpf"
	if req.TenantType == "pj" {
		documentType = "cnpj"
	}

	now := time.Now()
	tenant := models.Tenant{
		ID:                    tenantID,
		Name:                  req.TenantName,
		Slug:                  slug,
		TenantType:            req.TenantType,
		Document:              req.Document,
		DocumentType:          documentType,
		BusinessType:          req.BusinessType,
		CRECI:                 req.TenantCRECI,
		Email:                 req.Email,
		Phone:                 req.Phone,
		SubscriptionPlan:      "full",
		SubscriptionStatus:    "active",
		SubscriptionStartedAt: &now,
		IsActive:              true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	_, err = h.firestoreDB.Collection("tenants").Doc(tenantID).Set(ctx, tenant)
	if err != nil {
		log.Printf("Error creating tenant: %v", err)
		// Rollback: delete Firebase user
		h.firebaseAuth.DeleteUser(ctx, userRecord.UID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// 7. Create User (unified - no more separate brokers collection)
	// All users go into /tenants/{id}/users collection
	// Role determines if they are broker, admin, or both
	userID := uuid.New().String()

	// Determine role based on whether user is a broker
	var role string
	var creci string
	var permissions []string

	if req.IsUserBroker {
		// User is a broker (has CRECI)
		role = "broker_admin" // First broker is always admin too
		creci = req.UserCRECI
		permissions = []string{
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
		}
		log.Printf("✅ Creating broker admin with CRECI: %s", req.UserCRECI)
	} else {
		// User is admin but not a broker
		role = "admin"
		creci = ""
		permissions = []string{
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
		}
		log.Printf("✅ Creating administrative user (no CRECI)")
	}

	user := models.User{
		ID:          userID,
		TenantID:    tenantID,
		FirebaseUID: userRecord.UID,
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		CRECI:       creci,
		Role:        role,
		IsActive:    true,
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
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

	entityID := userID

	// 8. Set custom claims
	claims := map[string]interface{}{
		"tenant_id": tenantID,
		"role":      role,
		"broker_id": entityID, // For backwards compatibility
		"user_id":   entityID,
	}

	err = h.firebaseAuth.SetCustomUserClaims(ctx, userRecord.UID, claims)
	if err != nil {
		log.Printf("Error setting custom claims: %v", err)
		// Continue anyway, claims can be set later
	}

	// 9. Generate custom token
	token, err := h.firebaseAuth.CustomToken(ctx, userRecord.UID)
	if err != nil {
		log.Printf("Error creating custom token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// 10. Log activity
	go h.logActivity(ctx, tenantID, "tenant_created", map[string]interface{}{
		"tenant_type":   req.TenantType,
		"business_type": req.BusinessType,
		"has_creci":     req.TenantCRECI != "",
	})

	if req.IsUserBroker {
		go h.logActivity(ctx, tenantID, "broker_created", map[string]interface{}{
			"broker_id": entityID,
			"creci":     req.UserCRECI,
		})
	} else {
		go h.logActivity(ctx, tenantID, "user_created", map[string]interface{}{
			"user_id": entityID,
		})
	}

	// 11. Return response
	c.JSON(http.StatusCreated, SignupResponse{
		TenantID:      tenantID,
		BrokerID:      entityID, // For backwards compatibility
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

	// 2. Find user in /users collection (includes both brokers and admin users)
	log.Printf("Looking for user with firebase_uid: %s", userRecord.UID)
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
		log.Printf("❌ User not found for firebase_uid: %s (email: %s)", userRecord.UID, userRecord.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found. Please contact your administrator."})
		return
	}

	// 3. Load user data
	userDoc := userDocs[0]
	var user models.User
	if err := userDoc.DataTo(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user data"})
		return
	}
	user.ID = userDoc.Ref.ID

	tenantID := user.TenantID
	role := user.Role
	entityID := user.ID
	entityName := user.Name
	entityEmail := user.Email
	isActive := user.IsActive

	log.Printf("✅ Found user: %s (tenant: %s, role: %s)", user.ID, user.TenantID, user.Role)

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
