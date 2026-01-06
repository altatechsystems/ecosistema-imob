package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// This utility migrates administrative users (without valid CRECI) from /brokers to /users collection
// Part of PROMPT 10: Separating brokers (real estate agents) from administrative users

func main() {
	ctx := context.Background()

	projectID := "ecosistema-imob-dev"
	databaseID := "imob-dev"

	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		credentialsPath := "./config/firebase-adminsdk.json"
	}

	// Initialize Firestore
	opt := option.WithCredentialsFile(credentialsPath)
	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID, opt)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("PROMPT 10: Migration Tool - Brokers to Users")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Step 1: Find all tenants
	fmt.Println("📋 Step 1: Finding all tenants...")
	tenantsIter := client.Collection("tenants").Documents(ctx)
	tenantDocs, err := tenantsIter.GetAll()
	if err != nil {
		log.Fatalf("Failed to get tenants: %v", err)
	}

	fmt.Printf("✅ Found %d tenants\n\n", len(tenantDocs))

	totalBrokers := 0
	totalMigrated := 0
	totalKept := 0

	// Process each tenant
	for _, tenantDoc := range tenantDocs {
		tenantID := tenantDoc.Ref.ID
		var tenant models.Tenant
		if err := tenantDoc.DataTo(&tenant); err != nil {
			log.Printf("❌ Error loading tenant %s: %v", tenantID, err)
			continue
		}

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("🏢 Tenant: %s (%s)\n", tenant.Name, tenantID)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Step 2: Get all brokers for this tenant
		brokersPath := fmt.Sprintf("tenants/%s/brokers", tenantID)
		brokersIter := client.Collection(brokersPath).Documents(ctx)

		brokerCount := 0
		migratedCount := 0
		keptCount := 0

		for {
			brokerDoc, err := brokersIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("❌ Error iterating brokers: %v", err)
				break
			}

			brokerCount++
			var broker models.Broker
			if err := brokerDoc.DataTo(&broker); err != nil {
				log.Printf("❌ Error loading broker %s: %v", brokerDoc.Ref.ID, err)
				continue
			}
			broker.ID = brokerDoc.Ref.ID

			// Check if broker has valid CRECI
			hasValidCRECI := isValidCRECI(broker.CRECI)

			if hasValidCRECI {
				// Keep in brokers collection
				fmt.Printf("   ✅ KEEP: %s (%s) - CRECI: %s\n", broker.Name, broker.Email, broker.CRECI)
				keptCount++
			} else {
				// Migrate to users collection
				fmt.Printf("   🔄 MIGRATE: %s (%s) - No valid CRECI (value: '%s')\n", broker.Name, broker.Email, broker.CRECI)

				// Create user from broker
				user := &models.User{
					TenantID:    broker.TenantID,
					FirebaseUID: broker.FirebaseUID,
					Name:        broker.Name,
					Email:       broker.Email,
					Phone:       broker.Phone,
					Document:    broker.Document,
					DocumentType: broker.DocumentType,
					Role:        mapBrokerRoleToUserRole(broker.Role),
					IsActive:    broker.IsActive,
					PhotoURL:    broker.PhotoURL,
					CreatedAt:   broker.CreatedAt,
					UpdatedAt:   broker.UpdatedAt,
				}

				// Set default permissions based on role
				if user.Role == "admin" {
					user.Permissions = []string{
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
				} else if user.Role == "manager" {
					user.Permissions = []string{
						"properties.view_all",
						"properties.edit_all",
						"brokers.view",
						"users.view",
						"leads.view_all",
						"leads.edit_all",
					}
				}

				// Add to users collection
				usersPath := fmt.Sprintf("tenants/%s/users", tenantID)
				newUserRef, _, err := client.Collection(usersPath).Add(ctx, user)
				if err != nil {
					log.Printf("      ❌ Failed to create user: %v", err)
					continue
				}

				fmt.Printf("      ✅ Created user with ID: %s\n", newUserRef.ID)

				// Delete from brokers collection
				if err := brokerDoc.Ref.Delete(ctx); err != nil {
					log.Printf("      ⚠️  Warning: Failed to delete broker (user was created): %v", err)
				} else {
					fmt.Printf("      ✅ Deleted from brokers collection\n")
				}

				migratedCount++
			}
		}

		totalBrokers += brokerCount
		totalMigrated += migratedCount
		totalKept += keptCount

		fmt.Println()
		fmt.Printf("   📊 Tenant Summary: %d total, %d migrated to users, %d kept as brokers\n", brokerCount, migratedCount, keptCount)
		fmt.Println()
	}

	// Final Summary
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("MIGRATION COMPLETE")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("📊 Total brokers processed: %d\n", totalBrokers)
	fmt.Printf("🔄 Migrated to /users: %d\n", totalMigrated)
	fmt.Printf("✅ Kept as /brokers: %d\n", totalKept)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Verify the migration by checking Firestore console")
	fmt.Println("2. Test login with both brokers and users")
	fmt.Println("3. Update frontend to use /users endpoint for administrative users")
	fmt.Println("4. Make CRECI mandatory for new broker creations")
	fmt.Println()
}

// isValidCRECI checks if a CRECI string is valid
// Valid format: XXXXX-F/UF or XXXXX-J/UF (e.g., "12345-F/SP", "67890-J/RJ")
// Invalid: empty, "-", "N/A", "PENDENTE", etc.
func isValidCRECI(creci string) bool {
	// Trim whitespace
	creci = strings.TrimSpace(creci)

	// Empty or common placeholder values
	if creci == "" || creci == "-" || creci == "N/A" || creci == "n/a" ||
		strings.ToUpper(creci) == "PENDENTE" || strings.ToUpper(creci) == "PENDING" {
		return false
	}

	// Valid CRECI should have at least 8 characters (e.g., "1234-F/SP")
	if len(creci) < 8 {
		return false
	}

	// Should contain "-" and "/"
	if !strings.Contains(creci, "-") || !strings.Contains(creci, "/") {
		return false
	}

	// Should have F or J (Física or Jurídica)
	upper := strings.ToUpper(creci)
	if !strings.Contains(upper, "-F/") && !strings.Contains(upper, "-J/") {
		return false
	}

	return true
}

// mapBrokerRoleToUserRole maps broker roles to user roles
// broker_admin -> admin
// admin -> admin
// broker -> manager (if they have no CRECI, they're probably a manager)
// manager -> manager
func mapBrokerRoleToUserRole(brokerRole string) string {
	switch brokerRole {
	case "broker_admin", "admin":
		return "admin"
	case "manager":
		return "manager"
	case "broker":
		// If a "broker" has no CRECI, they're probably a manager
		return "manager"
	default:
		// Default to admin for safety
		return "admin"
	}
}
