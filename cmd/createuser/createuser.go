package createuser

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

var (
	firestoreClient *firestore.Client
	logger          sawmill.Logger
)

func init() {
	log.Println("CreateUser Initialized")
}

// CreateUser will write a new Firebase user to Firestore
func CreateUser(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initWithEnvErr, nil)
		return
	}

	var createUserReq social.CreateUserReq

	jsonDecodeErr := json.NewDecoder(request.Body).Decode(&createUserReq)
	if jsonDecodeErr != nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("CreateUserReq Decoder", jsonDecodeErr, request)
		return
	}

	logger.Log("CreateUser", "Decoded Request successfully.")

	// Get Document references for social platform
	socialPlatDocRef := firestoreClient.Doc(createUserReq.SocialPlatformPath)
	if socialPlatDocRef == nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("CreateUserReq Decoder", jsonDecodeErr, request)
		return
	}

	logger.Log("CreateUser", "Receieved Social Document Reference successfully.")

	// Create Firestore user
	firestoreUser := firebase.FirestoreUser{
		Email:                   createUserReq.Email,
		ID:                      createUserReq.ID,
		Playlists:               []*firestore.DocumentRef{},
		PreferredSocialPlatform: socialPlatDocRef,
		ProfileImage:            createUserReq.ProfileImage,
		SocialPlatforms:         []*firestore.DocumentRef{socialPlatDocRef},
		DisplayName:             createUserReq.DisplayName,
		Username:                createUserReq.Username,
	}

	// Write FirestoreUser to Firestore
	_, writeErr := firestoreClient.Collection("users").Doc(firestoreUser.ID).Set(context.Background(), firestoreUser)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("FireStore Set", writeErr, nil)
		return
	}

	logger.Log("CreateUser", "Successfully wrote new data to Firestore.")

	// Return Firestore User
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(firestoreUser)
}
