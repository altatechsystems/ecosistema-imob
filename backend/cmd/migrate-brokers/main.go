package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	projectID := "ecosistema-imob-dev"
	databaseID := "imob-dev"
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if credentialsPath == "" {
		credentialsPath = "./config/firebase-adminsdk.json"
	}

	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	fmt.Println("🔍 Buscando todos os tenants...")

	// Get all tenants
	tenantsIter := client.Collection("tenants").Documents(ctx)
	tenants := make([]string, 0)

	for {
		doc, err := tenantsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate tenants: %v", err)
		}
		tenants = append(tenants, doc.Ref.ID)
	}

	fmt.Printf("✅ Encontrados %d tenant(s)\n\n", len(tenants))

	totalBrokersCreated := 0
	totalPropertiesUpdated := 0

	// Process each tenant
	for _, tenantID := range tenants {
		fmt.Printf("📦 Processando tenant: %s\n", tenantID)

		// Map to track unique captadores and their broker IDs
		captadorToBrokerID := make(map[string]string)

		// Get all properties for this tenant
		// Properties are in root collection with tenant_id field
		propertiesIter := client.Collection("properties").Where("tenant_id", "==", tenantID).Documents(ctx)

		properties := make([]struct {
			ID           string
			CaptadorName string
		}, 0)

		// First pass: collect all unique captador names
		for {
			doc, err := propertiesIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("❌ Failed to iterate properties: %v", err)
				continue
			}

			data := doc.Data()
			captadorName, ok := data["captador_name"].(string)
			if !ok || captadorName == "" {
				continue
			}

			properties = append(properties, struct {
				ID           string
				CaptadorName string
			}{
				ID:           doc.Ref.ID,
				CaptadorName: captadorName,
			})

			if _, exists := captadorToBrokerID[captadorName]; !exists {
				captadorToBrokerID[captadorName] = "" // placeholder
			}
		}

		fmt.Printf("  📊 Encontrados %d imóveis com captador\n", len(properties))
		fmt.Printf("  👥 Captadores únicos: %d\n", len(captadorToBrokerID))

		// Second pass: create brokers for each unique captador
		brokersPath := fmt.Sprintf("tenants/%s/brokers", tenantID)
		brokersCreated := 0

		for captadorName := range captadorToBrokerID {
			// Check if broker already exists by name
			existingBroker := findBrokerByName(ctx, client, brokersPath, captadorName)

			if existingBroker != nil {
				fmt.Printf("  ℹ️  Broker '%s' já existe (ID: %s)\n", captadorName, existingBroker.ID)
				captadorToBrokerID[captadorName] = existingBroker.ID
				continue
			}

			// Create new broker
			broker := &models.Broker{
				TenantID: tenantID,
				Name:     captadorName,
				Email:    generateEmailFromName(captadorName),
				CRECI:    "PENDENTE", // Placeholder - needs to be filled manually
				Role:     "broker",
				IsActive: true,
				Bio:      fmt.Sprintf("Corretor cadastrado automaticamente a partir da importação de imóveis."),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Generate ID
			brokerRef := client.Collection(brokersPath).NewDoc()
			broker.ID = brokerRef.ID

			// Create broker document
			_, err := brokerRef.Set(ctx, broker)
			if err != nil {
				log.Printf("  ❌ Failed to create broker '%s': %v", captadorName, err)
				continue
			}

			captadorToBrokerID[captadorName] = broker.ID
			brokersCreated++
			fmt.Printf("  ✅ Broker criado: '%s' (ID: %s)\n", captadorName, broker.ID)
		}

		totalBrokersCreated += brokersCreated
		fmt.Printf("  🎉 Total de brokers criados: %d\n\n", brokersCreated)

		// Third pass: update properties with captador_id
		propertiesUpdated := 0

		for _, prop := range properties {
			brokerID, exists := captadorToBrokerID[prop.CaptadorName]
			if !exists || brokerID == "" {
				log.Printf("  ⚠️  Broker ID não encontrado para captador '%s'", prop.CaptadorName)
				continue
			}

			// Update property with captador_id (properties are in root collection)
			propRef := client.Collection("properties").Doc(prop.ID)
			_, err := propRef.Update(ctx, []firestore.Update{
				{Path: "captador_id", Value: brokerID},
				{Path: "updated_at", Value: time.Now()},
			})

			if err != nil {
				log.Printf("  ❌ Failed to update property %s: %v", prop.ID, err)
				continue
			}

			propertiesUpdated++
		}

		totalPropertiesUpdated += propertiesUpdated
		fmt.Printf("  ✅ Imóveis atualizados com captador_id: %d\n\n", propertiesUpdated)
	}

	fmt.Println("============================================================")
	fmt.Printf("🎉 MIGRAÇÃO CONCLUÍDA\n")
	fmt.Printf("📊 Total de brokers criados: %d\n", totalBrokersCreated)
	fmt.Printf("🏠 Total de imóveis atualizados: %d\n", totalPropertiesUpdated)
	fmt.Println("============================================================")
}

// findBrokerByName searches for a broker by name in the given collection
func findBrokerByName(ctx context.Context, client *firestore.Client, brokersPath, name string) *models.Broker {
	query := client.Collection(brokersPath).Where("name", "==", name).Limit(1)
	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil
	}
	if err != nil {
		return nil
	}

	var broker models.Broker
	if err := doc.DataTo(&broker); err != nil {
		return nil
	}

	broker.ID = doc.Ref.ID
	return &broker
}

// generateEmailFromName creates a placeholder email from the broker's name
func generateEmailFromName(name string) string {
	// Convert name to lowercase and replace spaces with dots
	email := strings.ToLower(name)
	email = strings.ReplaceAll(email, " ", ".")

	// Remove accents (basic implementation)
	replacements := map[string]string{
		"á": "a", "à": "a", "ã": "a", "â": "a",
		"é": "e", "è": "e", "ê": "e",
		"í": "i", "ì": "i", "î": "i",
		"ó": "o", "ò": "o", "õ": "o", "ô": "o",
		"ú": "u", "ù": "u", "û": "u",
		"ç": "c",
	}

	for old, new := range replacements {
		email = strings.ReplaceAll(email, old, new)
	}

	// Add domain
	email = email + "@pendente.com.br"

	return email
}
