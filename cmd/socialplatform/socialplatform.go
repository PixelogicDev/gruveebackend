package socialplatform

// eminyilmazz - "If I got corona, this line is my legacy." (03/12/20)
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

var firestoreClient *firestore.Client
var logger sawmill.Logger

// JackGamesFTW - "TriHard 7" (03/18/20)
func init() {
	log.Println("CreateSocialPlatform intialized")
}

// CreateSocialPlatform will write a new social platform to firestore
func CreateSocialPlatform(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnvErr", initWithEnvErr, nil)
		return
	}

	var socialPlatform firebase.FirestoreSocialPlatform

	socialPlatformErr := json.NewDecoder(request.Body).Decode(&socialPlatform)
	if socialPlatformErr != nil {
		http.Error(writer, socialPlatformErr.Error(), http.StatusInternalServerError)
		logger.LogErr("SocialPlatform Decoder", socialPlatformErr, request)
		return
	}

	// Write SocialPlatform to Firestore
	_, writeErr := firestoreClient.Collection("social_platforms").Doc(socialPlatform.ID).Set(context.Background(), socialPlatform)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("FireStore Set", writeErr, nil)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

// Helpers

// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("SocialTokenRefresh [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "CreateSocialPlatform")
	if err != nil {
		log.Printf("CreateSocialPlatform [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}
