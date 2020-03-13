package socialplatform

// eminyilmazz - "If I got corona, this line is my legacy." (03/12/20)
import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
)

var firestoreClient *firestore.Client

func init() {
	client, err := firestore.NewClient(context.Background(), "gruvee-3b7c4")
	if err != nil {
		log.Printf("SocialPlatform [Init Firestore]: %v", err)
		return
	}

	firestoreClient = client
	log.Printf("SocialPlatform intialized.")
}

// CreateSocialPlatform will write a new social platform to firestore
func CreateSocialPlatform(writer http.ResponseWriter, request *http.Request) {
	// Get payload from body (firebase.FirestoreSocialPlatform)
	var socialPlatform firebase.FirestoreSocialPlatform

	socialPlatformErr := json.NewDecoder(request.Body).Decode(&socialPlatform)
	if socialPlatformErr != nil {
		http.Error(writer, socialPlatformErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateSocialPlatform [socialPlatform Decoder]: %v", socialPlatformErr)
		return
	}

	log.Printf("Received FirestoreSocial payload")
	// Write SocialPlatform to Firestore
	_, writeErr := firestoreClient.Collection("social_platforms").Doc(socialPlatform.ID).Set(context.Background(), socialPlatform)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateSocialPlatform [fireStore Set]: %v", writeErr)
		return
	}

	// If success, write back that we did good!
	writer.WriteHeader(http.StatusOK)
}
