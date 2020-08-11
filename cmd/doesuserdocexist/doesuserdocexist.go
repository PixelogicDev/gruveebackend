package doesuserdocexist

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// doesUserDocExistReq includes the uid of the user we are checking
type doesUserDocExistReq struct {
	UID string `json:"uid"`
}

// doesUserDocExistResp includes a result of true or false
type doesUserDocExistResp struct {
	Result bool `json:"result"`
}

var firestoreClient *firestore.Client
var logger sawmill.Logger

func init() {
	log.Println("DoesUserDocExist intialized")
}

// DoesUserDocExist checks to see if there is already a Firebase user document for someone right before they sign in
func DoesUserDocExist(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	doesUserDocExist := false
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr(initWithEnvErr, "initWithEnv", nil)
		return
	}

	// Get UserId
	var reqData doesUserDocExistReq

	reqDataErr := json.NewDecoder(request.Body).Decode(&reqData)
	if reqDataErr != nil {
		http.Error(writer, reqDataErr.Error(), http.StatusInternalServerError)
		logger.LogErr(reqDataErr, "reqData Decoder", request)
		return
	}

	// Make a Firebase request to see if user document is already create with the given uid
	snapshot, snapshotErr := firestoreClient.Collection("users").Doc(reqData.UID).Get(context.Background())
	if status.Code(snapshotErr) != codes.NotFound && snapshot.Exists() {
		doesUserDocExist = true
	}

	// Create result object
	result := doesUserDocExistResp{
		Result: doesUserDocExist,
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
		return fmt.Errorf("DoesUserDocExist [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), "NOT DEV", "DoesUserDocExist")
	if err != nil {
		log.Printf("DoesUserDocExist [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}
