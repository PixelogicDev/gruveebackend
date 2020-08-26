package createprovideruser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

// updateProviderUserReq takes in the Firebase Provider UID and the platform provider UID to map
type createProviderUserReq struct {
	FirebaseProviderUID string `json:"firebaseProviderUID"`
	PlatformProviderUID string `json:"platformProviderUID"`
}

// providerUser takes the platformUser document reference and stores in new collection
type providerUser struct {
	PlatformUserRef *firestore.DocumentRef `firestore:"platformUserReference"`
}

var firestoreClient *firestore.Client
var logger sawmill.Logger

// CreateProviderUser will check to see if the newly created user needs to be added to the providers_users collection
func CreateProviderUser(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initErr := initWithEnv()
	if initErr != nil {
		http.Error(writer, initErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initErr, nil)
		return
	}

	// Decode
	var reqData createProviderUserReq

	reqDataErr := json.NewDecoder(request.Body).Decode(&reqData)
	if reqDataErr != nil {
		http.Error(writer, reqDataErr.Error(), http.StatusInternalServerError)
		logger.LogErr("ReqData Decoder", reqDataErr, request)
		return
	}

	// Create document references
	firebaseProviderDocRef := firestoreClient.Doc("provider_users/" + reqData.FirebaseProviderUID)
	platformProviderDocRef := firestoreClient.Doc("users/" + reqData.PlatformProviderUID)

	// Create ProviderUser Object
	providerUserData := providerUser{
		PlatformUserRef: platformProviderDocRef,
	}

	// Write to Firestore
	_, writeErr := firebaseProviderDocRef.Set(context.Background(), providerUserData)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("FireStore Set", writeErr, request)
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

	// Get Project ID
	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("CreateProviderUser [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "CreateProviderUser")
	if err != nil {
		log.Printf("CreateAppleDevToken [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}
