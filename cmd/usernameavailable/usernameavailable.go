package usernameavailable

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

// usernameAvailableReq includes the username to query the user collection
type usernameAvailableReq struct {
	Username string `json:"username"`
}

// usernameAvailableResp includes a result of true or false
type usernameAvailableResp struct {
	Result bool `json:"result"`
}

var firestoreClient *firestore.Client
var logger sawmill.Logger

func init() {
	log.Println("UsernameAvailable intialized")
}

// UsernameAvailable checks to see if the given username is available to use
func UsernameAvailable(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	isUsernameAvailable := true
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr(initWithEnvErr, "initWithEnv", nil)
		return
	}

	// Get Username
	var reqData usernameAvailableReq

	reqDataErr := json.NewDecoder(request.Body).Decode(&reqData)
	if reqDataErr != nil {
		http.Error(writer, reqDataErr.Error(), http.StatusInternalServerError)
		logger.LogErr(reqDataErr, "reqData Decoder", request)
		return
	}

	// Make a Firebase request to see if user document is already create with the given uid
	snapshots := firestoreClient.Collection("users").Where("username", "==", reqData.Username).Snapshots(context.Background())
	documents, documentsErr := snapshots.Query.Documents(context.Background()).GetAll()
	if documentsErr != nil {
		http.Error(writer, documentsErr.Error(), http.StatusInternalServerError)
		logger.LogErr(documentsErr, "Firebase GetDocumentsQuery", request)
		return
	}

	if len(documents) > 0 {
		log.Printf("[UsernameAvailable] %s has already been taken", reqData.Username)
		isUsernameAvailable = false
	}

	// Create result object
	result := usernameAvailableResp{
		Result: isUsernameAvailable,
	}

	// Send response
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(result)
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
		return fmt.Errorf("UsernameAvailable [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "UsernameAvailable")
	if err != nil {
		log.Printf("UsernameAvailable [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}
