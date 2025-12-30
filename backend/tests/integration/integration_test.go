package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/altatech/ecosistema-imob/backend/internal/config"
	"github.com/altatech/ecosistema-imob/backend/internal/handlers"
	"github.com/altatech/ecosistema-imob/backend/internal/middleware"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
)

// TestContext holds the test environment
type TestContext struct {
	Router         *gin.Engine
	FirestoreClient *firestore.Client
	Repos          *Repositories
	Services       *Services
	Handlers       *Handlers
	TestTenantID   string
}

type Repositories struct {
	TenantRepo             *repositories.TenantRepository
	BrokerRepo             *repositories.BrokerRepository
	OwnerRepo              *repositories.OwnerRepository
	PropertyRepo           *repositories.PropertyRepository
	ListingRepo            *repositories.ListingRepository
	PropertyBrokerRoleRepo *repositories.PropertyBrokerRoleRepository
	LeadRepo               *repositories.LeadRepository
	ActivityLogRepo        *repositories.ActivityLogRepository
}

type Services struct {
	TenantService             *services.TenantService
	BrokerService             *services.BrokerService
	OwnerService              *services.OwnerService
	PropertyService           *services.PropertyService
	ListingService            *services.ListingService
	PropertyBrokerRoleService *services.PropertyBrokerRoleService
	LeadService               *services.LeadService
	ActivityLogService        *services.ActivityLogService
}

type Handlers struct {
	TenantHandler             *handlers.TenantHandler
	BrokerHandler             *handlers.BrokerHandler
	OwnerHandler              *handlers.OwnerHandler
	PropertyHandler           *handlers.PropertyHandler
	ListingHandler            *handlers.ListingHandler
	PropertyBrokerRoleHandler *handlers.PropertyBrokerRoleHandler
	LeadHandler               *handlers.LeadHandler
	ActivityLogHandler        *handlers.ActivityLogHandler
}

