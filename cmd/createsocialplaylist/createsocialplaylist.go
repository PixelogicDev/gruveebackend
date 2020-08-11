package createsocialplaylist

// Dragonfleas - "bobby drop tables wuz here pog - Dragonfleas - Relevant XKCD" (03/23/20)
// HMigo - "EN LØK HAR FLERE LAG" (03/26/20)
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

// createSocialPlaylistRequest includes the socialPlatform and playlist that will be added
type createSocialPlaylistRequest struct {
	SocialPlatform firebase.FirestoreSocialPlatform `json:"socialPlatform"`
	PlaylistName   string                           `json:"playlistName"`
}

// createSocialPlaylistResponse includes the refreshToken for the platform if there is one
type createSocialPlaylistResponse struct {
	PlatformName string            `json:"platformName"`
	RefreshToken firebase.APIToken `json:"refreshToken"`
}

// appleMusicPlaylistRequest includes the payload needed to create an Apple Music Playlist
type appleMusicPlaylistRequest struct {
	Attributes struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"attributes"`
}

// spotifyPlaylistRequest includes the payload needed to create a Spotify Playlist
type spotifyPlaylistRequest struct {
	Name          string `json:"name"`
	Public        bool   `json:"public"`
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
}

var firestoreClient *firestore.Client
var logger sawmill.Logger
var httpClient *http.Client
var hostname string

// ywnklme - "At least something in my life is social 😞" (03/23/20)
func init() {
	// Set httpClient
	httpClient = &http.Client{}

	log.Println("CreateSocialPlaylist Initialized")
}

// CreateSocialPlaylist will take in a SocialPlatform and will go create a playlist on the social account itself
func CreateSocialPlaylist(writer http.ResponseWriter, request *http.Request) {
	// Initialize paths
	err := initWithEnv()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		logger.LogErr(err, "initWithEnv", nil)
		return
	}

	var socialPlaylistReq createSocialPlaylistRequest

	// Decode our object
	jsonDecodeErr := json.NewDecoder(request.Body).Decode(&socialPlaylistReq)
	if jsonDecodeErr != nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		logger.LogErr(jsonDecodeErr, "socialPlaylistReq Decoder", request)
		return
	}

	// Figure out what service we are going to create a playlist in
	var platformEndpoint string
	var socialRefreshTokens *social.RefreshTokensResponse
	var socialRefreshTokenErr error

	if socialPlaylistReq.SocialPlatform.PlatformName == "spotify" {
		log.Printf("Creating playlist for Spotify")
		platformEndpoint = "https://api.spotify.com/v1/users/" + socialPlaylistReq.SocialPlatform.ID + "/playlists"

		// This is sort of weird, but I haven't been able to find any resources on an Apple Music tokens expiring
		// Therefore, this check should only be done on Spotify at the moment
		socialRefreshTokens, socialRefreshTokenErr = refreshToken(socialPlaylistReq.SocialPlatform)
		if socialRefreshTokenErr != nil {
			http.Error(writer, socialRefreshTokenErr.Error(), http.StatusBadRequest)
			logger.LogErr(socialRefreshTokenErr, "refreshToken", request)
			return
		}
	} else if socialPlaylistReq.SocialPlatform.PlatformName == "apple" {
		log.Printf("Creating playlist for Apple Music")
		platformEndpoint = "https://api.music.apple.com/v1/me/library/playlists"
	}

	// fr3fou - "i fixed this Kappa" (04/10/20)
	// Setup resonse if we have a token to return
	var response *createSocialPlaylistResponse

	// Again, this is solely for Spotify at the moment
	if socialPlaylistReq.SocialPlatform.PlatformName == "spotify" && socialRefreshTokens != nil {
		// Get token for specified platform
		platformRefreshToken, doesExist := socialRefreshTokens.RefreshTokens[socialPlaylistReq.SocialPlatform.PlatformName]
		if doesExist == true {
			log.Println("Setting new APIToken on socialPlatform")
			socialPlaylistReq.SocialPlatform.APIToken.Token = platformRefreshToken.Token

			// Write new apiToken as response
			response = &createSocialPlaylistResponse{
				PlatformName: socialPlaylistReq.SocialPlatform.PlatformName,
				RefreshToken: platformRefreshToken,
			}
		} else {
			// Another token needed refresh, but not the one we were looking for
			log.Printf("%s was not refreshed", socialPlaylistReq.SocialPlatform.PlatformName)
		}
	}

	// Call API to create playlist with data
	createReqErr := createPlaylist(platformEndpoint, socialPlaylistReq.SocialPlatform, socialPlaylistReq.PlaylistName)
	if createReqErr != nil {
		http.Error(writer, createReqErr.Error(), http.StatusBadRequest)
		logger.LogErr(createReqErr, "createPlaylist", request)
		return
	}

	if response != nil {
		json.NewEncoder(writer).Encode(response)
	} else {
		writer.WriteHeader(http.StatusNoContent)
	}
}

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

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("SocialTokenRefresh [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), "NOT DEV", "CreateSocialPlaylist")
	if err != nil {
		log.Printf("CreateSocial Playlist [Init Sawmill]: %v", err)
	}

	// DR_DinoMight - "Note to self! Welcome, Dr_DinoMight, Otherwise he'll spit his dummy out!" (06.15.20)
	firestoreClient = client
	logger = sawmillLogger
	return nil
}

// createPlaylist takes the social platform and playlist information and creates a playlist on the user's preferred platform
func createPlaylist(endpoint string, platform firebase.FirestoreSocialPlatform,
	playlistName string) error {
	var request *http.Request
	var requestErr error

	// Check for platform
	if platform.PlatformName == "spotify" {
		request, requestErr = createSpotifyPlaylistRequest(playlistName, endpoint, platform.APIToken.Token)
	} else if platform.PlatformName == "apple" {
		request, requestErr = createAppleMusicPlaylistRequest(playlistName, endpoint, platform.APIToken.Token)
	}

	if requestErr != nil {
		log.Printf("[createPlaylist] %v", requestErr.Error())
		return requestErr
	}

	createPlaylistResp, httpErr := httpClient.Do(request)
	if httpErr != nil {
		log.Printf("[createPlaylist] %v", httpErr.Error())
		return httpErr
	}

	// If we have errors, lets parse 'em out
	if createPlaylistResp.StatusCode != http.StatusOK && createPlaylistResp.StatusCode != http.StatusCreated {
		if platform.PlatformName == "spotify" {
			var spotifyErrorObj social.SpotifyRequestError

			err := json.NewDecoder(createPlaylistResp.Body).Decode(&spotifyErrorObj)
			if err != nil {
				log.Printf("[createPlaylist] %v", err.Error())
				return err
			}

			return fmt.Errorf("Status Code %v: "+spotifyErrorObj.Error.Message, spotifyErrorObj.Error.Status)
		} else if platform.PlatformName == "apple" {
			var appleMusicReqErr social.AppleMusicRequestError

			err := json.NewDecoder(createPlaylistResp.Body).Decode(&appleMusicReqErr)
			if err != nil {
				log.Printf("[createPlaylist] %v", err.Error())
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
	// Go to Firebase and see if appleDevToken exists
	snapshot, snapshotErr := firestoreClient.Collection("internal_tokens").Doc("appleDevToken").Get(context.Background())
	if status.Code(snapshotErr) == codes.NotFound {
		log.Println("[getAppleDevToken] AppleDevToken not found in DB.")
		return nil, nil
	}

	if snapshotErr != nil {
		log.Printf("[getAppleDevToken] %v", snapshotErr.Error())
		return nil, snapshotErr
	}

	var appleDevToken firebase.FirestoreAppleDevJWT
	dataToErr := snapshot.DataTo(&appleDevToken)
	if dataToErr != nil {
		log.Printf("[getAppleDevToken] %v", dataToErr.Error())
		return nil, dataToErr
	}

	return &appleDevToken, nil
}

// createAppleMusicPlaylistRequest will generate the proper request needed for adding a playlist to Apple Music account
func createAppleMusicPlaylistRequest(playlistName string, endpoint string, apiToken string) (*http.Request, error) {
	// Create playlist data
	var appleMusicPlaylistReq appleMusicPlaylistRequest
	appleMusicPlaylistReq.Attributes.Name = "Grüvee: " + playlistName
	appleMusicPlaylistReq.Attributes.Description = "Created with love from Grüvee ❤️"

	// Create json body
	jsonPlaylist, jsonErr := json.Marshal(appleMusicPlaylistReq)
	if jsonErr != nil {
		log.Printf("[createAppleMusicPlaylistRequest] %v", jsonErr.Error())
		return nil, jsonErr
	}

	// Create request object
	createPlaylistReq, createPlaylistReqErr := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPlaylist))
	if createPlaylistReqErr != nil {
		log.Printf("[createAppleMusicPlaylistRequest] %v", createPlaylistReqErr.Error())
		return nil, createPlaylistReqErr
	}

	// Get Apple Developer Token
	devJWT, err := getAppleDevToken()
	if err != nil {
		return nil, err
	}

	// Add headers
	createPlaylistReq.Header.Add("Content-Type", "application/json")
	createPlaylistReq.Header.Add("Music-User-Token", apiToken)
	createPlaylistReq.Header.Add("Authorization", "Bearer "+devJWT.Token)

	return createPlaylistReq, nil
}

// createSpotifyPlaylistRequest will generate the proper request needed for adding a playlist to Spotify account
func createSpotifyPlaylistRequest(playlistName string, endpoint string, apiToken string) (*http.Request, error) {
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

	// Create request object
	createPlaylistReq, createPlaylistReqErr := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPlaylist))
	if createPlaylistReqErr != nil {
		return nil, createPlaylistReqErr
	}

	// Add headers
	createPlaylistReq.Header.Add("Content-Type", "application/json")
	createPlaylistReq.Header.Add("Authorization", "Bearer "+apiToken)

	return createPlaylistReq, nil
}

// refreshToken takes all socialPlatforms and checks to see if their tokens need to be refreshed
func refreshToken(platform firebase.FirestoreSocialPlatform) (*social.RefreshTokensResponse, error) {
	var refreshReq = social.TokenRefreshRequest{
		UID: platform.PlatformName + ":" + platform.ID,
	}

	var tokenRefreshURI = hostname + "/socialTokenRefresh"
	jsonTokenRefresh, jsonErr := json.Marshal(refreshReq)
	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	tokenRefreshReq, tokenRefreshReqErr := http.NewRequest("POST", tokenRefreshURI, bytes.NewBuffer(jsonTokenRefresh))
	if tokenRefreshReqErr != nil {
		return nil, fmt.Errorf(tokenRefreshReqErr.Error())
	}

	tokenRefreshReq.Header.Add("Content-Type", "application/json")
	tokenRefreshReq.Header.Add("User-Type", "Gruvee-Backend")
	refreshedTokensResp, httpErr := httpClient.Do(tokenRefreshReq)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	if refreshedTokensResp.StatusCode == http.StatusNoContent {
		log.Println("Tokens did not need refresh")
		return nil, nil
	}

	// Receive payload that includes uid
	var refreshedTokens social.RefreshTokensResponse

	// Decode payload
	refreshedTokensErr := json.NewDecoder(refreshedTokensResp.Body).Decode(&refreshedTokens)
	if refreshedTokensErr != nil {
		return nil, fmt.Errorf(refreshedTokensErr.Error())
	}

	return &refreshedTokens, nil
}
