package createuser

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

var firestoreClient *firestore.Client

func init() {
	client, err := firestore.NewClient(context.Background(), "gruvee-3b7c4")
	if err != nil {
		log.Printf("CreateUser [Init Firestore]: %v", err)
		return
	}
	firestoreClient = client
	log.Println("CreateUser Initialized")
}

// CreateUser will write a new Firebase user to Firestore
func CreateUser(writer http.ResponseWriter, request *http.Request) {
	var createUserReq social.CreateUserReq

	jsonDecodeErr := json.NewDecoder(request.Body).Decode(&createUserReq)
	if jsonDecodeErr != nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [social.CreateUserReq Decoder]: %v", jsonDecodeErr)
		return
	}

	// Get Document references for social platform
	socialPlatDocRef := firestoreClient.Doc(createUserReq.SocialPlatformPath)
	if socialPlatDocRef == nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [social.CreateUserReq Decoder]: %v", jsonDecodeErr)
		return
	}

	// Create Firestore user
	firestoreUser := firebase.FirestoreUser{
		Email:                   createUserReq.Email,
		ID:                      createUserReq.ID,
		Playlists:               []*firestore.DocumentRef{},
		PreferredSocialPlatform: socialPlatDocRef,
		ProfileImage:            createUserReq.ProfileImage,
		SocialPlatforms:         []*firestore.DocumentRef{socialPlatDocRef},
		Username:                createUserReq.Username,
	}

	// Write FirestoreUser to Firestore
	_, writeErr := firestoreClient.Collection("users").Doc(firestoreUser.ID).Set(context.Background(), firestoreUser)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [fireStore Set]: %v", writeErr)
		return
	}

	// Return Firestore User
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(firestoreUser)
}
