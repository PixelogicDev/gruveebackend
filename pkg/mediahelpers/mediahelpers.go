package mediahelpers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
)

// GetMediaFromFirestore takes in a firestoreClient,  mediaProvider and the mediaId and checks to see if the media exists
func GetMediaFromFirestore(firestoreClient firestore.Client, mediaProvider string, mediaID string) (*firebase.FirestoreMedia, error) {
	var queryString string

	switch mediaProvider {
	case "apple":
		log.Println("[GetMediaFromFirestore] Apple provider found.")
		queryString = "apple.id"
	case "spotify":
		log.Println("[GetMediaFromFirestore] Spotify provider found.")
		queryString = "spotify.id"
	default:
		return nil, fmt.Errorf("[GetMediaFromFirestore] Provider passed in was not found: %s", mediaProvider)
	}

	// Construct query
	query := firestoreClient.Collection("songs").Where(queryString, "==", mediaID)

	// Execute & check result length
	results, resultsErr := query.Documents(context.Background()).GetAll()
	if resultsErr != nil {
		return nil, fmt.Errorf("[GetMediaFromFirestore] Error getting all documents from query: %v", resultsErr)
	}

	// Song does exist, return document
	if len(results) > 0 {
		log.Println("[GetMediaFromFirestore] Found song document")
		bytes, bytesErr := json.Marshal(results[0].Data())
		if bytesErr != nil {
			return nil, fmt.Errorf("[GetMediaFromFirestore] Cannot marshal FirestoreMedia data: %v", bytesErr)
		}

		var firestoreMedia firebase.FirestoreMedia
		unmarhsalErr := json.Unmarshal(bytes, &firestoreMedia)
		if unmarhsalErr != nil {
			return nil, fmt.Errorf("[GetMediaFromFirestore] Cannot unmarshal bytes data: %v", unmarhsalErr)
		}

		return &firestoreMedia, nil
	}

	// Song does not exist, need to create new song document
	log.Println("[GetMediaFromFirestore] Did not find song document")
	return nil, nil
}
