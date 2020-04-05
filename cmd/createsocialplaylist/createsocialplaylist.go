package createsocialplaylist

// Dragonfleas - "bobby drop tables wuz here pog - Dragonfleas - Relevant XKCD" (03/23/20)
// HMigo - "EN LØK HAR FLERE LAG" (03/26/20)
import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

// createSocialPlaylistRequest includes the socialPlatform and playlist that will be added
type createSocialPlaylistRequest struct {
	SocialPlatform firebase.FirestoreSocialPlatform `json:"socialPlatform"`
	Playlist       firebase.FirestorePlaylist       `json:"playlist"`
}

// spotifyPlaylistRequest includes the payload needed to create a Spotify Playlist
type spotifyPlaylistRequest struct {
	Name          string `json:"name"`
	Public        bool   `json:"public"`
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
}

var httpClient *http.Client
var hostname string

func init() {
	// Set httpClient
	httpClient = &http.Client{}

	log.Println("CreateSocialPlaylist Initialized")
}

// ywnklme - "At least something in my life is social 😞" (03/23/20)
// CreateSocialPlaylist will take in a SocialPlatform and will go create a playlist on the social account itself
func CreateSocialPlaylist(writer http.ResponseWriter, request *http.Request) {
	// Initialize paths
	if os.Getenv("ENVIRONMENT") == "DEV" {
		hostname = "http://localhost:8080"
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		hostname = "https://us-central1-gruvee-3b7c4.cloudfunctions.net"
	}

	var socialPlaylistReq createSocialPlaylistRequest

	// Decode our object
	jsonDecodeErr := json.NewDecoder(request.Body).Decode(&socialPlaylistReq)
	if jsonDecodeErr != nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateSocialPlaylist [socialPlaylistReq Decoder]: %v", jsonDecodeErr)
		return
	}

	// Figure out what service we are going to create a playlist in
	var platformEndpoint string
	if socialPlaylistReq.SocialPlatform.PlatformName == "spotify" {
		platformEndpoint = "https://api.spotify.com/v1/users/" + socialPlaylistReq.SocialPlatform.ID + "/playlists"
	}

	// Check if API token needs refresh
	apiToken, apiTokenErr := refreshToken(socialPlaylistReq.SocialPlatform)
	if apiTokenErr != nil {
		http.Error(writer, apiTokenErr.Error(), http.StatusBadRequest)
		log.Printf("CreateSocialPlaylist [refreshToken]: %v", apiTokenErr)
		return
	}

	if apiToken != nil {
		log.Println("Setting new APIToken on socialPlatform")
		socialPlaylistReq.SocialPlatform.APIToken.Token = *apiToken
	}

	// Call API to create playlist with data
	createReqErr := createPlaylist(platformEndpoint, socialPlaylistReq.SocialPlatform, socialPlaylistReq.Playlist)
	if createReqErr != nil {
		http.Error(writer, createReqErr.Error(), http.StatusBadRequest)
		log.Printf("CreateSocialPlaylist [createPlaylist]: %v", createReqErr)
		return
	}
}

// createPlaylist takes the social platform and playlist information and creates a playlist on the user's preferred platform
func createPlaylist(endpoint string, platform firebase.FirestoreSocialPlatform,
	playlist firebase.FirestorePlaylist) error {
	var spotifyPlaylistRequest = spotifyPlaylistRequest{
		Name:          "Grüvee: " + playlist.Name,
		Public:        true,
		Collaborative: false,
		Description:   "Created with love from Grüvee ❤️",
	}

	// Create jsonBody
	jsonPlaylist, jsonErr := json.Marshal(spotifyPlaylistRequest)
	if jsonErr != nil {
		return fmt.Errorf(jsonErr.Error())
	}

	createPlaylistReq, createPlaylistReqErr := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPlaylist))
	if createPlaylistReqErr != nil {
		return fmt.Errorf(createPlaylistReqErr.Error())
	}

	createPlaylistReq.Header.Add("Content-Type", "application/json")
	createPlaylistReq.Header.Add("Authorization", "Bearer "+platform.APIToken.Token)
	customTokenResp, httpErr := httpClient.Do(createPlaylistReq)
	if httpErr != nil {
		return fmt.Errorf(httpErr.Error())
	}

	if customTokenResp.StatusCode != http.StatusOK && customTokenResp.StatusCode != http.StatusCreated {
		// Convert Spotify Error Object
		var spotifyErrorObj social.SpotifyRequestError

		err := json.NewDecoder(customTokenResp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		return fmt.Errorf(spotifyErrorObj.Error.Message, spotifyErrorObj.Error.Status)
	}

	return nil
}

func refreshToken(platform firebase.FirestoreSocialPlatform) (*string, error) {
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

	// Get token for specified platform
	platformRefreshToken, doesExist := refreshedTokens.RefreshTokens[platform.PlatformName]
	if doesExist == false {
		// Another token needed refresh, but not the one we were looking for
		log.Printf("%s was not refreshed", platform.PlatformName)
		return nil, nil
	}

	return &platformRefreshToken.APIToken, nil
}
