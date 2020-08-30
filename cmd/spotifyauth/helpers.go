package spotifyauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
		hostname = os.Getenv("HOSTNAME_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
		hostname = os.Getenv("HOSTNAME_PROD")
	}

	// Initialize client
	httpClient = &http.Client{}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("AuthorizeWithSpotify [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "AuthorizeWithSpotify")
	if err != nil {
		log.Printf("AuthorizeWithSpotify [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}

// sillyonly - "So 140 char? is this twitter or a coding stream!" (03/02/20)
func getUser(uid string) (*social.AuthorizeWithSpotifyResponse, error) {
	logger.Log("GetUser", "Starting...")

	// Go to firestore and check for uid
	fbID := "spotify:" + uid
	userRef := firestoreClient.Doc("users/" + fbID)
	if userRef == nil {
		return nil, fmt.Errorf("doesUserExist: users/%s is an odd path", fbID)
	}

	logger.Log("GetUser", "Received user doc.")

	// If uid does not exist return nil
	userSnap, err := userRef.Get(context.Background())
	if status.Code(err) == codes.NotFound {
		logger.Log("GetUser", fmt.Sprintf("User with id %s was not found", fbID))
		return nil, nil
	}

	logger.Log("GetUser", "Received user doc.")

	// UID does exist, return firestore user
	var firestoreUser firebase.FirestoreUser
	dataErr := userSnap.DataTo(&firestoreUser)
	if dataErr != nil {
		return nil, fmt.Errorf("doesUserExist: %v", dataErr)
	}

	logger.Log("GetUser", "Received data for user doc.")

	// Get references from socialPlatforms
	socialPlatformSnaps, socialPlatformSnapsErr := fetchSnapshots(firestoreUser.SocialPlatforms)
	if socialPlatformSnapsErr != nil {
		return nil, fmt.Errorf("FetchSnapshots: %v", socialPlatformSnapsErr)
	}

	logger.Log("GetUser", "Received snapshots.")

	// Convert socialPlatforms to data
	socialPlatforms, preferredPlatform := snapsToSocialPlatformData(socialPlatformSnaps)

	logger.Log("GetUser", "Received data from socialplatforms.")

	// Get references from playlists
	playlistsSnaps, playlistSnapsErr := fetchSnapshots(firestoreUser.Playlists)
	if playlistSnapsErr != nil {
		return nil, fmt.Errorf("FetchSnapshots: %v", playlistSnapsErr)
	}

	logger.Log("GetUser", "Snapshots received.")

	// Convert playlists to data
	playlists := snapsToPlaylistData(playlistsSnaps)

	// Convert user to response object
	authorizeWithSpotifyResponse := social.AuthorizeWithSpotifyResponse{
		Email:                   firestoreUser.Email,
		ID:                      firestoreUser.ID,
		Playlists:               playlists,
		PreferredSocialPlatform: preferredPlatform,
		SocialPlatforms:         socialPlatforms,
		Username:                firestoreUser.Username,
	}

	logger.Log("GetUser", "Successfully received auth response.")

	return &authorizeWithSpotifyResponse, nil
}

// fetchSnapshots takes in an array for Firestore Documents references and return their DocumentSnapshots
func fetchSnapshots(refs []*firestore.DocumentRef) ([]*firestore.DocumentSnapshot, error) {
	logger.Log("FetchSnapshots", "Starting...")

	docsnaps, err := firestoreClient.GetAll(context.Background(), refs)
	if err != nil {
		return nil, fmt.Errorf("fetchChildRefs: %v", err)
	}

	logger.Log("FetchSnapshots", "Successfully received snapshots.")

	return docsnaps, nil
}

// snapsToPlaylistData takes in array of Firestore DocumentSnapshots and retursn array of FirestorePlaylists
func snapsToPlaylistData(snaps []*firestore.DocumentSnapshot) []firebase.FirestorePlaylist {
	logger.Log("SnapsToPlaylistData", "Starting...")

	var playlists []firebase.FirestorePlaylist

	for _, playlistSnap := range snaps {
		var playlist firebase.FirestorePlaylist

		dataErr := playlistSnap.DataTo(&playlist)
		if dataErr != nil {
			logger.Log("SnapsToPlaylistData", "Encountered error while parsing playlist snapshot.")
			logger.LogErr("SnapsToPlaylistData", dataErr, nil)
			continue
		}

		playlists = append(playlists, playlist)
	}

	logger.Log("SnapsToPlaylistData", "Successfully received playlists.")

	return playlists
}

// snapsToSocialPlatformData takes in array of Firestore DocumentSnapshots and retursn array of FirestoreSocialPlatforms & PreferredPlatform
func snapsToSocialPlatformData(snaps []*firestore.DocumentSnapshot) ([]firebase.FirestoreSocialPlatform, firebase.FirestoreSocialPlatform) {
	logger.Log("SnapsToSocialPlatformData", "Starting...")

	var socialPlatforms []firebase.FirestoreSocialPlatform
	var preferredService firebase.FirestoreSocialPlatform

	for _, socialSnaps := range snaps {
		var socialPlatform firebase.FirestoreSocialPlatform

		dataErr := socialSnaps.DataTo(&socialPlatform)
		if dataErr != nil {
			logger.Log("SnapsToSocialPlatformData", "Encountered error while parsing socialSnaps.")
			logger.LogErr("SnapsToSocialPlatformData", dataErr, nil)
			continue
		}

		socialPlatforms = append(socialPlatforms, socialPlatform)

		if socialPlatform.IsPreferredService {
			preferredService = socialPlatform
		}
	}

	logger.Log("SnapsToSocialPlatformData", "Successfully receives social platforms.")

	return socialPlatforms, preferredService
}

// createUser takes in the spotify response and returns a new firebase user
func createUser(spotifyResp social.SpotifyMeResponse,
	socialPlatDocRef *firestore.DocumentRef) (*firebase.FirestoreUser, error) {
	logger.Log("CreateUser", "Starting...")

	var createUserURI = hostname + "/createUser"

	// Get profile image
	var profileImage firebase.SpotifyImage
	if len(spotifyResp.Images) > 0 {
		profileImage = spotifyResp.Images[0]
	} else {
		profileImage = firebase.SpotifyImage{}
	}

	// Create, CreateUser Request object
	var createUserReq = social.CreateUserReq{
		Email:              spotifyResp.Email,
		ID:                 "spotify:" + spotifyResp.ID,
		SocialPlatformPath: "social_platforms/" + socialPlatDocRef.ID,
		ProfileImage:       &profileImage,
		Username:           spotifyResp.DisplayName,
	}

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(createUserReq)
	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	logger.Log("CreateUser", "Successfully created JSON body.")

	// Create Request
	createUser, createUserErr := http.NewRequest("POST", createUserURI, bytes.NewBuffer(jsonPlatform))
	if createUserErr != nil {
		return nil, fmt.Errorf(createUserErr.Error())
	}

	logger.Log("CreateUser", "Generated request.")

	createUser.Header.Add("Content-Type", "application/json")
	createUserResp, httpErr := httpClient.Do(createUser)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	logger.Log("CreateUser", "Received response.")

	if createUserResp.StatusCode != http.StatusOK {
		logger.Log("CreateUser", "Error in response. Decoding...")

		// Get error from body
		var body []byte
		body, _ = ioutil.ReadAll(createUserResp.Body)
		return nil, fmt.Errorf((string(body)))
	}

	var firestoreUser firebase.FirestoreUser
	respDecodeErr := json.NewDecoder(createUserResp.Body).Decode(&firestoreUser)
	if respDecodeErr != nil {
		return nil, fmt.Errorf(respDecodeErr.Error())
	}

	logger.Log("CreateUser", "Successfully decoded firestoreUser.")

	return &firestoreUser, nil
}

// createSocialPlatform calls our CreateSocialPlatform Firebase Function to create & write new platform to DB
func createSocialPlatform(spotifyResp social.SpotifyMeResponse,
	authReq social.SpotifyAuthRequest) (*firestore.DocumentRef, *firebase.FirestoreSocialPlatform, error) {
	logger.Log("CreateSocialPlatform", "Starting...")

	var createSocialPlatformURI = hostname + "/createSocialPlatform"

	// Create request body
	var isPremium = false
	if spotifyResp.Product == "premium" {
		isPremium = true
	}

	var profileImage firebase.SpotifyImage
	if len(spotifyResp.Images) > 0 {
		profileImage = spotifyResp.Images[0]
	} else {
		profileImage = firebase.SpotifyImage{}
	}

	// Adds the expiresIn time to current time
	var expiredAtStr = time.Now().Add(time.Second * time.Duration(authReq.ExpiresIn))

	var apiToken = firebase.APIToken{
		CreatedAt: time.Now().Format(time.RFC3339),
		ExpiredAt: expiredAtStr.Format(time.RFC3339),
		ExpiresIn: authReq.ExpiresIn,
		Token:     authReq.APIToken,
	}

	logger.Log("CreateSocialPlatform", "Generated API Token.")

	// Object that we will write to Firestore
	var platform = firebase.FirestoreSocialPlatform{
		APIToken:           apiToken,
		RefreshToken:       authReq.RefreshToken,
		Email:              spotifyResp.Email,
		ID:                 spotifyResp.ID,
		IsPreferredService: true, // If creating a new user, this is the first platform which should be the default
		IsPremium:          isPremium,
		PlatformName:       "spotify",
		ProfileImage:       profileImage,
		Username:           spotifyResp.DisplayName,
	}

	logger.Log("CreateSocialPlatform", "Generated Platform.")

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(platform)
	if jsonErr != nil {
		return nil, nil, fmt.Errorf(jsonErr.Error())
	}

	logger.Log("CreateSocialPlatform", "Generated JSON body")

	// Create Request
	socialPlatformReq, newReqErr := http.NewRequest("POST", createSocialPlatformURI, bytes.NewBuffer(jsonPlatform))
	if newReqErr != nil {
		return nil, nil, fmt.Errorf(newReqErr.Error())
	}

	logger.Log("CreateSocialPlatform", "Generated request.")

	// Run firebase function to write platform to database
	socialPlatformReq.Header.Add("Content-Type", "application/json")
	socialPlatformResp, httpErr := httpClient.Do(socialPlatformReq)
	if httpErr != nil {
		return nil, nil, fmt.Errorf(httpErr.Error())
	}

	logger.Log("CreateSocialPlatform", "Received response.")

	if socialPlatformResp.StatusCode != http.StatusOK {
		logger.Log("CreateSocialPlatform", "Error in response. Decoding...")

		// Get error from body
		var body, _ = ioutil.ReadAll(socialPlatformResp.Body)
		return nil, nil, fmt.Errorf(string(body))
	}

	// Get Document reference
	platformRef := firestoreClient.Doc("social_platforms/" + platform.ID)
	if platformRef == nil {
		return nil, nil, fmt.Errorf("Odd number of IDs or the ID was empty")
	}

	logger.Log("CreateSocialPlatform", "Successfully received social platform doc.")

	return platformRef, &platform, nil
}

// getCustomRoken calles our GenerateToken Firebase Function to create & return custom JWT
func getCustomToken(uid string) (*social.GenerateTokenResponse, error) {
	logger.Log("GetCustomToken", "Starting...")

	var generateTokenURI = hostname + "/generateCustomToken"
	var tokenRequest = social.GenerateTokenRequest{
		UID: uid,
	}

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(tokenRequest)
	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	logger.Log("GetCustomToken", "Generated JSON body.")

	// Create Request
	customTokenReq, customTokenReqErr := http.NewRequest("POST", generateTokenURI, bytes.NewBuffer(jsonPlatform))
	if customTokenReqErr != nil {
		return nil, fmt.Errorf(customTokenReqErr.Error())
	}

	logger.Log("GetCustomToken", "Generated request.")

	customTokenReq.Header.Add("Content-Type", "application/json")
	customTokenResp, httpErr := httpClient.Do(customTokenReq)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	logger.Log("GetCustomToken", "Received response.")

	// Decode the token to send back
	var tokenResponse social.GenerateTokenResponse
	customTokenDecodeErr := json.NewDecoder(customTokenResp.Body).Decode(&tokenResponse)
	if customTokenDecodeErr != nil {
		return nil, fmt.Errorf(customTokenDecodeErr.Error())
	}

	logger.Log("GetCustomToken", "Successfully decoded token.")

	return &tokenResponse, nil
}
