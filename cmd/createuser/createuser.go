package createuser

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
		log.Printf("CreateUser [Init Firestore]: %v", err)
		return
	}
	firestoreClient = client
	log.Println("CreateUser Initialized")
}

// CreateUser will write a new Firebase user to Firestore
func CreateUser(writer http.ResponseWriter, request *http.Request) {
	var firestoreUser firebase.FirestoreUser

	jsonDecodeErr := json.NewDecoder(request.Body).Decode(&firestoreUser)
	if jsonDecodeErr != nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [firestoreUser Decoder]: %v", jsonDecodeErr)
		return
	}

	// Write FirestoreUser to Firestore
	_, writeErr := firestoreClient.Collection("users").Doc(firestoreUser.ID).Set(context.Background(), firestoreUser)
	if writeErr != nil {
		http.Error(writer, writeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateUser [fireStore Set]: %v", writeErr)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
