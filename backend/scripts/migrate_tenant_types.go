package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Get Firebase credentials path from environment
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credPath == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS environment variable not set")
	}

	// Initialize Firebase App
	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	// Get Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error getting Firestore client: %v", err)
	}
	defer client.Close()

	log.Println("🚀 Starting tenant type migration...")
	log.Println("⚠️  This will update all existing tenants with tenant_type and subscription fields")

	// Fetch all tenants
	tenants, err := client.Collection("tenants").Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("Error fetching tenants: %v", err)
	}

	log.Printf("📊 Found %d tenants to migrate", len(tenants))

	successCount := 0
	errorCount := 0

	for i, tenantDoc := range tenants {
		tenantID := tenantDoc.Ref.ID

		var tenantData map[string]interface{}
		if err := tenantDoc.DataTo(&tenantData); err != nil {
			log.Printf("❌ Error reading tenant %s: %v", tenantID, err)
			errorCount++
			continue
		}

		log.Printf("\n[%d/%d] Processing tenant: %s", i+1, len(tenants), tenantID)

		// Check if already migrated
		if _, exists := tenantData["tenant_type"]; exists {
			log.Printf("   ⏭️  Already migrated, skipping...")
			continue
		}

		// Infer tenant_type from document_type
		tenantType := "pj" // Default to PJ
		documentType, _ := tenantData["document_type"].(string)
		if documentType == "cpf" {
			tenantType = "pf"
		}

		// Infer business_type if not exists
		businessType, _ := tenantData["business_type"].(string)
		if businessType == "" {
			if tenantType == "pf" {
				businessType = "corretor_autonomo"
			} else {
				// Default PJ to imobiliaria (safest assumption)
				businessType = "imobiliaria"
			}
		}

		// Prepare updates
		now := time.Now()
		updates := []firestore.Update{
			{Path: "tenant_type", Value: tenantType},
			{Path: "business_type", Value: businessType},
			{Path: "subscription_plan", Value: "full"},
			{Path: "subscription_status", Value: "active"},
			{Path: "subscription_started_at", Value: now},
		}

		// Apply updates
		_, err := tenantDoc.Ref.Update(ctx, updates)
		if err != nil {
			log.Printf("   ❌ Error updating tenant %s: %v", tenantID, err)
			errorCount++
			continue
		}

		log.Printf("   ✅ Migrated successfully")
		log.Printf("      - Tenant Type: %s", tenantType)
		log.Printf("      - Business Type: %s", businessType)
		log.Printf("      - Subscription: full (active)")

		successCount++
	}

	log.Println("\n" + "=".repeat(60))
	log.Println("📊 MIGRATION SUMMARY")
	log.Println("=".repeat(60))
	log.Printf("Total tenants: %d", len(tenants))
	log.Printf("✅ Successfully migrated: %d", successCount)
	log.Printf("❌ Errors: %d", errorCount)
	log.Println("=".repeat(60))

	if errorCount > 0 {
		log.Println("⚠️  Some tenants failed to migrate. Check the logs above for details.")
		os.Exit(1)
	}

	log.Println("🎉 Migration completed successfully!")
}

// Helper function to repeat strings (Go doesn't have strings.Repeat for single char)
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
