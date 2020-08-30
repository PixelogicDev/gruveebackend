package createsocialplaylist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Helpers
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

	// Set httpClient
	httpClient = &http.Client{}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("SocialTokenRefresh [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "CreateSocialPlaylist")
	if err != nil {
		log.Printf("CreateSocial Playlist [Init Sawmill]: %v", err)
	}

	// DR_DinoMight - "Note to self! Welcome, Dr_DinoMight, Otherwise he'll spit his dummy out!" (06.15.20)
	firestoreClient = client
	logger = sawmillLogger
	return nil
}

// createPlaylist takes the social platform and playlist information and creates a playlist on the user's preferred platform
func createPlaylist(endpoint string, platform firebase.FirestoreSocialPlatform, playlistName string) error {
	logger.Log("CreatePlaylist", "Starting...")

	var request *http.Request
	var requestErr error

	// Check for platform
	if platform.PlatformName == "spotify" {
		logger.Log("CreatePlaylist", "Creating playlist for Spotify")
		request, requestErr = createSpotifyPlaylistRequest(playlistName, endpoint, platform.APIToken.Token)
	} else if platform.PlatformName == "apple" {
		logger.Log("CreatePlaylist", "Creating playlist for Apple Music")
		request, requestErr = createAppleMusicPlaylistRequest(playlistName, endpoint, platform.APIToken.Token)
	}

	if requestErr != nil {
		return requestErr
	}

	createPlaylistResp, httpErr := httpClient.Do(request)
	if httpErr != nil {
		return httpErr
	}

	logger.Log("CreatePlaylist", "Successfully called create playlist API")

	// If we have errors, lets parse 'em out
	if createPlaylistResp.StatusCode != http.StatusOK && createPlaylistResp.StatusCode != http.StatusCreated {
		if platform.PlatformName == "spotify" {
			logger.Log("CreatePlaylist", "Received Spotify error. Parsing...")

			var spotifyErrorObj social.SpotifyRequestError

			err := json.NewDecoder(createPlaylistResp.Body).Decode(&spotifyErrorObj)
			if err != nil {
				return err
			}

			return fmt.Errorf("Status Code %v: "+spotifyErrorObj.Error.Message, spotifyErrorObj.Error.Status)
		} else if platform.PlatformName == "apple" {
			logger.Log("CreatePlaylist", "Received Apple Music error. Parsing...")

			var appleMusicReqErr social.AppleMusicRequestError

			err := json.NewDecoder(createPlaylistResp.Body).Decode(&appleMusicReqErr)
			if err != nil {
				return err
			}

			// The first error is the most important so for now let's just grab that
			return fmt.Errorf("Status Code %v: "+appleMusicReqErr.Errors[0].Detail, appleMusicReqErr.Errors[0].Status)
		}
	}

	return nil
}

// getAppleDevToken will check our DB for appleDevJWT and return it if there
func getAppleDevToken() (*firebase.FirestoreAppleDevJWT, error) {
	logger.Log("GetAppleDevToken", "Starting...")

	// Go to Firebase and see if appleDevToken exists
	snapshot, snapshotErr := firestoreClient.Collection("internal_tokens").Doc("appleDevToken").Get(context.Background())
	if status.Code(snapshotErr) == codes.NotFound {
		return nil, nil
	}

	if snapshotErr != nil {
		return nil, snapshotErr
	}

	logger.Log("GetAppleDevToken", "Recieved Firebase snapshot")

	var appleDevToken firebase.FirestoreAppleDevJWT
	dataToErr := snapshot.DataTo(&appleDevToken)
	if dataToErr != nil {
		return nil, dataToErr
	}

	logger.Log("GetAppleDevToken", "Decoded data from Firebase")

	return &appleDevToken, nil
}

