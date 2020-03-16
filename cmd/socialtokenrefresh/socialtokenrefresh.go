package socialtokenrefresh

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
)

var firestoreClient *firestore.Client

func init() {
	client, err := firestore.NewClient(context.Background(), "gruvee-3b7c4")
	if err != nil {
		log.Printf("AuthorizeWithSpotify [Init Firestore]: %v", err)
		return
	}

	firestoreClient = client
	log.Println("SocialTokenRefreshRequest initialized")
}

// socialTokenRefreshRequest includes uid to grab all social platforms for user
type socialTokenRefreshRequest struct {
	UID string `json:"uid"`
}

// SocialTokenRefresh checks to see if we need to refresh current API tokens for social platforms
func SocialTokenRefresh(writer http.ResponseWriter, request *http.Request) {
	// Receive payload that includes uid
	var socialTokenReq socialTokenRefreshRequest

	// Decode payload
	socialTokenErr := json.NewDecoder(request.Body).Decode(&socialTokenReq)
	if socialTokenErr != nil {
		http.Error(writer, socialTokenErr.Error(), http.StatusInternalServerError)
		log.Printf("SocialTokenRefresh [socialTokenReq Decoder]: %v", socialTokenErr)
		return
	}

	// Go to Firestore and get the platforms for user
	platsToRefresh, platformErr := getUserPlatformsToRefresh(socialTokenReq.UID)
	if platformErr != nil {
		http.Error(writer, platformErr.Error(), http.StatusInternalServerError)
		log.Printf("SocialTokenRefresh [getUserPlatforms]: %v", platformErr)
		return
	}

	// Run logic to check if tokens need to be refreshed
	refreshTokenErr := refreshTokens(*platsToRefresh)
	if refreshTokenErr != nil {
		http.Error(writer, refreshTokenErr.Error(), http.StatusInternalServerError)
		log.Printf("SocialTokenRefresh [refreshTokenErr]: %v", refreshTokenErr)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

// Helpers
func getUserPlatformsToRefresh(uid string) (*[]firebase.FirestoreSocialPlatform, error) {
	// Go to Firebase and get document references for all social platforms
	snapshot, snapshotErr := firestoreClient.Collection("users").Doc(uid).Get(context.Background())
	if snapshotErr != nil {
		return nil, fmt.Errorf(snapshotErr.Error())
	}

	// Grab socialPlatforms array
	var firestoreUser firebase.FirestoreUser
	dataToErr := snapshot.DataTo(&firestoreUser)
	if dataToErr != nil {
		return nil, fmt.Errorf(dataToErr.Error())
	}

	// Check to see which platforms need to be refreshed
	socialPlatforms, fetchRefErr := fetchChildRefs(firestoreUser.SocialPlatforms)
	if fetchRefErr != nil {
		return nil, fmt.Errorf(fetchRefErr.Error())
	}

	// Return those platforms to main
	return socialPlatforms, nil
}

// fetchChildRefs will convert document references to FiresstoreSocilaPlatform Objects
func fetchChildRefs(refs []*firestore.DocumentRef) (*[]firebase.FirestoreSocialPlatform, error) {
	docsnaps, err := firestoreClient.GetAll(context.Background(), refs)
	if err != nil {
		return nil, fmt.Errorf("fetchChildRefs: %v", err)
	}

	var socialPlatforms []firebase.FirestoreSocialPlatform
	for _, userSnap := range docsnaps {
		var socialPlatform firebase.FirestoreSocialPlatform

		dataErr := userSnap.DataTo(&socialPlatform)
		if dataErr != nil {
			log.Printf("Encountered error while parsing userSnapshot.")
			log.Printf("%v", dataErr)
			continue
		}

		socialPlatforms = append(socialPlatforms, socialPlatform)
	}

	return &socialPlatforms, nil
}

// refreshTokens goes through social platform objects and refreshes tokens as necessary
func refreshTokens(socialPlatforms []firebase.FirestoreSocialPlatform) error {
	// Get current time
	var currentTime = time.Now()
	for _, platform := range socialPlatforms {
		fmt.Printf("Expires In: %d seconds\n", platform.APIToken.ExpiresIn)
		fmt.Printf("Expired At: %s\n", platform.APIToken.ExpiredAt)
		fmt.Printf("Created At: %s\n", platform.APIToken.CreatedAt)

		expiredAtTime, expiredAtTimeErr := time.Parse(time.RFC3339, platform.APIToken.ExpiredAt)
		if expiredAtTimeErr != nil {
			fmt.Println(expiredAtTimeErr.Error())
			continue
		}

		if currentTime.After(expiredAtTime) {
			// Call API refresh
			fmt.Printf("%s access token is expired.", platform.PlatformName)
		}
	}

	return nil
}
