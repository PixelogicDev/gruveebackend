package updatealgolia

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/functions/metadata"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
)

// algoliaUser implements a partial amount of data from firestoreUser to use for indexing
type algoliaUser struct {
	ObjectID        string `json:"objectID"`
	ID              string `json:"id"`
	Email           string `json:"email"`
	ProfileImageURI string `json:"profileImage"`
	DisplayName     string `json:"displayName"`
	Username        string `json:"username"`
}

// UpdateAlgolia sends new data to Algolia service for indexing
func UpdateAlgolia(ctx context.Context, event firebase.FirestoreEvent) error {
	log.Println("[UpdateAlgolia] Starting update...")

	// Get IDs
	algoliaAppID := os.Getenv("ALGOLIA_APP_ID")
	if algoliaAppID == "" {
		log.Println("Algolia App ID was empty in yaml file")
		return fmt.Errorf("Algolia App ID was empty in yaml file")
	}

	algoliaSecretID := os.Getenv("ALGOLIA_SECRET_ID")
	if algoliaSecretID == "" {
		log.Println("Algolia Secret ID was empty in yaml file")
		return fmt.Errorf("Algolia Secret ID was empty in yaml file")
	}

	var algoliaIndexName string
	if os.Getenv("ENVIRONMENT") == "DEV" {
		algoliaIndexName = os.Getenv("ALGOLIA_INDEX_NAME_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		algoliaIndexName = os.Getenv("ALGOLIA_INDEX_NAME_PROD")
	}

	if algoliaIndexName == "" {
		log.Println("Algolia Index Name was empty in yaml file")
		return fmt.Errorf("Algolia Index Name was empty in yaml file")
	}

	// Init our client
	client := search.NewClient(algoliaAppID, algoliaSecretID)
	index := client.InitIndex(algoliaIndexName)

	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
	}

	meta, err := metadata.FromContext(ctx)
	if err != nil {
		logger.LogErr(err, "metadata.FromContext", nil)
		return fmt.Errorf("metadata.FromContext: %v", err)
	}

	// Print out our trigger data
	log.Printf("Function triggered by change to: %v", meta.Resource)
	log.Printf("Event Trigger: %v", event)

	// Write objects to Algolia
	res, err := index.SaveObject(algoliaUser{
		ObjectID:        event.Value.Fields.ID.StringValue,
		ID:              event.Value.Fields.ID.StringValue,
		Email:           event.Value.Fields.Email.StringValue,
		ProfileImageURI: event.Value.Fields.ProfileImage.MapValue.Fields.URL.StringValue,
		DisplayName:     event.Value.Fields.DisplayName.StringValue,
		Username:        event.Value.Fields.Username.StringValue,
	})

	log.Printf("[UpdateAlgolia] SaveObject Res: %v", res)

	if err != nil {
		logger.LogErr(err, "index.SaveObject", nil)
		return fmt.Errorf(err.Error())
	}

	return nil
}
