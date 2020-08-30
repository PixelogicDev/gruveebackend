package createsocialplaylist

// Dragonfleas - "bobby drop tables wuz here pog - Dragonfleas - Relevant XKCD" (03/23/20)
// HMigo - "EN LØK HAR FLERE LAG" (03/26/20)
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

var (
	firestoreClient *firestore.Client
	logger          sawmill.Logger
	httpClient      *http.Client
	hostname        string
)

// ywnklme - "At least something in my life is social 😞" (03/23/20)
func init() {
	log.Println("CreateSocialPlaylist Initialized")
}

// CreateSocialPlaylist will take in a SocialPlatform and will go create a playlist on the social account itself
func CreateSocialPlaylist(writer http.ResponseWriter, request *http.Request) {
	// Initialize paths
	err := initWithEnv()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", err, nil)
		return
	}

	var socialPlaylistReq createSocialPlaylistRequest

	// Decode our object
	jsonDecodeErr := json.NewDecoder(request.Body).Decode(&socialPlaylistReq)
	if jsonDecodeErr != nil {
		http.Error(writer, jsonDecodeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("SocialPlaylistReq Decoder", jsonDecodeErr, request)
		return
	}

	logger.Log("CreateSocialPlaylist", "Decoded request")

	// Figure out what service we are going to create a playlist in
	var platformEndpoint string
	var socialRefreshTokens *social.RefreshTokensResponse
	var socialRefreshTokenErr error

	if socialPlaylistReq.SocialPlatform.PlatformName == "spotify" {
		logger.Log("CreateSocialPlaylist", "Creating Spotify Playlist")
		platformEndpoint = "https://api.spotify.com/v1/users/" + socialPlaylistReq.SocialPlatform.ID + "/playlists"

		// This is sort of weird, but I haven't been able to find any resources on an Apple Music tokens expiring
		// Therefore, this check should only be done on Spotify at the moment
		socialRefreshTokens, socialRefreshTokenErr = refreshToken(socialPlaylistReq.SocialPlatform)
		if socialRefreshTokenErr != nil {
			http.Error(writer, socialRefreshTokenErr.Error(), http.StatusBadRequest)
			logger.LogErr("RefreshToken", socialRefreshTokenErr, request)
			return
		}

		logger.Log("RefreshToken", "Succesfully refreshed tokens.")
	}

	if socialPlaylistReq.SocialPlatform.PlatformName == "apple" {
		logger.Log("CreateSocialPlaylist", "Creating Apple Music Playlist")
		platformEndpoint = "https://api.music.apple.com/v1/me/library/playlists"
	}

	if socialPlaylistReq.SocialPlatform.PlatformName == "youtube" {
		logger.Log("CreateSocialPlaylist", "Creating YouTube Music Playlist")
	}

	// fr3fou - "i fixed this Kappa" (04/10/20)
	// Setup resonse if we have a token to return
	var response *createSocialPlaylistResponse

	// Again, this is solely for Spotify at the moment
	if socialPlaylistReq.SocialPlatform.PlatformName == "spotify" && socialRefreshTokens != nil {
		logger.Log("CreateSocialPlaylist", "Spotify token was refreshed for user.")

		// Get token for specified platform
		platformRefreshToken, doesExist := socialRefreshTokens.RefreshTokens[socialPlaylistReq.SocialPlatform.PlatformName]
		if doesExist == true {
			logger.Log("CreateSocialPlaylist", "Setting new APIToken on socialPlatform")
			socialPlaylistReq.SocialPlatform.APIToken.Token = platformRefreshToken.Token

			// Write new apiToken as response
			response = &createSocialPlaylistResponse{
				PlatformName: socialPlaylistReq.SocialPlatform.PlatformName,
				RefreshToken: platformRefreshToken,
			}
		} else {
			// Another token needed refresh, but not the one we were looking for
			logger.Log("CreateSocialPlaylist", fmt.Sprintf("%s was not refreshed", socialPlaylistReq.SocialPlatform.PlatformName))
			log.Printf("%s was not refreshed", socialPlaylistReq.SocialPlatform.PlatformName)
		}
	}

	// Call API to create playlist with data
	createReqErr := createPlaylist(platformEndpoint, socialPlaylistReq.SocialPlatform, socialPlaylistReq.PlaylistName)
	if createReqErr != nil {
		http.Error(writer, createReqErr.Error(), http.StatusBadRequest)
		logger.LogErr("CreatePlaylist", createReqErr, request)
		return
	}

	logger.Log("CreateSocialPlaylist", "CreatePlaylist call was successful.")

	// If a new token was generated, send back to the client
	if response != nil {
		json.NewEncoder(writer).Encode(response)
	} else {
		writer.WriteHeader(http.StatusNoContent)
	}
}
