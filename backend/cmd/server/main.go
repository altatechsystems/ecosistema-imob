package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/altatech/ecosistema-imob/backend/internal/config"
	"github.com/altatech/ecosistema-imob/backend/internal/handlers"
	"github.com/altatech/ecosistema-imob/backend/internal/middleware"
	"github.com/altatech/ecosistema-imob/backend/internal/repositories"
	"github.com/altatech/ecosistema-imob/backend/internal/services"
	"github.com/altatech/ecosistema-imob/backend/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize Firebase
	ctx := context.Background()
	firebaseApp, authClient, firestoreClient, err := initializeFirebase(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	defer firestoreClient.Close()
	_ = firebaseApp // keep reference

	log.Println("Firebase initialized successfully")

	// Initialize repositories
	repos := initializeRepositories(firestoreClient)
	log.Println("Repositories initialized")

	// Initialize services
	services := initializeServices(ctx, cfg, repos, firestoreClient)
	log.Println("Services initialized")

	// Initialize handlers
	handlers := initializeHandlers(authClient, firestoreClient, services)
	log.Println("Handlers initialized")

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authClient)
	tenantMiddleware := middleware.NewTenantMiddleware(repos.TenantRepo)

	// Setup router
	router := setupRouter(cfg, handlers, authMiddleware, tenantMiddleware)
	log.Println("Router configured")

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.ServerAddr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// initializeFirebase initializes Firebase Admin SDK
func initializeFirebase(ctx context.Context, cfg *config.Config) (*firebase.App, *auth.Client, *firestore.Client, error) {
	// Initialize Firebase app
	opt := option.WithCredentialsFile(cfg.FirebaseCredentials)
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: cfg.FirebaseProjectID,
	}, opt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	// Initialize Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize Firebase Auth: %w", err)
	}

	// Initialize Firestore client with named database "imob-dev" (where the data is stored)
	log.Println("Connecting to Firestore database: imob-dev")
	firestoreClient, err := firestore.NewClientWithDatabase(ctx, cfg.FirebaseProjectID, "imob-dev", opt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize Firestore: %w", err)
	}
	log.Println("✅ Connected to Firestore database: imob-dev")

	return app, authClient, firestoreClient, nil
}

// Repositories holds all repository instances
type Repositories struct {
	TenantRepo                    *repositories.TenantRepository
	BrokerRepo                    *repositories.BrokerRepository
	UserRepo                      *repositories.UserRepository                      // Administrative users (PROMPT 10)
	OwnerRepo                     *repositories.OwnerRepository
	PropertyRepo                  *repositories.PropertyRepository
	ListingRepo                   *repositories.ListingRepository
	PropertyBrokerRoleRepo        *repositories.PropertyBrokerRoleRepository
	LeadRepo                      *repositories.LeadRepository
	ActivityLogRepo               *repositories.ActivityLogRepository
	OwnerConfirmationTokenRepo    *repositories.OwnerConfirmationTokenRepository    // PROMPT 08
	ScheduledConfirmationRepo     *repositories.ScheduledConfirmationRepository     // Monthly confirmations
}

// initializeRepositories initializes all repositories
func initializeRepositories(client *firestore.Client) *Repositories {
	return &Repositories{
		TenantRepo:                 repositories.NewTenantRepository(client),
		BrokerRepo:                 repositories.NewBrokerRepository(client),
		UserRepo:                   repositories.NewUserRepository(client),                    // PROMPT 10
		OwnerRepo:                  repositories.NewOwnerRepository(client),
		PropertyRepo:               repositories.NewPropertyRepository(client),
		ListingRepo:                repositories.NewListingRepository(client),
		PropertyBrokerRoleRepo:     repositories.NewPropertyBrokerRoleRepository(client),
		LeadRepo:                   repositories.NewLeadRepository(client),
		ActivityLogRepo:            repositories.NewActivityLogRepository(client),
		OwnerConfirmationTokenRepo: repositories.NewOwnerConfirmationTokenRepository(client), // PROMPT 08
		ScheduledConfirmationRepo:  repositories.NewScheduledConfirmationRepository(client),  // Monthly confirmations
	}
}

