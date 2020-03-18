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

// JackGamesFTW - "TriHard 7" (03/18/20)
func init() {
	client, err := firestore.NewClient(context.Background(), "gruvee-3b7c4")
	if err != nil {
		log.Printf("CreateSocialPlatform [Init Firestore]: %v", err)
		return
	}
	firestoreClient = client
	log.Println("CreateSocialPlatform intialized")
}

// CreateSocialPlatform will write a new social platform to firestore
func CreateSocialPlatform(writer http.ResponseWriter, request *http.Request) {
	var socialPlatform firebase.FirestoreSocialPlatform

	socialPlatformErr := json.NewDecoder(request.Body).Decode(&socialPlatform)
	if socialPlatformErr != nil {
		http.Error(writer, socialPlatformErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateSocialPlatform [socialPlatform Decoder]: %v", socialPlatformErr)
		return
	}

	// Write SocialPlatform to Firestore
	_, writeErr := firestoreClient.Collection("social_platforms").Doc(socialPlatform.ID).Set(context.Background(), socialPlatform)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateSocialPlatform [fireStore Set]: %v", writeErr)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
