package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	projectID := "ecosistema-imob-dev"
	databaseID := "imob-dev"
	tenantID := "bd71c02b-5fa5-43df-8b46-a1df2206f1ef"
	brokerID := "VbAGkmF0XQQmrup71IFu" // The broker ID from the previous creation

	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		credentialsPath = "./config/firebase-adminsdk.json"
	}

	// Initialize Firestore
	opt := option.WithCredentialsFile(credentialsPath)
	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID, opt)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	// Update broker with tenant_id
	fmt.Printf("🔧 Updating broker %s with tenant_id...\n", brokerID)

	brokerRef := client.Doc(fmt.Sprintf("tenants/%s/brokers/%s", tenantID, brokerID))

	_, err = brokerRef.Update(ctx, []firestore.Update{
		{Path: "tenant_id", Value: tenantID},
	})
	if err != nil {
		log.Fatalf("❌ Failed to update broker: %v", err)
	}

	fmt.Printf("✅ Broker updated successfully with tenant_id: %s\n", tenantID)
	fmt.Println("🎉 You can now login!")
}