// Services holds all service instances
type Services struct {
	TenantService                 *services.TenantService
	BrokerService                 *services.BrokerService
	UserService                   *services.UserService                   // Administrative users (PROMPT 10)
	OwnerService                  *services.OwnerService
	PropertyService               *services.PropertyService
	ListingService                *services.ListingService
	PropertyBrokerRoleService     *services.PropertyBrokerRoleService
	LeadService                   *services.LeadService
	ActivityLogService            *services.ActivityLogService
	StorageService                *storage.StorageService
	PhotoProcessor                *services.PhotoProcessor
	ImportService                 *services.ImportService
	OwnerConfirmationService      *services.OwnerConfirmationService      // PROMPT 08
	MonthlyConfirmationScheduler  *services.MonthlyConfirmationScheduler  // Monthly confirmations
}

// initializeServices initializes all services
func initializeServices(ctx context.Context, cfg *config.Config, repos *Repositories, client *firestore.Client) *Services {
	// Initialize Storage service
	storageService, err := storage.NewStorageService(
		ctx,
		cfg.GCSBucketName,
		repos.ActivityLogRepo,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize Storage service: %v", err)
		// Continue without storage service for now
	}

	// Initialize GCSClient for photo processing
	// Use empty credentials file since Firebase already initialized with credentials
	var importService *services.ImportService
	var photoProcessor *services.PhotoProcessor
	gcsClient, err := storage.NewGCSClient(ctx, cfg.FirebaseProjectID, cfg.GCSBucketName, "")
	if err == nil && gcsClient != nil {
		photoProcessor = services.NewPhotoProcessor(gcsClient)
		importService = services.NewImportServiceWithPhotos(client, photoProcessor)
		log.Println("✅ ImportService initialized with photo processing enabled")
	} else {
		if err != nil {
			log.Printf("⚠️  Failed to initialize GCSClient for photos: %v", err)
		}
		importService = services.NewImportService(client)
		log.Println("⚠️  ImportService initialized WITHOUT photo processing")
	}

	// PROMPT 08: Initialize OwnerConfirmationService
	ownerConfirmationService := services.NewOwnerConfirmationService(
		repos.OwnerConfirmationTokenRepo,
		repos.PropertyRepo,
		repos.OwnerRepo,
		repos.BrokerRepo,
		repos.ListingRepo,
		repos.ActivityLogRepo,
	)

	// Initialize MonthlyConfirmationScheduler
	monthlyConfirmationScheduler := services.NewMonthlyConfirmationScheduler(
		repos.ScheduledConfirmationRepo,
		repos.PropertyRepo,
		repos.OwnerRepo,
		ownerConfirmationService,
	)

	// Initialize PropertyService
	propertyService := services.NewPropertyService(
		repos.PropertyRepo,
		repos.ListingRepo,
		repos.OwnerRepo,
		repos.BrokerRepo,
		repos.TenantRepo,
		repos.ActivityLogRepo,
	)
	// PROMPT 08: Inject OwnerConfirmationService into PropertyService
	propertyService.SetOwnerConfirmationService(ownerConfirmationService)

	return &Services{
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
		UserService: services.NewUserService( // PROMPT 10
			repos.UserRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		OwnerService: services.NewOwnerService(
			repos.OwnerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		PropertyService: propertyService, // Use the pre-configured instance
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
		StorageService:              storageService,
		PhotoProcessor:              photoProcessor,
		ImportService:               importService,
		OwnerConfirmationService:    ownerConfirmationService,    // PROMPT 08
		MonthlyConfirmationScheduler: monthlyConfirmationScheduler, // Monthly confirmations
	}
}

// Handlers holds all handler instances
type Handlers struct {
	AuthHandler                  *handlers.AuthHandler
	TenantHandler                *handlers.TenantHandler
	BrokerHandler                *handlers.BrokerHandler
	UserHandler                  *handlers.UserHandler                  // Administrative users (PROMPT 10)
	OwnerHandler                 *handlers.OwnerHandler
	PropertyHandler              *handlers.PropertyHandler
	ListingHandler               *handlers.ListingHandler
	PropertyBrokerRoleHandler    *handlers.PropertyBrokerRoleHandler
	LeadHandler                  *handlers.LeadHandler
	ActivityLogHandler           *handlers.ActivityLogHandler
	StorageHandler               *handlers.StorageHandler
	ImportHandler                *handlers.ImportHandler
	OwnerConfirmationHandler     *handlers.OwnerConfirmationHandler     // PROMPT 08
	ScheduledConfirmationHandler *handlers.ScheduledConfirmationHandler // Monthly confirmations
	// Public handlers (cross-tenant, no tenant_id required)
	PublicPropertyHandler *handlers.PublicPropertyHandler // Portal agregador property endpoints
	PublicLeadHandler     *handlers.PublicLeadHandler     // Portal agregador lead endpoints
	PublicBrokerHandler   *handlers.PublicBrokerHandler   // Portal agregador broker endpoints
}

// initializeHandlers initializes all handlers
func initializeHandlers(authClient *auth.Client, firestoreClient *firestore.Client, services *Services) *Handlers {
	var storageHandler *handlers.StorageHandler
	if services.StorageService != nil {
		storageHandler = handlers.NewStorageHandler(services.StorageService, services.PhotoProcessor)
	}

	return &Handlers{
		AuthHandler:                  handlers.NewAuthHandler(authClient, firestoreClient),
		TenantHandler:                handlers.NewTenantHandler(services.TenantService),
		BrokerHandler:                handlers.NewBrokerHandler(services.BrokerService, services.StorageService),
		UserHandler:                  handlers.NewUserHandler(services.UserService),                               // PROMPT 10
		OwnerHandler:                 handlers.NewOwnerHandler(services.OwnerService),
		PropertyHandler:              handlers.NewPropertyHandler(services.PropertyService),
		ListingHandler:               handlers.NewListingHandler(services.ListingService),
		PropertyBrokerRoleHandler:    handlers.NewPropertyBrokerRoleHandler(services.PropertyBrokerRoleService),
		LeadHandler:                  handlers.NewLeadHandler(services.LeadService),
		ActivityLogHandler:           handlers.NewActivityLogHandler(services.ActivityLogService),
		StorageHandler:               storageHandler,
		ImportHandler:                handlers.NewImportHandler(services.ImportService),
		OwnerConfirmationHandler:     handlers.NewOwnerConfirmationHandler(services.OwnerConfirmationService),     // PROMPT 08
		ScheduledConfirmationHandler: handlers.NewScheduledConfirmationHandler(services.MonthlyConfirmationScheduler), // Monthly confirmations
		// Public handlers (cross-tenant, no tenant_id required)
		PublicPropertyHandler: handlers.NewPublicPropertyHandler(services.PropertyService),
		PublicLeadHandler:     handlers.NewPublicLeadHandler(services.LeadService, services.PropertyService),
		PublicBrokerHandler:   handlers.NewPublicBrokerHandler(services.BrokerService),
	}
}

// setupRouter sets up the Gin router with middleware and routes
func setupRouter(cfg *config.Config, handlers *Handlers, authMiddleware *middleware.AuthMiddleware, tenantMiddleware *middleware.TenantMiddleware) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.ErrorRecovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins:     cfg.AllowedOrigins,
		AllowedMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowedCredentials: true,
		MaxAge:             43200, // 12 hours
	}))

	// Health check endpoint (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  "healthy",
			"service": "ecosistema-imob-api",
		})
	})

	// Metrics endpoint (public)
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"metrics": gin.H{
				"uptime": time.Since(time.Now()).String(),
			},
		})
	})

	// PROMPT 08: Public owner confirmation routes (no authentication required)
	if handlers.OwnerConfirmationHandler != nil {
		handlers.OwnerConfirmationHandler.RegisterPublicRoutes(router)
	}

	// API routes
	api := router.Group("/api/v1")

	// Authentication routes (PUBLIC - no auth required)
	auth := api.Group("/auth")
	{
		auth.POST("/signup", handlers.AuthHandler.Signup)
		auth.POST("/login", handlers.AuthHandler.Login)
		auth.POST("/refresh", authMiddleware.AuthRequired(), handlers.AuthHandler.RefreshToken)
	}

	// Tenant routes (public for creation, auth required for others)
	handlers.TenantHandler.RegisterRoutes(router)

	// Public routes FIRST (no authentication) - frontend público
	// Apply strict rate limiting to public endpoints to prevent abuse
	public := api.Group("/:tenant_id")
	public.Use(middleware.StrictRateLimit())
	{
		// Public property endpoints
		public.GET("/properties", handlers.PropertyHandler.ListProperties)
		public.GET("/properties/:id", handlers.PropertyHandler.GetProperty)
		public.GET("/properties/slug/:slug", handlers.PropertyHandler.GetPropertyBySlug)

		// Public broker endpoints
		public.GET("/brokers/:id/public", handlers.BrokerHandler.GetBrokerPublicProfile)
		public.GET("/brokers/:id/properties", handlers.BrokerHandler.GetBrokerProperties)

		// PROMPT 07: WhatsApp Flow - Public lead creation endpoints
		// Extra strict rate limiting to prevent spam
		public.POST("/leads", handlers.LeadHandler.CreateLead)
		public.POST("/properties/:property_id/leads/whatsapp", handlers.LeadHandler.CreateWhatsAppLead)
		public.POST("/properties/:property_id/leads/form", handlers.LeadHandler.CreateFormLead)

		// Public images
		public.GET("/property-images/:property_id", handlers.StorageHandler.ListImages)
		public.GET("/property-images/:property_id/:image_id", handlers.StorageHandler.GetImageURL)
	}

	// Public routes for portal agregador (NO tenant_id required)
	// List properties/brokers/leads across ALL tenants
	publicPortal := api.Group("/public")
	publicPortal.Use(middleware.StrictRateLimit())
	{
		// Public property endpoints (cross-tenant)
		publicPortal.GET("/properties", handlers.PublicPropertyHandler.ListPublicProperties)
		publicPortal.GET("/properties/:id", handlers.PublicPropertyHandler.GetPublicProperty)
		publicPortal.GET("/properties/slug/:slug", handlers.PublicPropertyHandler.GetPublicPropertyBySlug)

		// Public lead creation endpoints (cross-tenant)
		// Tenant is resolved automatically from property_id
		publicPortal.POST("/properties/:property_id/leads/whatsapp", handlers.PublicLeadHandler.CreatePublicWhatsAppLead)
		publicPortal.POST("/properties/:property_id/leads/form", handlers.PublicLeadHandler.CreatePublicFormLead)

		// Public broker endpoints (cross-tenant)
		publicPortal.GET("/brokers/:id/profile", handlers.PublicBrokerHandler.GetPublicBrokerProfile)
		publicPortal.GET("/brokers/:id/properties", handlers.PublicBrokerHandler.GetPublicBrokerProperties)
	}

	// Protected routes (require authentication) - admin dashboard
	protected := api.Group("/admin")
	protected.Use(authMiddleware.AuthRequired())
	protected.Use(middleware.RateLimit()) // Less strict for authenticated users
	{
		tenantScoped := protected.Group("/:tenant_id")
		tenantScoped.Use(tenantMiddleware.ValidateTenant())
		{
			// Admin-only routes
			handlers.PropertyHandler.RegisterRoutes(tenantScoped)
			handlers.BrokerHandler.RegisterRoutes(tenantScoped)
			handlers.UserHandler.RegisterRoutes(tenantScoped) // PROMPT 10
			handlers.OwnerHandler.RegisterRoutes(tenantScoped)
			handlers.ListingHandler.RegisterRoutes(tenantScoped)
			handlers.PropertyBrokerRoleHandler.RegisterRoutes(tenantScoped)
			handlers.LeadHandler.RegisterRoutes(tenantScoped)
			handlers.ActivityLogHandler.RegisterRoutes(tenantScoped)
			if handlers.StorageHandler != nil {
				handlers.StorageHandler.RegisterRoutes(tenantScoped)
			}

			// Import routes
			if handlers.ImportHandler != nil {
				tenantScoped.POST("/import/properties", handlers.ImportHandler.ImportFromFiles)
				tenantScoped.GET("/import/batches/:batchId", handlers.ImportHandler.GetImportStatus)
				tenantScoped.GET("/import/batches/:batchId/errors", handlers.ImportHandler.GetBatchErrors)
			}

			// Monthly confirmation scheduler routes
			tenantScoped.POST("/scheduled-confirmations/schedule", handlers.ScheduledConfirmationHandler.ScheduleMonthlyConfirmations)
			tenantScoped.POST("/scheduled-confirmations/process", handlers.ScheduledConfirmationHandler.ProcessPendingConfirmations)
			tenantScoped.GET("/scheduled-confirmations/metrics", handlers.ScheduledConfirmationHandler.GetConfirmationMetrics)
			tenantScoped.GET("/scheduled-confirmations/broker/:broker_id", handlers.ScheduledConfirmationHandler.GetBrokerScheduledConfirmations)
			tenantScoped.GET("/scheduled-confirmations", handlers.ScheduledConfirmationHandler.GetScheduledConfirmations)
		}
	}

	return router
}
