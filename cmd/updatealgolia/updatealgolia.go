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
	ObjectID string `json:"objectId"`
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
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

	algoliaIndexName := os.Getenv("ALGOLIA_INDEX_NAME")
	if algoliaIndexName == "" {
		log.Println("Algolia Index Name was empty in yaml file")
		return fmt.Errorf("Algolia Index Name was empty in yaml file")
	}

	// Init our client
	client := search.NewClient(algoliaAppID, algoliaSecretID)
	index := client.InitIndex(algoliaIndexName)

	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %v", err)
	}

	// Print out our trigger data
	log.Printf("Function triggered by change to: %v", meta.Resource)
	log.Printf("New value: %v", event)

	// Write objects to Algolia
	res, err := index.SaveObject(algoliaUser{
		ObjectID: event.Value.Fields.ID.StringValue,
		ID:       event.Value.Fields.ID.StringValue,
		Email:    event.Value.Fields.Email.StringValue,
		Username: event.Value.Fields.Username.StringValue,
	})

	log.Printf("[UpdateAlgolia] SaveObject Res: %v", res)

	if err != nil {
		log.Println(err.Error())
		return fmt.Errorf(err.Error())
	}

	return nil
}
