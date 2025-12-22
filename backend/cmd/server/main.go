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
	services := initializeServices(repos)
	log.Println("Services initialized")

	// Initialize handlers
	handlers := initializeHandlers(services)
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

	// Initialize Firestore client
	firestoreClient, err := app.Firestore(ctx)
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
}

// initializeServices initializes all services
func initializeServices(repos *Repositories) *Services {
	return &Services{
		TenantService: services.NewTenantService(
			repos.TenantRepo,
			repos.ActivityLogRepo,
		),
		BrokerService: services.NewBrokerService(
			repos.BrokerRepo,
			repos.TenantRepo,
			repos.ActivityLogRepo,
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
}

// Handlers holds all handler instances
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

// initializeHandlers initializes all handlers
func initializeHandlers(services *Services) *Handlers {
	return &Handlers{
		TenantHandler:             handlers.NewTenantHandler(services.TenantService),
		BrokerHandler:             handlers.NewBrokerHandler(services.BrokerService),
		OwnerHandler:              handlers.NewOwnerHandler(services.OwnerService),
		PropertyHandler:           handlers.NewPropertyHandler(services.PropertyService),
		ListingHandler:            handlers.NewListingHandler(services.ListingService),
		PropertyBrokerRoleHandler: handlers.NewPropertyBrokerRoleHandler(services.PropertyBrokerRoleService),
		LeadHandler:               handlers.NewLeadHandler(services.LeadService),
		ActivityLogHandler:        handlers.NewActivityLogHandler(services.ActivityLogService),
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
	api := router.Group("/api")

	// Tenant routes (public for creation, auth required for others)
	handlers.TenantHandler.RegisterRoutes(router)

	// Protected API routes (require authentication)
	protected := api.Group("")
	protected.Use(authMiddleware.AuthRequired())
	{
		// Routes that require tenant validation
		tenantScoped := protected.Group("")
		tenantScoped.Use(tenantMiddleware.ValidateTenant())
		{
			// Register tenant-scoped routes
			handlers.BrokerHandler.RegisterRoutes(tenantScoped)
			handlers.OwnerHandler.RegisterRoutes(tenantScoped)
			handlers.PropertyHandler.RegisterRoutes(tenantScoped)
			handlers.ListingHandler.RegisterRoutes(tenantScoped)
			handlers.PropertyBrokerRoleHandler.RegisterRoutes(tenantScoped)
			handlers.LeadHandler.RegisterRoutes(tenantScoped)
			handlers.ActivityLogHandler.RegisterRoutes(tenantScoped)
		}
	}

	// Public listings endpoint (optional auth for personalized results)
	public := api.Group("/public")
	public.Use(authMiddleware.OptionalAuth())
	{
		// Public property listings can be added here
		// Example: public.GET("/listings", handlers.ListingHandler.PublicListings)
	}

	return router
}
