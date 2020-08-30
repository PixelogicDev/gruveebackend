package socialtokenrefresh

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

const spotifyRefreshTokenURI = "https://accounts.spotify.com/api/token"

var (
	httpClient      *http.Client
	firestoreClient *firestore.Client
	logger          sawmill.Logger
)

func init() {
	log.Println("SocialTokenRefresh initialized")
}

// SocialTokenRefresh checks to see if we need to refresh current API tokens for social platforms
func SocialTokenRefresh(writer http.ResponseWriter, request *http.Request) {
	// Check to see if we have env variables
	if os.Getenv("SPOTIFY_CLIENTID") == "" || os.Getenv("SPOTIFY_SECRET") == "" {
		logger.LogErr("SocialTokenRefresh", fmt.Errorf("Spotify ClientID not found in environment env"), nil)
		return
	}

	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initWithEnvErr, nil)
		return
	}

	// Receive payload that includes uid
	var socialTokenReq social.TokenRefreshRequest

	// Decode payload
	socialTokenErr := json.NewDecoder(request.Body).Decode(&socialTokenReq)
	if socialTokenErr != nil {
		http.Error(writer, socialTokenErr.Error(), http.StatusInternalServerError)
		logger.LogErr("SocialTokenReq Decoder", socialTokenErr, request)
		return
	}

	logger.Log("SocialTokenRefresh", "Decoded request")

	// Go to Firestore and get the platforms for user
	platsToRefresh, platformErr := getUserPlatformsToRefresh(socialTokenReq.UID)
	if platformErr != nil {
		http.Error(writer, platformErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GetUserPlatforms", platformErr, request)
		return
	}

	logger.Log("SocialTokenRefresh", "Received SocialPlatform.")

	if platsToRefresh != nil && len(*platsToRefresh) == 0 {
		logger.Log("SocialTokenRefresh", "No social platforms need to be refreshed.")

		// No refresh needed, lets return this with no content
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	// Run refresh token logic
	refreshTokenResp := refreshTokens(*platsToRefresh)

	logger.Log("SocialTokenRefresh", "Tokens successfully refreshed.")

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(refreshTokenResp)
}
