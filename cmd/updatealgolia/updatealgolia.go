package updatealgolia

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/functions/metadata"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

var (
	logger   sawmill.Logger
	hostname string
)

func init() {
	log.Println("UpdateAlgolia intialized")
}

// UpdateAlgolia sends new data to Algolia service for indexing
func UpdateAlgolia(ctx context.Context, event firebase.FirestoreEvent) error {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		logger.LogErr("InitWithEnvErr", initWithEnvErr, nil)
		return initWithEnvErr
	}

	// Get IDs
	algoliaAppID := os.Getenv("ALGOLIA_APP_ID")
	if algoliaAppID == "" {
		error := fmt.Errorf("Algolia App ID was empty in yaml file")
		logger.LogErr("UpdateAlgolia", error, nil)
		return error
	}

	algoliaSecretID := os.Getenv("ALGOLIA_SECRET_ID")
	if algoliaSecretID == "" {
		error := fmt.Errorf("Algolia Secret ID was empty in yaml file")
		logger.LogErr("UpdateAlgolia", error, nil)
		return error
	}

	var algoliaIndexName string
	if os.Getenv("ENVIRONMENT") == "DEV" {
		algoliaIndexName = os.Getenv("ALGOLIA_INDEX_NAME_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		algoliaIndexName = os.Getenv("ALGOLIA_INDEX_NAME_PROD")
	}

	if algoliaIndexName == "" {
		error := fmt.Errorf("Algolia Index Name was empty in yaml file")
		logger.LogErr("UpdateAlgolia", error, nil)
		return error
	}

	// Init our client
	client := search.NewClient(algoliaAppID, algoliaSecretID)
	index := client.InitIndex(algoliaIndexName)

	meta, err := metadata.FromContext(ctx)
	if err != nil {
		error := fmt.Errorf("metadata.FromContext: %v", err)
		logger.LogErr("UpdateAlgolia", error, nil)
		return error
	}

	logger.Log("UpdateAlgolia", "Algolia clients created.")

	// Print out our trigger data
	logger.Log("UpdateAlgolia", fmt.Sprintf("Function triggered by change to: %v", meta.Resource))
	logger.Log("UpdateAlgolia", fmt.Sprintf("Event Trigger: %v", event))

	// Write objects to Algolia
	res, err := index.SaveObject(algoliaUser{
		ObjectID:        event.Value.Fields.ID.StringValue,
		ID:              event.Value.Fields.ID.StringValue,
		Email:           event.Value.Fields.Email.StringValue,
		ProfileImageURI: event.Value.Fields.ProfileImage.MapValue.Fields.URL.StringValue,
		DisplayName:     event.Value.Fields.DisplayName.StringValue,
		Username:        event.Value.Fields.Username.StringValue,
	})

	logger.Log("UpdateAlgolia", fmt.Sprintf("SaveObject Res: %v", res))

	if err != nil {
		logger.LogErr("UpdateAlgolia", err, nil)
		return err
	}

	logger.Log("UpdateAlgolia", "Successfully wrote new user to Aloglia.")

	return nil
}
