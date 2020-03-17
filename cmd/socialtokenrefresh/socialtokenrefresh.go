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
)

var httpClient *http.Client
var firestoreClient *firestore.Client
var spotifyRefreshTokenURI = "https://accounts.spotify.com/api/token"

func init() {
	// Set httpClient
	httpClient = &http.Client{}

	// Get Firestore Client
	client, err := firestore.NewClient(context.Background(), "gruvee-3b7c4")
	if err != nil {
		log.Printf("SocialTokenRefresh [Init Firestore]: %v", err)
		return
	}
	firestoreClient = client
	log.Println("SocialTokenRefreshRequest initialized")
}

// socialTokenRefreshRequest includes uid to grab all social platforms for user
type socialTokenRefreshRequest struct {
	UID string `json:"uid"`
}

// spotifyRefreshTokenResponse includes the object returned from a Spotify token refresh
type spotifyRefreshTokenResponse struct {
	APIToken  string `json:"access_token"`
	TokenType string `json:"token_type"`
	Scope     string `json:"scope"`
	ExpiresIn int    `json:"expires_in"`
}

// SocialTokenRefresh checks to see if we need to refresh current API tokens for social platforms
func SocialTokenRefresh(writer http.ResponseWriter, request *http.Request) {
	// Check to see if we have env variables
	if os.Getenv("SPOTIFY_CLIENTID") == "" || os.Getenv("SPOTIFY_SECRET") == "" {
		log.Printf("SocialTokenRefresh [Check Env Props]: PROPS NOT HERE.")
		return
	}

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

	// Run refresh token logic
	refreshTokens(*platsToRefresh)

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
func refreshTokens(socialPlatforms []firebase.FirestoreSocialPlatform) {
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
			fmt.Printf("%s access token is expired. Calling Refresh...\n", platform.PlatformName)
			spotifyRefreshToken, tokenActionErr := refreshTokenAction(platform)
			if tokenActionErr != nil {
				fmt.Println(tokenActionErr.Error())
				continue
			}

			// Set new token data in database
			writeTokenErr := writeToken(platform.ID, *spotifyRefreshToken)
			if writeTokenErr != nil {
				fmt.Println(writeTokenErr.Error())
				continue
			}
		}
	}
}

// refreshTokenAction will call the API per platform and return the data needed
func refreshTokenAction(platform firebase.FirestoreSocialPlatform) (*spotifyRefreshTokenResponse, error) {
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
	var refreshTokenResp spotifyRefreshTokenResponse
	refreshTokenDecodeErr := json.NewDecoder(customTokenResp.Body).Decode(&refreshTokenResp)
	if refreshTokenDecodeErr != nil {
		return nil, fmt.Errorf(refreshTokenDecodeErr.Error())
	}

	return &refreshTokenResp, nil
}

// writeToken will write the new APIToken object to the social platform document
func writeToken(platformID string, token spotifyRefreshTokenResponse) error {
	// Create New APIToken Object

	var expiredAtStr = time.Now().Add(time.Second * time.Duration(token.ExpiresIn))
	var apiToken = firebase.APIToken{
		CreatedAt: time.Now().Format(time.RFC3339),
		ExpiredAt: expiredAtStr.Format(time.RFC3339),
		ExpiresIn: token.ExpiresIn,
		Token:     token.APIToken,
	}

	// Write new APIToken, ExpiredAt, ExpiresIn, CreatedAt
	platformDoc := firestoreClient.Collection("social_platforms").Doc(platformID)
	if platformDoc == nil {
		return fmt.Errorf("platformId %s could not be found", platformID)
	}

	// Time to update
	_, writeErr := platformDoc.Update(context.Background(), []firestore.Update{{Path: "apiToken", Value: apiToken}})
	if writeErr != nil {
		return fmt.Errorf(writeErr.Error())
	}

	return nil
}
