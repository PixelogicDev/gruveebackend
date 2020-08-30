package createprovideruser

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

var (
	firestoreClient *firestore.Client
	logger          sawmill.Logger
)

func init() {
	log.Println("CreateProviderUser initialized.")
}

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

	logger.Log("CreateProviderUser", "Request was decoded successfully")

	// Create document references
	firebaseProviderDocRef := firestoreClient.Doc("provider_users/" + reqData.FirebaseProviderUID)
	platformProviderDocRef := firestoreClient.Doc("users/" + reqData.PlatformProviderUID)

	logger.Log("CreateProviderUser", "FirebaseProvider & PlatformProvider docs creates")

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

	logger.Log("CreateProviderUser", "Write FirebaseProvider was successful.")

	writer.WriteHeader(http.StatusOK)
}
