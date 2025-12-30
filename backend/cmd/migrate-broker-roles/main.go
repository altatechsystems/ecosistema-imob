package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
)

func main() {
	ctx := context.Background()

	// Get Firebase credentials from environment
	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT")
	if serviceAccountPath == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT environment variable is required")
	}

	// Initialize Firebase
	opt := option.WithCredentialsFile(serviceAccountPath)
	_, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Get Firestore client with named database
	firestoreClient, err := firestore.NewClientWithDatabase(ctx, "ecosistema-imob-dev", "imob-dev", opt)
	if err != nil {
		log.Fatalf("Failed to get Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	log.Println("Starting migration: Create PropertyBrokerRole for properties with CaptadorID")

	// Get all tenants
	tenants, err := getTenants(ctx, firestoreClient)
	if err != nil {
		log.Fatalf("Failed to get tenants: %v", err)
	}

	log.Printf("Found %d tenant(s)\n", len(tenants))

	totalCreated := 0
	totalSkipped := 0

	for _, tenantID := range tenants {
		log.Printf("\nProcessing tenant: %s", tenantID)

		// Get all properties for this tenant
		properties, err := getProperties(ctx, firestoreClient, tenantID)
		if err != nil {
			log.Printf("Error getting properties for tenant %s: %v", tenantID, err)
			continue
		}

		log.Printf("Found %d properties for tenant %s", len(properties), tenantID)

		for _, property := range properties {
			// Skip if no CaptadorID
			if property.CaptadorID == "" {
				totalSkipped++
				continue
			}

			// Check if PropertyBrokerRole already exists
			exists, err := propertyBrokerRoleExists(ctx, firestoreClient, tenantID, property.ID, property.CaptadorID)
			if err != nil {
				log.Printf("Error checking PropertyBrokerRole for property %s: %v", property.ID, err)
				continue
			}

			if exists {
				log.Printf("PropertyBrokerRole already exists for property %s (broker %s), skipping", property.ID, property.CaptadorID)
				totalSkipped++
				continue
			}

			// Create PropertyBrokerRole
			if err := createPropertyBrokerRole(ctx, firestoreClient, tenantID, property.ID, property.CaptadorID); err != nil {
				log.Printf("Error creating PropertyBrokerRole for property %s: %v", property.ID, err)
				continue
			}

			log.Printf("✓ Created PropertyBrokerRole for property %s (broker %s)", property.ID, property.CaptadorID)
			totalCreated++
		}
	}

	log.Printf("\n=== Migration Complete ===")
	log.Printf("Total PropertyBrokerRole created: %d", totalCreated)
	log.Printf("Total skipped: %d", totalSkipped)
}

func getTenants(ctx context.Context, db *firestore.Client) ([]string, error) {
	// Get all documents from /tenants collection
	iter := db.Collection("tenants").Documents(ctx)
	defer iter.Stop()

	var tenantIDs []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate tenants: %w", err)
		}

		tenantIDs = append(tenantIDs, doc.Ref.ID)
	}

	return tenantIDs, nil
}

func getProperties(ctx context.Context, db *firestore.Client, tenantID string) ([]*models.Property, error) {
	// Properties are stored in root-level /properties collection with tenant_id field
	iter := db.Collection("properties").Where("tenant_id", "==", tenantID).Documents(ctx)
	defer iter.Stop()

	var properties []*models.Property
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate properties: %w", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			log.Printf("Warning: failed to parse property %s: %v", doc.Ref.ID, err)
			continue
		}

		property.ID = doc.Ref.ID
		properties = append(properties, &property)
	}

	return properties, nil
}

func propertyBrokerRoleExists(ctx context.Context, db *firestore.Client, tenantID, propertyID, brokerID string) (bool, error) {
	collectionPath := fmt.Sprintf("tenants/%s/property_broker_roles", tenantID)

	// Query for existing role
	iter := db.Collection(collectionPath).
		Where("property_id", "==", propertyID).
		Where("broker_id", "==", brokerID).
		Where("role", "==", string(models.BrokerPropertyRoleOriginating)).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func createPropertyBrokerRole(ctx context.Context, db *firestore.Client, tenantID, propertyID, brokerID string) error {
	now := time.Now()
	roleID := uuid.New().String()

	role := models.PropertyBrokerRole{
		ID:                   roleID,
		TenantID:             tenantID,
		PropertyID:           propertyID,
		BrokerID:             brokerID,
		Role:                 models.BrokerPropertyRoleOriginating, // CAPTADOR
		CommissionPercentage: 0,                                    // Será definido depois
		IsPrimary:            true,                                 // Primeiro corretor é primary
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	collectionPath := fmt.Sprintf("tenants/%s/property_broker_roles", tenantID)

	// Create document with specific ID
	_, err := db.Collection(collectionPath).Doc(roleID).Set(ctx, map[string]interface{}{
		"tenant_id":             role.TenantID,
		"property_id":           role.PropertyID,
		"broker_id":             role.BrokerID,
		"role":                  string(role.Role),
		"commission_percentage": role.CommissionPercentage,
		"is_primary":            role.IsPrimary,
		"created_at":            role.CreatedAt,
		"updated_at":            role.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create PropertyBrokerRole: %w", err)
	}

	return nil
}
