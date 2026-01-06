package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"cloud.google.com/go/firestore"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"google.golang.org/api/option"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <email>")
	}

	email := os.Args[1]
	ctx := context.Background()

	projectID := "ecosistema-imob-dev"
	databaseID := "imob-dev"
	tenantID := "bd71c02b-5fa5-43df-8b46-a1df2206f1ef"
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if credentialsPath == "" {
		credentialsPath = "./config/firebase-adminsdk.json"
	}

	// Initialize Firebase Admin SDK
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Failed to create Firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to create Auth client: %v", err)
	}

	// Get user by email
	fmt.Printf("🔍 Looking for Firebase user with email: %s\n", email)
	userRecord, err := authClient.GetUserByEmail(ctx, email)
	if err != nil {
		log.Fatalf("❌ User not found in Firebase Auth: %v\n   Please create the user first in Firebase Console", err)
	}

	fmt.Printf("✅ Found Firebase user: %s (UID: %s)\n", userRecord.Email, userRecord.UID)

	// Initialize Firestore
	firestoreClient, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID, opt)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Check if broker already exists
	fmt.Printf("🔍 Checking if broker already exists...\n")
	brokersQuery := firestoreClient.CollectionGroup("brokers").
		Where("firebase_uid", "==", userRecord.UID).
		Limit(1)

	existingDocs, err := brokersQuery.Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("Error querying brokers: %v", err)
	}

	if len(existingDocs) > 0 {
		fmt.Printf("⚠️  Broker already exists with ID: %s\n", existingDocs[0].Ref.ID)
		fmt.Printf("   Updating to admin role...\n")

		// Update existing broker to admin
		_, err = existingDocs[0].Ref.Update(ctx, []firestore.Update{
			{Path: "role", Value: "admin"},
			{Path: "is_active", Value: true},
			{Path: "updated_at", Value: time.Now()},
		})
		if err != nil {
			log.Fatalf("Failed to update broker: %v", err)
		}
		fmt.Printf("✅ Broker updated successfully!\n")
		return
	}

	// Create new broker
	fmt.Printf("📝 Creating new admin broker...\n")

	broker := models.Broker{
		TenantID:    tenantID,
		FirebaseUID: userRecord.UID,
		Name:        userRecord.DisplayName,
		Email:       userRecord.Email,
		Phone:       userRecord.PhoneNumber,
		Role:        "admin",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// If display name is empty, use email prefix
	if broker.Name == "" {
		broker.Name = "Admin"
	}

	// Add broker to Firestore
	docRef, _, err := firestoreClient.Collection(fmt.Sprintf("tenants/%s/brokers", tenantID)).Add(ctx, broker)
	if err != nil {
		log.Fatalf("❌ Failed to create broker: %v", err)
	}

	fmt.Printf("✅ Admin broker created successfully!\n")
	fmt.Printf("   Broker ID: %s\n", docRef.ID)
	fmt.Printf("   Name: %s\n", broker.Name)
	fmt.Printf("   Email: %s\n", broker.Email)
	fmt.Printf("   Role: %s\n", broker.Role)
	fmt.Printf("   Firebase UID: %s\n", broker.FirebaseUID)
	fmt.Printf("\n🎉 You can now login with this email!\n")
}
