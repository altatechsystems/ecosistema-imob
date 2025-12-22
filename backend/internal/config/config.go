package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Firebase configuration
	FirebaseProjectID   string
	FirebaseCredentials string
	GCSBucketName       string

	// Server configuration
	Port        string
	Host        string
	Environment string
	GinMode     string

	// CORS configuration
	AllowedOrigins []string

	// Logging configuration
	LogLevel string
}

// Load loads configuration from environment variables
// It attempts to load .env file first, then reads from environment
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		// Firebase
		FirebaseProjectID:   getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseCredentials: getEnv("GOOGLE_APPLICATION_CREDENTIALS", "./config/firebase-adminsdk.json"),
		GCSBucketName:       getEnv("GCS_BUCKET_NAME", ""),

		// Server
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		GinMode:     getEnv("GIN_MODE", "debug"),

		// CORS
		AllowedOrigins: parseCSV(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:3002")),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.FirebaseProjectID == "" {
		return fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}

	if c.FirebaseCredentials == "" {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS is required")
	}

	// Check if credentials file exists
	if _, err := os.Stat(c.FirebaseCredentials); os.IsNotExist(err) {
		return fmt.Errorf("Firebase credentials file not found at: %s", c.FirebaseCredentials)
	}

	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// ServerAddr returns the server address in host:port format
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// parseCSV parses a comma-separated string into a slice
func parseCSV(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
