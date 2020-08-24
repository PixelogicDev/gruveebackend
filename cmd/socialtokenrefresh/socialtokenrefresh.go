package socialtokenrefresh

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

var httpClient *http.Client
var firestoreClient *firestore.Client
var logger sawmill.Logger
var spotifyRefreshTokenURI = "https://accounts.spotify.com/api/token"

func init() {
	// Set httpClient
	httpClient = &http.Client{}

	log.Println("SocialTokenRefresh initialized")
}

// spotifyRefreshTokenRes contains the response from Spotify when trying to refresh the access token
type spotifyRefreshTokenRes struct {
	PlatformName string `json:"platformName"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
}

// SocialTokenRefresh checks to see if we need to refresh current API tokens for social platforms
func SocialTokenRefresh(writer http.ResponseWriter, request *http.Request) {
	// Check to see if we have env variables
	if os.Getenv("SPOTIFY_CLIENTID") == "" || os.Getenv("SPOTIFY_SECRET") == "" {
		log.Fatalln("SocialTokenRefresh [Check Env Props]: PROPS NOT HERE.")
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

	// Go to Firestore and get the platforms for user
	platsToRefresh, platformErr := getUserPlatformsToRefresh(socialTokenReq.UID)
	if platformErr != nil {
		http.Error(writer, platformErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GetUserPlatforms", platformErr, request)
		return
	}

	if platsToRefresh != nil && len(*platsToRefresh) == 0 {
		// No refresh needed, lets return this with no content
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	// Run refresh token logic
	refreshTokenResp := refreshTokens(*platsToRefresh)

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(refreshTokenResp)
}

// Helpers
// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("SocialTokenRefresh [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "SocialTokenRefersh")
	if err != nil {
		log.Printf("SocialTokenRefresh [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}

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

	// WE SHOULD BE: checking to see which platforms need to be refreshed
	// Currently we are not, just returning all platforms
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
func refreshTokens(socialPlatforms []firebase.FirestoreSocialPlatform) social.RefreshTokensResponse {
	// Get current time
	var currentTime = time.Now()
	var refreshTokensResp = social.RefreshTokensResponse{
		RefreshTokens: map[string]firebase.APIToken{},
	}

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
			fmt.Printf("%s access token is expired. Calling Refresh...\n", platform.PlatformName)
			refreshToken, tokenActionErr := refreshTokenAction(platform)
			if tokenActionErr != nil {
				fmt.Println(tokenActionErr.Error())
				continue
			}

			var expiredAtStr = time.Now().Add(time.Second * time.Duration(refreshToken.ExpiresIn))
			var refreshedAPIToken = firebase.APIToken{
				CreatedAt: time.Now().Format(time.RFC3339),
				ExpiredAt: expiredAtStr.Format(time.RFC3339),
				ExpiresIn: refreshToken.ExpiresIn,
				Token:     refreshToken.AccessToken,
			}

			// Set refresh token in map
			refreshTokensResp.RefreshTokens[platform.PlatformName] = refreshedAPIToken

			// Set new token data in database
			writeTokenErr := writeToken(platform.ID, refreshedAPIToken)
			if writeTokenErr != nil {
				fmt.Println(writeTokenErr.Error())
				continue
			}
		}
	}

	return refreshTokensResp
}

// refreshTokenAction will call the API per platform and return the data needed
func refreshTokenAction(platform firebase.FirestoreSocialPlatform) (*spotifyRefreshTokenRes, error) {
	var authStr = os.Getenv("SPOTIFY_CLIENTID") + ":" + os.Getenv("SPOTIFY_SECRET")

	// Create Request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", platform.RefreshToken)

	refreshTokenReq, refreshTokenReqErr := http.NewRequest("POST", spotifyRefreshTokenURI,
		strings.NewReader(data.Encode()))
	if refreshTokenReqErr != nil {
		return nil, fmt.Errorf(refreshTokenReqErr.Error())
	}

	refreshTokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	refreshTokenReq.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(authStr)))
	customTokenResp, httpErr := httpClient.Do(refreshTokenReq)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	// Decode the token to send back
	var spotifyRefreshRes spotifyRefreshTokenRes
	refreshTokenDecodeErr := json.NewDecoder(customTokenResp.Body).Decode(&spotifyRefreshRes)
	if refreshTokenDecodeErr != nil {
		return nil, fmt.Errorf(refreshTokenDecodeErr.Error())
	}

	// Make sure to add platform name here before continuing
	spotifyRefreshRes.PlatformName = platform.PlatformName
	return &spotifyRefreshRes, nil
}

// writeToken will write the new APIToken object to the social platform document
func writeToken(platformID string, token firebase.APIToken) error {
	// Write new APIToken, ExpiredAt, ExpiresIn, CreatedAt
	platformDoc := firestoreClient.Collection("social_platforms").Doc(platformID)
	if platformDoc == nil {
		return fmt.Errorf("platformId %s could not be found", platformID)
	}

	// Time to update
	_, writeErr := platformDoc.Update(context.Background(), []firestore.Update{{Path: "apiToken", Value: token}})
	if writeErr != nil {
		return fmt.Errorf(writeErr.Error())
	}

	return nil
}
