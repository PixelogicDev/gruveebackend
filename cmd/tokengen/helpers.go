package tokengen

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
		hostname = os.Getenv("HOSTNAME_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
		hostname = os.Getenv("HOSTNAME_PROD")
	}

	// Init Firebase App
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("Firebase.NewApp: %v", err)
	}

	// Init Fireabase Auth Admin
	client, err = app.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("Auth.Client: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "AuthorizeWithSpotify")
	if err != nil {
		log.Printf("AuthorizeWithSpotify [Init Sawmill]: %v", err)
	}

	logger = sawmillLogger

	return nil
}
