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
	// Set environment variable for Firestore database
	// This tells Firestore SDK to use the named database instead of (default)
	os.Setenv("FIRESTORE_DATABASE", "imob-dev")

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

	// Initialize Firestore client directly with named database
	// Use firestore.NewClient with DatabaseID to connect to non-default database
	firestoreClient, err := firestore.NewClientWithDatabase(ctx, cfg.FirebaseProjectID, "imob-dev", opt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize Firestore: %w", err)
	}

	return app, authClient, firestoreClient, nil
}

// Repositories holds all repository instances
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

// initializeRepositories initializes all repositories
func initializeRepositories(client *firestore.Client) *Repositories {
	return &Repositories{
		TenantRepo:             repositories.NewTenantRepository(client),
		BrokerRepo:             repositories.NewBrokerRepository(client),
		OwnerRepo:              repositories.NewOwnerRepository(client),
		PropertyRepo:           repositories.NewPropertyRepository(client),
		ListingRepo:            repositories.NewListingRepository(client),
		PropertyBrokerRoleRepo: repositories.NewPropertyBrokerRoleRepository(client),
		LeadRepo:               repositories.NewLeadRepository(client),
		ActivityLogRepo:        repositories.NewActivityLogRepository(client),
	}
}

// Services holds all service instances
type Services struct {
	TenantService             *services.TenantService
	BrokerService             *services.BrokerService
	OwnerService              *services.OwnerService
	PropertyService           *services.PropertyService
	ListingService            *services.ListingService
	PropertyBrokerRoleService *services.PropertyBrokerRoleService
	LeadService               *services.LeadService
	ActivityLogService        *services.ActivityLogService
	StorageService            *storage.StorageService
	PhotoProcessor            *services.PhotoProcessor
	ImportService             *services.ImportService
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
		),
		OwnerService: services.NewOwnerService(
			repos.OwnerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		PropertyService: services.NewPropertyService(
			repos.PropertyRepo,
			repos.ListingRepo,
			repos.OwnerRepo,
			repos.BrokerRepo,
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
		StorageService: storageService,
		PhotoProcessor: photoProcessor,
		ImportService:  importService,
	}
}

// Handlers holds all handler instances
type Handlers struct {
	AuthHandler               *handlers.AuthHandler
	TenantHandler             *handlers.TenantHandler
	BrokerHandler             *handlers.BrokerHandler
	OwnerHandler              *handlers.OwnerHandler
	PropertyHandler           *handlers.PropertyHandler
	ListingHandler            *handlers.ListingHandler
	PropertyBrokerRoleHandler *handlers.PropertyBrokerRoleHandler
	LeadHandler               *handlers.LeadHandler
	ActivityLogHandler        *handlers.ActivityLogHandler
	StorageHandler            *handlers.StorageHandler
	ImportHandler             *handlers.ImportHandler
}

// initializeHandlers initializes all handlers
func initializeHandlers(authClient *auth.Client, firestoreClient *firestore.Client, services *Services) *Handlers {
	var storageHandler *handlers.StorageHandler
	if services.StorageService != nil {
		storageHandler = handlers.NewStorageHandler(services.StorageService, services.PhotoProcessor)
	}

	return &Handlers{
		AuthHandler:               handlers.NewAuthHandler(authClient, firestoreClient),
		TenantHandler:             handlers.NewTenantHandler(services.TenantService),
		BrokerHandler:             handlers.NewBrokerHandler(services.BrokerService, services.StorageService),
		OwnerHandler:              handlers.NewOwnerHandler(services.OwnerService),
		PropertyHandler:           handlers.NewPropertyHandler(services.PropertyService),
		ListingHandler:            handlers.NewListingHandler(services.ListingService),
		PropertyBrokerRoleHandler: handlers.NewPropertyBrokerRoleHandler(services.PropertyBrokerRoleService),
		LeadHandler:               handlers.NewLeadHandler(services.LeadService),
		ActivityLogHandler:        handlers.NewActivityLogHandler(services.ActivityLogService),
		StorageHandler:            storageHandler,
		ImportHandler:             handlers.NewImportHandler(services.ImportService),
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

		// Public lead creation (extra strict - prevent spam)
		public.POST("/leads", handlers.LeadHandler.CreateLead)

		// Public images
		public.GET("/property-images/:property_id", handlers.StorageHandler.ListImages)
		public.GET("/property-images/:property_id/:image_id", handlers.StorageHandler.GetImageURL)
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
			handlers.OwnerHandler.RegisterRoutes(tenantScoped)
			handlers.ListingHandler.RegisterRoutes(tenantScoped)
			handlers.PropertyBrokerRoleHandler.RegisterRoutes(tenantScoped)
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
		}
	}

	return router
}
