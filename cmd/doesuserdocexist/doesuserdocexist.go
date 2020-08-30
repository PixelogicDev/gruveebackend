package doesuserdocexist

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	firestoreClient *firestore.Client
	logger          sawmill.Logger
)

func init() {
	log.Println("DoesUserDocExist intialized")
}

// DoesUserDocExist checks to see if there is already a Firebase user document for someone right before they sign in
func DoesUserDocExist(writer http.ResponseWriter, request *http.Request) {
	doesUserDocExist := false

	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initWithEnvErr, nil)
		return
	}

	// Get UserId
	var reqData doesUserDocExistReq

	reqDataErr := json.NewDecoder(request.Body).Decode(&reqData)
	if reqDataErr != nil {
		http.Error(writer, reqDataErr.Error(), http.StatusInternalServerError)
		logger.LogErr("ReqData Decoder", reqDataErr, request)
		return
	}

	logger.Log("DoesUserDocExist", "Request data decoded successfully.")

	// Make a Firebase request to see if user document is already create with the given uid
	snapshot, snapshotErr := firestoreClient.Collection("users").Doc(reqData.UID).Get(context.Background())
	if status.Code(snapshotErr) != codes.NotFound && snapshot.Exists() {
		logger.Log("DoesUserDocExist", "Found snapshot for user.")
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
