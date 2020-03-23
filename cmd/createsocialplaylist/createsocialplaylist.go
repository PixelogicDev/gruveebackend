package createsocialplaylist

// Dragonfleas - "bobby drop tables wuz here pog - Dragonfleas - Relevant XKCD" (03/23/20)
import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
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

func init() {
	// Set httpClient
	httpClient = &http.Client{}

	log.Println("CreateSocialPlaylist Initialized")
}

// ywnklme - "At least something in my life is social 😞" (03/23/20)
// CreateSocialPlaylist will take in a SocialPlatform and will go create a playlist on the social account itself
func CreateSocialPlaylist(writer http.ResponseWriter, request *http.Request) {
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

	// TODO: Check if API token needs refresh

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
		var spotifyErrorObj firebase.SpotifyRequestError

		err := json.NewDecoder(customTokenResp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		return fmt.Errorf(spotifyErrorObj.Error.Message, spotifyErrorObj.Error.Status)
	}

	return nil
}

func refreshToken(platform firebase.FirestoreSocialPlatform) firebase.APIToken {
	// Create jsonBody
	// jsonPlaylist, jsonErr := json.Marshal(spotifyPlaylistRequest)
	// if jsonErr != nil {
	// 	return fmt.Errorf(jsonErr.Error())
	// }

	// createPlaylistReq, createPlaylistReqErr := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPlaylist))
	// if createPlaylistReqErr != nil {
	// 	return fmt.Errorf(createPlaylistReqErr.Error())
	// }

	// createPlaylistReq.Header.Add("Content-Type", "application/json")
	// createPlaylistReq.Header.Add("Authorization", "Bearer "+platform.APIToken.Token)
	// customTokenResp, httpErr := httpClient.Do(createPlaylistReq)
	// if httpErr != nil {
	// 	return fmt.Errorf(httpErr.Error())
	// }
}
