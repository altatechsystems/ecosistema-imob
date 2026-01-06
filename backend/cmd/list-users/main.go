package main

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

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

	fmt.Println("📋 Listing Firebase Authentication users:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	iter := authClient.Users(ctx, "")
	count := 0
	for {
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error listing users: %v", err)
		}

		count++
		fmt.Printf("\n👤 User %d:\n", count)
		fmt.Printf("   UID: %s\n", user.UID)
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Name: %s\n", user.DisplayName)
		fmt.Printf("   Phone: %s\n", user.PhoneNumber)
		fmt.Printf("   Verified: %v\n", user.EmailVerified)
		fmt.Printf("   Disabled: %v\n", user.Disabled)
		fmt.Printf("   Created: %v\n", user.UserMetadata.CreationTimestamp)
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total users: %d\n", count)
}
