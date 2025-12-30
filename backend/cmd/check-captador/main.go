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

	// Firebase credentials
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		credentialsPath = "C:\\Users\\danie\\OneDrive\\Documentos\\Altatech Systems\\ecosystem\\ecosistema-imob\\backend\\config\\firebase-adminsdk.json"
	}

	projectID := "ecosistema-imob-dev"
	databaseID := "imob-dev"

	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	fmt.Println("🔍 Checking properties for captador_name field...")

	// Query property CA00301 specifically
	propertyID := "06afc07c-9872-4864-b32f-12b31b9759a2" // CA00301
	doc, err := client.Collection("properties").Doc(propertyID).Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get property: %v", err)
	}

	data := doc.Data()
	fmt.Printf("Property ID: %s\n", doc.Ref.ID)
	fmt.Printf("Reference: %v\n", data["reference"])
	fmt.Printf("Captador Name: %v\n", data["captador_name"])
	fmt.Printf("Captador ID: %v\n\n", data["captador_id"])

	fmt.Println("Checking all fields:")
	for key, value := range data {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Also query first 10 properties
	fmt.Println("\n\n📊 Checking first 10 properties:")
	docs, err := client.Collection("properties").Limit(10).Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("Failed to get properties: %v", err)
	}

	fmt.Printf("📊 Found %d properties\n\n", len(docs))

	for i, doc := range docs {
		data := doc.Data()
		reference := data["reference"]
		captadorName := data["captador_name"]
		captadorID := data["captador_id"]

		fmt.Printf("%d. Property ID: %s\n", i+1, doc.Ref.ID)
		fmt.Printf("   Reference: %v\n", reference)
		fmt.Printf("   Captador Name: %v\n", captadorName)
		fmt.Printf("   Captador ID: %v\n", captadorID)
		fmt.Println()
	}
}
