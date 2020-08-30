package socialplatform

// eminyilmazz - "If I got corona, this line is my legacy." (03/12/20)
import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

var (
	firestoreClient *firestore.Client
	logger          sawmill.Logger
)

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

	logger.Log("CreateSocialPlatform", "Decoded socialPlatform.")

	// Write SocialPlatform to Firestore
	_, writeErr := firestoreClient.Collection("social_platforms").Doc(socialPlatform.ID).Set(context.Background(), socialPlatform)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("FireStore Set", writeErr, nil)
		return
	}

	logger.Log("CreateSocialPlatform", "Successfully wrote platform to firestore.")

	writer.WriteHeader(http.StatusOK)
}