// createAppleMusicPlaylistRequest will generate the proper request needed for adding a playlist to Apple Music account
func createAppleMusicPlaylistRequest(playlistName string, endpoint string, apiToken string) (*http.Request, error) {
	logger.Log("CreateAppleMusicPlaylistRequest", "Starting...")

	// Create playlist data
	var appleMusicPlaylistReq appleMusicPlaylistRequest
	appleMusicPlaylistReq.Attributes.Name = "Grüvee: " + playlistName
	appleMusicPlaylistReq.Attributes.Description = "Created with love from Grüvee ❤️"

	// Create json body
	jsonPlaylist, jsonErr := json.Marshal(appleMusicPlaylistReq)
	if jsonErr != nil {
		return nil, jsonErr
	}

	logger.Log("CreateAppleMusicPlaylistRequest", "JSON body created successfully.")

	// Create request object
	createPlaylistReq, createPlaylistReqErr := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPlaylist))
	if createPlaylistReqErr != nil {
		return nil, createPlaylistReqErr
	}

	logger.Log("CreateAppleMusicPlaylistRequest", "Request object created successfully")

	// Get Apple Developer Token
	devJWT, err := getAppleDevToken()
	if err != nil {
		return nil, err
	}

	logger.Log("CreateAppleMusicPlaylistRequest", "Received Apple Dev Token successfully.")

	// Add headers
	createPlaylistReq.Header.Add("Content-Type", "application/json")
	createPlaylistReq.Header.Add("Music-User-Token", apiToken)
	createPlaylistReq.Header.Add("Authorization", "Bearer "+devJWT.Token)

	return createPlaylistReq, nil
}

// createSpotifyPlaylistRequest will generate the proper request needed for adding a playlist to Spotify account
func createSpotifyPlaylistRequest(playlistName string, endpoint string, apiToken string) (*http.Request, error) {
	logger.Log("CreateSpotifyPlaylistRequest", "Starting...")

	// Create playlist data
	var spotifyPlaylistRequest = spotifyPlaylistRequest{
		Name:          "Grüvee: " + playlistName,
		Public:        true,
		Collaborative: false,
		Description:   "Created with love from Grüvee ❤️",
	}

	// Create json body
	jsonPlaylist, jsonErr := json.Marshal(spotifyPlaylistRequest)
	if jsonErr != nil {
		return nil, jsonErr
	}

	logger.Log("CreateSpotifyPlaylistRequest", "JSON body created")

	// Create request object
	createPlaylistReq, createPlaylistReqErr := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPlaylist))
	if createPlaylistReqErr != nil {
		return nil, createPlaylistReqErr
	}

	logger.Log("CreateSpotifyPlaylistRequest", "Request body created")

	// Add headers
	createPlaylistReq.Header.Add("Content-Type", "application/json")
	createPlaylistReq.Header.Add("Authorization", "Bearer "+apiToken)

	return createPlaylistReq, nil
}

// refreshToken takes all socialPlatforms and checks to see if their tokens need to be refreshed
func refreshToken(platform firebase.FirestoreSocialPlatform) (*social.RefreshTokensResponse, error) {
	logger.Log("RefreshToken", "Starting...")

	var refreshReq = social.TokenRefreshRequest{
		UID: platform.PlatformName + ":" + platform.ID,
	}

	jsonTokenRefresh, jsonErr := json.Marshal(refreshReq)
	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	logger.Log("RefreshToken", "JSON body created successfully")

	var tokenRefreshURI = hostname + "/socialTokenRefresh"
	tokenRefreshReq, tokenRefreshReqErr := http.NewRequest("POST", tokenRefreshURI, bytes.NewBuffer(jsonTokenRefresh))
	if tokenRefreshReqErr != nil {
		return nil, fmt.Errorf(tokenRefreshReqErr.Error())
	}

	logger.Log("RefreshToken", "Request created successfully")

	tokenRefreshReq.Header.Add("Content-Type", "application/json")
	tokenRefreshReq.Header.Add("User-Type", "Gruvee-Backend")
	refreshedTokensResp, httpErr := httpClient.Do(tokenRefreshReq)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	logger.Log("RefreshToken", "Successfully called Token Refresh")

	if refreshedTokensResp.StatusCode == http.StatusNoContent {
		logger.Log("RefreshToken", "Tokens did not need to be refreshed.")
		return nil, nil
	}

	// Receive payload that includes uid
	var refreshedTokens social.RefreshTokensResponse

	// Decode payload
	refreshedTokensErr := json.NewDecoder(refreshedTokensResp.Body).Decode(&refreshedTokens)
	if refreshedTokensErr != nil {
		return nil, fmt.Errorf(refreshedTokensErr.Error())
	}

	logger.Log("RefreshToken", "Decoded response payload successfully.")

	return &refreshedTokens, nil
}
