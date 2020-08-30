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

	// Set httpClient
	httpClient = &http.Client{}

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
	logger.Log("GetUserPlatformsToRefresh", "Starting...")

	// Go to Firebase and get document references for all social platforms
	snapshot, snapshotErr := firestoreClient.Collection("users").Doc(uid).Get(context.Background())
	if snapshotErr != nil {
		return nil, fmt.Errorf(snapshotErr.Error())
	}

	logger.Log("GetUserPlatformsToRefresh", "Received snapshot.")

	// Grab socialPlatforms array
	var firestoreUser firebase.FirestoreUser
	dataToErr := snapshot.DataTo(&firestoreUser)
	if dataToErr != nil {
		return nil, fmt.Errorf(dataToErr.Error())
	}

	logger.Log("GetUserPlatformsToRefresh", "Received snapshot data.")

	// WE SHOULD BE: checking to see which platforms need to be refreshed
	// Currently we are not, just returning all platforms
	socialPlatforms, fetchRefErr := fetchChildRefs(firestoreUser.SocialPlatforms)
	if fetchRefErr != nil {
		return nil, fmt.Errorf(fetchRefErr.Error())
	}

	logger.Log("GetUserPlatformsToRefresh", "Successfully refreshed tokens.")

	// Return those platforms to main
	return socialPlatforms, nil
}

// fetchChildRefs will convert document references to FiresstoreSocilaPlatform Objects
func fetchChildRefs(refs []*firestore.DocumentRef) (*[]firebase.FirestoreSocialPlatform, error) {
	logger.Log("FetchChildRefs", "Starting...")

	docsnaps, err := firestoreClient.GetAll(context.Background(), refs)
	if err != nil {
		return nil, fmt.Errorf("fetchChildRefs: %v", err)
	}

	logger.Log("FetchChildRefs", "Received all child documents.")

	var socialPlatforms []firebase.FirestoreSocialPlatform
	for _, userSnap := range docsnaps {
		var socialPlatform firebase.FirestoreSocialPlatform

		dataErr := userSnap.DataTo(&socialPlatform)
		if dataErr != nil {
			logger.Log("FetchChildRefs", "Encountered error while parsing userSnapshot.")
			logger.LogErr("FetchChildRefs", dataErr, nil)
			continue
		}

		socialPlatforms = append(socialPlatforms, socialPlatform)
	}

	logger.Log("FetchChildRefs", "Successfully received social platforms.")

	return &socialPlatforms, nil
}

// refreshTokens goes through social platform objects and refreshes tokens as necessary
func refreshTokens(socialPlatforms []firebase.FirestoreSocialPlatform) social.RefreshTokensResponse {
	logger.Log("RefreshTokens", "Starting...")

	// Get current time
	var currentTime = time.Now()
	var refreshTokensResp = social.RefreshTokensResponse{
		RefreshTokens: map[string]firebase.APIToken{},
	}

	for _, platform := range socialPlatforms {
		logger.Log("RefreshTokens", fmt.Sprintf("Expires In: %d seconds\n", platform.APIToken.ExpiresIn))
		logger.Log("RefreshTokens", fmt.Sprintf("Expired At: %s\n", platform.APIToken.ExpiredAt))
		logger.Log("RefreshTokens", fmt.Sprintf("Created At: %s\n", platform.APIToken.CreatedAt))

		expiredAtTime, expiredAtTimeErr := time.Parse(time.RFC3339, platform.APIToken.ExpiredAt)
		if expiredAtTimeErr != nil {
			fmt.Println(expiredAtTimeErr.Error())
			continue
		}

		logger.Log("RefreshTokens", "Successfully parsed expiredAtTime")

		if currentTime.After(expiredAtTime) {
			logger.Log("RefreshTokens", fmt.Sprintf("%s access token is expired. Calling Refresh...\n", platform.PlatformName))

			// Call API refresh
			refreshToken, tokenActionErr := refreshTokenAction(platform)
			if tokenActionErr != nil {
				logger.LogErr("RefreshToken", tokenActionErr, nil)
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
				logger.LogErr("RefreshToken", writeTokenErr, nil)
				continue
			}
		}
	}

	logger.Log("RefreshToken", "Successfully refreshed all platforms.")

	return refreshTokensResp
}

// refreshTokenAction will call the API per platform and return the data needed
func refreshTokenAction(platform firebase.FirestoreSocialPlatform) (*spotifyRefreshTokenRes, error) {
	logger.Log("RefreshTokenAction", "Starting...")

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

	logger.Log("RefreshTokenAction", "Generated Request.")

	refreshTokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	refreshTokenReq.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(authStr)))
	customTokenResp, httpErr := httpClient.Do(refreshTokenReq)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	logger.Log("RefreshTokenAction", "Received CustomToken Response.")

	// Decode the token to send back
	var spotifyRefreshRes spotifyRefreshTokenRes
	refreshTokenDecodeErr := json.NewDecoder(customTokenResp.Body).Decode(&spotifyRefreshRes)
	if refreshTokenDecodeErr != nil {
		return nil, fmt.Errorf(refreshTokenDecodeErr.Error())
	}

	logger.Log("RefreshTokenAction", "Successfully decoded response.")

	// Make sure to add platform name here before continuing
	spotifyRefreshRes.PlatformName = platform.PlatformName
	return &spotifyRefreshRes, nil
}

// writeToken will write the new APIToken object to the social platform document
func writeToken(platformID string, token firebase.APIToken) error {
	logger.Log("WriteToken", "Starting...")

	// Write new APIToken, ExpiredAt, ExpiresIn, CreatedAt
	platformDoc := firestoreClient.Collection("social_platforms").Doc(platformID)
	if platformDoc == nil {
		return fmt.Errorf("platformId %s could not be found", platformID)
	}

	logger.Log("WriteToken", "Received document.")

	// Time to update
	_, writeErr := platformDoc.Update(context.Background(), []firestore.Update{{Path: "apiToken", Value: token}})
	if writeErr != nil {
		return fmt.Errorf(writeErr.Error())
	}

	logger.Log("WriteToken", "Successfully wrote document.")

	return nil
}