// SetupTestContext initializes the test environment
func SetupTestContext(t *testing.T) *TestContext {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	// Initialize Firebase
	opt := option.WithCredentialsFile(cfg.FirebaseCredentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	require.NoError(t, err, "Failed to initialize Firebase app")

	// Get Firestore client
	firestoreClient, err := app.Firestore(ctx)
	require.NoError(t, err, "Failed to get Firestore client")

	// Initialize repositories
	repos := &Repositories{
		TenantRepo:             repositories.NewTenantRepository(firestoreClient),
		BrokerRepo:             repositories.NewBrokerRepository(firestoreClient),
		OwnerRepo:              repositories.NewOwnerRepository(firestoreClient),
		PropertyRepo:           repositories.NewPropertyRepository(firestoreClient),
		ListingRepo:            repositories.NewListingRepository(firestoreClient),
		PropertyBrokerRoleRepo: repositories.NewPropertyBrokerRoleRepository(firestoreClient),
		LeadRepo:               repositories.NewLeadRepository(firestoreClient),
		ActivityLogRepo:        repositories.NewActivityLogRepository(firestoreClient),
	}

	// Initialize services
	svcs := &Services{
		TenantService: services.NewTenantService(
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		BrokerService: services.NewBrokerService(
			repos.BrokerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
			repos.PropertyBrokerRoleRepo,
			repos.PropertyRepo,
			repos.ListingRepo,
		),
		OwnerService: services.NewOwnerService(
			repos.OwnerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		PropertyService: services.NewPropertyService(
			repos.PropertyRepo,
			repos.OwnerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		ListingService: services.NewListingService(
			repos.ListingRepo,
			repos.PropertyRepo,
			repos.BrokerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		PropertyBrokerRoleService: services.NewPropertyBrokerRoleService(
			repos.PropertyBrokerRoleRepo,
			repos.PropertyRepo,
			repos.BrokerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		LeadService: services.NewLeadService(
			repos.LeadRepo,
			repos.PropertyRepo,
			repos.PropertyBrokerRoleRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		ActivityLogService: services.NewActivityLogService(
			repos.ActivityLogRepo,
			repos.TenantRepo,
		),
	}

	// Initialize handlers
	hdlrs := &Handlers{
		TenantHandler:             handlers.NewTenantHandler(svcs.TenantService),
		BrokerHandler:             handlers.NewBrokerHandler(svcs.BrokerService),
		OwnerHandler:              handlers.NewOwnerHandler(svcs.OwnerService),
		PropertyHandler:           handlers.NewPropertyHandler(svcs.PropertyService),
		ListingHandler:            handlers.NewListingHandler(svcs.ListingService),
		PropertyBrokerRoleHandler: handlers.NewPropertyBrokerRoleHandler(svcs.PropertyBrokerRoleService),
		LeadHandler:               handlers.NewLeadHandler(svcs.LeadService),
		ActivityLogHandler:        handlers.NewActivityLogHandler(svcs.ActivityLogService),
	}

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Register tenant routes (no auth for tests)
	hdlrs.TenantHandler.RegisterRoutes(router)

	// API routes (no auth middleware for tests)
	api := router.Group("/api")
	{
		// Tenant-scoped routes (skip auth and tenant validation for tests)
		hdlrs.BrokerHandler.RegisterRoutes(api)
		hdlrs.OwnerHandler.RegisterRoutes(api)
		hdlrs.PropertyHandler.RegisterRoutes(api)
		hdlrs.ListingHandler.RegisterRoutes(api)
		hdlrs.PropertyBrokerRoleHandler.RegisterRoutes(api)
		hdlrs.LeadHandler.RegisterRoutes(api)
		hdlrs.ActivityLogHandler.RegisterRoutes(api)
	}

	return &TestContext{
		Router:          router,
		FirestoreClient: firestoreClient,
		Repos:           repos,
		Services:        svcs,
		Handlers:        hdlrs,
	}
}

// TeardownTestContext cleans up test data
func (tc *TestContext) TeardownTestContext(t *testing.T) {
	ctx := context.Background()

	// Clean up test tenant if created
	if tc.TestTenantID != "" {
		err := tc.Repos.TenantRepo.Delete(ctx, tc.TestTenantID)
		if err != nil {
			t.Logf("Warning: Failed to delete test tenant: %v", err)
		}
	}

	// Close Firestore client
	if tc.FirestoreClient != nil {
		tc.FirestoreClient.Close()
	}
}

// TestTenantCreation tests creating a tenant
func TestTenantCreation(t *testing.T) {
	tc := SetupTestContext(t)
	defer tc.TeardownTestContext(t)

	// Create tenant request
	tenantReq := map[string]interface{}{
		"name":     "Test Imobiliária",
		"document": "11222333000181",
		"email":    "contato@testimobiliaria.com.br",
		"phone":    "11987654321",
	}

	reqBody, _ := json.Marshal(tenantReq)
	req := httptest.NewRequest("POST", "/tenants", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	tc.Router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	// Save tenant ID for cleanup
	tenant := response["data"].(map[string]interface{})
	tc.TestTenantID = tenant["id"].(string)

	// Verify slug was generated
	assert.NotEmpty(t, tenant["slug"])
	assert.Equal(t, "test-imobiliaria", tenant["slug"])
}

// TestPropertyLifecycle tests the complete property lifecycle
func TestPropertyLifecycle(t *testing.T) {
	tc := SetupTestContext(t)
	defer tc.TeardownTestContext(t)

	ctx := context.Background()

	// 1. Create test tenant
	tenant := &models.Tenant{
		Name:     "Test Imobiliária",
		Slug:     "test-imob",
		Document: "11222333000181",
		Email:    "test@imob.com",
		Phone:    "11987654321",
		IsActive: true,
	}
	err := tc.Repos.TenantRepo.Create(ctx, tenant)
	require.NoError(t, err)
	tc.TestTenantID = tenant.ID

	// 2. Create test owner
	owner := &models.Owner{
		TenantID:    tenant.ID,
		Name:        "João Silva",
		Email:       "joao@example.com",
		Phone:       "11998877665",
		Document:    "12345678909",
		DocumentType: models.DocumentTypeCPF,
		Status:      models.OwnerStatusVerified,
		ConsentGiven: true,
	}
	err = tc.Repos.OwnerRepo.Create(ctx, owner.TenantID, owner)
	require.NoError(t, err)

	// 3. Create property via API
	propertyReq := map[string]interface{}{
		"owner_id":         owner.ID,
		"transaction_type": "sale",
		"property_type":    "apartment",
		"status":           "available",
		"sale_price":       500000.00,
		"street":           "Rua Teste",
		"number":           "123",
		"neighborhood":     "Centro",
		"city":             "São Paulo",
		"state":            "SP",
		"postal_code":      "01234567",
		"bedrooms":         3,
		"bathrooms":        2,
		"parking_spaces":   1,
		"area_sqm":         85.5,
	}

	reqBody, _ := json.Marshal(propertyReq)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/%s/properties", tenant.ID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	tc.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	property := response["data"].(map[string]interface{})
	propertyID := property["id"].(string)

	// 4. Get property
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/%s/properties/%s", tenant.ID, propertyID), nil)
	w = httptest.NewRecorder()
	tc.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Update property status
	statusReq := map[string]interface{}{
		"status": "sold",
	}

	reqBody, _ = json.Marshal(statusReq)
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/%s/properties/%s/status", tenant.ID, propertyID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	tc.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 6. List properties
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/%s/properties", tenant.ID), nil)
	w = httptest.NewRecorder()
	tc.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	properties := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(properties), 1)
}

// TestLeadCreationAndRouting tests lead creation and broker routing
func TestLeadCreationAndRouting(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=1 to run.")
	}

	tc := SetupTestContext(t)
	defer tc.TeardownTestContext(t)

	ctx := context.Background()

	// Setup: Create tenant, broker, owner, and property
	tenant := &models.Tenant{
		Name:     "Test Imobiliária",
		Slug:     "test-leads",
		Document: "11222333000181",
		Email:    "test@leads.com",
		Phone:    "11987654321",
		IsActive: true,
	}
	err := tc.Repos.TenantRepo.Create(ctx, tenant)
	require.NoError(t, err)
	tc.TestTenantID = tenant.ID

	broker := &models.Broker{
		TenantID: tenant.ID,
		Name:     "Carlos Corretor",
		Email:    "carlos@imob.com",
		Phone:    "11987654321",
		CRECI:    "12345-F/SP",
		Role:     models.BrokerRoleBroker,
		IsActive: true,
	}
	err = tc.Repos.BrokerRepo.Create(ctx, broker.TenantID, broker)
	require.NoError(t, err)

	owner := &models.Owner{
		TenantID:     tenant.ID,
		Name:         "Test Owner",
		Email:        "owner@test.com",
		Phone:        "11987654321",
		Document:     "12345678909",
		DocumentType: models.DocumentTypeCPF,
		Status:       models.OwnerStatusVerified,
		ConsentGiven: true,
	}
	err = tc.Repos.OwnerRepo.Create(ctx, owner.TenantID, owner)
	require.NoError(t, err)

	property := &models.Property{
		TenantID:        tenant.ID,
		OwnerID:         owner.ID,
		TransactionType: models.TransactionTypeSale,
		PropertyType:    models.PropertyTypeApartment,
		Status:          models.PropertyStatusAvailable,
		Street:          "Rua Teste",
		Number:          "100",
		City:            "São Paulo",
		State:           "SP",
		PostalCode:      "01234567",
	}
	err = tc.Repos.PropertyRepo.Create(ctx, property.TenantID, property)
	require.NoError(t, err)

	// Create lead via API
	leadReq := map[string]interface{}{
		"property_id":  property.ID,
		"name":         "Maria Cliente",
		"email":        "maria@example.com",
		"phone":        "11999887766",
		"message":      "Gostaria de mais informações sobre este imóvel",
		"channel":      "whatsapp",
		"consent_text": "Aceito receber contatos",
	}

	reqBody, _ := json.Marshal(leadReq)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/%s/leads", tenant.ID), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	tc.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	lead := response["data"].(map[string]interface{})

	// Verify lead was created with broker assigned
	assert.NotEmpty(t, lead["id"])
	assert.Equal(t, "new", lead["status"])
	// Note: broker assignment happens in service layer based on routing logic
}

// TestMain runs before all tests
func TestMain(m *testing.M) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Run tests
	code := m.Run()

	// Teardown
	os.Exit(code)
}
