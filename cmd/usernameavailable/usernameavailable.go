package usernameavailable

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

var (
	firestoreClient     *firestore.Client
	logger              sawmill.Logger
	isUsernameAvailable bool
)

func init() {
	log.Println("UsernameAvailable intialized")
}

// UsernameAvailable checks to see if the given username is available to use
func UsernameAvailable(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initWithEnvErr, nil)
		return
	}

	isUsernameAvailable = true

	// Get Username
	var reqData usernameAvailableReq

	reqDataErr := json.NewDecoder(request.Body).Decode(&reqData)
	if reqDataErr != nil {
		http.Error(writer, reqDataErr.Error(), http.StatusInternalServerError)
		logger.LogErr("ReqData Decoder", reqDataErr, request)
		return
	}

	logger.Log("UsernameAvailable", "Decoded request.")

	// Make a Firebase request to see if user document is already create with the given uid
	snapshots := firestoreClient.Collection("users").Where("username", "==", reqData.Username).Snapshots(context.Background())
	documents, documentsErr := snapshots.Query.Documents(context.Background()).GetAll()
	if documentsErr != nil {
		http.Error(writer, documentsErr.Error(), http.StatusInternalServerError)
		logger.LogErr("Firebase GetDocumentsQuery", documentsErr, request)
		return
	}

	logger.Log("UsernameAvailable", "Received snapshots.")

	if len(documents) > 0 {
		logger.Log("UsernameAvailable", fmt.Sprintf("%s has already been taken", reqData.Username))
		isUsernameAvailable = false
	}

	// Create result object
	result := usernameAvailableResp{
		Result: isUsernameAvailable,
	}

	logger.Log("UsernameAvailable", "Successfully created result.")

	// Send response
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(result)
}
