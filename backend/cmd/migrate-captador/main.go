package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/altatech/ecosistema-imob/backend/internal/adapters/union"
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

	// Path to XLS file
	xlsPath := "C:\\Users\\danie\\OneDrive\\Documentos\\Altatech Systems\\ecosystem\\ecosistema-imob\\univen-imoveis_20-12-2025.xlsx"
	if len(os.Args) > 1 {
		xlsPath = os.Args[1]
	}

	fmt.Printf("📂 Reading XLS file: %s\n", xlsPath)

	// Parse XLS
	records, err := union.ParseXLS(xlsPath)
	if err != nil {
		log.Fatalf("Failed to parse XLS: %v", err)
	}

	fmt.Printf("✅ Parsed %d records from XLS\n\n", len(records))

	// Create map of reference -> captador
	captadorMap := make(map[string]string)
	for _, rec := range records {
		if rec.Referencia != "" && rec.Captador != "" {
			captadorMap[rec.Referencia] = rec.Captador
		}
	}

	fmt.Printf("📊 Found %d properties with captador in XLS\n\n", len(captadorMap))

	// Get all properties from Firestore
	docs, err := client.Collection("properties").Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("Failed to get properties: %v", err)
	}

	fmt.Printf("🔍 Found %d properties in Firestore\n\n", len(docs))

	updated := 0
	skipped := 0

	for _, doc := range docs {
		data := doc.Data()
		reference, ok := data["reference"].(string)
		if !ok || reference == "" {
			skipped++
			continue
		}

		captadorName, hasCaptador := captadorMap[reference]
		if !hasCaptador {
			skipped++
			continue
		}

		// Check if already has captador_name
		existingCaptador, _ := data["captador_name"].(string)
		if existingCaptador == captadorName {
			fmt.Printf("⏭️  %s: Already has captador '%s'\n", reference, captadorName)
			skipped++
			continue
		}

		// Update property with captador_name
		_, err := doc.Ref.Update(ctx, []firestore.Update{
			{Path: "captador_name", Value: captadorName},
		})
		if err != nil {
			log.Printf("❌ Failed to update %s: %v\n", reference, err)
			continue
		}

		fmt.Printf("✅ %s: Added captador '%s'\n", reference, captadorName)
		updated++
	}

	fmt.Printf("\n📈 Migration Summary:\n")
	fmt.Printf("   Updated: %d properties\n", updated)
	fmt.Printf("   Skipped: %d properties\n", skipped)
	fmt.Printf("   Total: %d properties\n", len(docs))
}
