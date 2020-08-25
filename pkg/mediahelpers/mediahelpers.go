package mediahelpers

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
	"github.com/pixelogicdev/gruveebackend/pkg/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// -- CONSTANTS -- //
var (
	spotifyAccessTokenURI = "https://accounts.spotify.com/api/token"
	httpClient            = &http.Client{}
)

// -- TYPES -- //

// SpotifyTrackData includes the reponse from a track in Spotify
type SpotifyTrackData struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Artists      []SpotifyArtistData `json:"artists"`
	Type         string              `json:"type"`
	Album        SpotifyAlbumData    `json:"album"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// SpotifyArtistData includes the response from an artist in Spotify
type SpotifyArtistData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// SpotifyAlbumData includes the response from an album in Spotify
type SpotifyAlbumData struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Type         string                  `json:"type"`
	Artists      []SpotifyArtistData     `json:"artists"`
	Images       []firebase.SpotifyImage `json:"images"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// -- HELPERS -- //

// GetMediaFromFirestore takes in a firestoreClient,  mediaProvider and the mediaId and checks to see if the media exists
func GetMediaFromFirestore(firestoreClient firestore.Client, mediaProvider string, mediaID string) (*firebase.FirestoreMedia, error) {
	var queryString string

	switch mediaProvider {
	case "apple":
		log.Println("[GetMediaFromFirestore] Apple provider found.")
		queryString = "apple.id"
	case "spotify":
		log.Println("[GetMediaFromFirestore] Spotify provider found.")
		queryString = "spotify.id"
	default:
		return nil, fmt.Errorf("[GetMediaFromFirestore] Provider passed in was not found: %s", mediaProvider)
	}

	// Construct query
	query := firestoreClient.Collection("songs").Where(queryString, "==", mediaID)

	// Execute & check result length
	results, resultsErr := query.Documents(context.Background()).GetAll()
	if resultsErr != nil {
		return nil, fmt.Errorf("[GetMediaFromFirestore] Error getting all documents from query: %v", resultsErr)
	}

	// Song does exist, return document
	if len(results) > 0 {
		log.Println("[GetMediaFromFirestore] Found song document")
		bytes, bytesErr := json.Marshal(results[0].Data())
		if bytesErr != nil {
			return nil, fmt.Errorf("[GetMediaFromFirestore] Cannot marshal FirestoreMedia data: %v", bytesErr)
		}

		var firestoreMedia firebase.FirestoreMedia
		unmarhsalErr := json.Unmarshal(bytes, &firestoreMedia)
		if unmarhsalErr != nil {
			return nil, fmt.Errorf("[GetMediaFromFirestore] Cannot unmarshal bytes data: %v", unmarhsalErr)
		}

		return &firestoreMedia, nil
	}

	// Song does not exist, need to create new song document
	log.Println("[GetMediaFromFirestore] Did not find song document")
	return nil, nil
}

// FetchSpotifyAuthToken takes in a firestoreClient and returns a SpotifyAuthToken
func FetchSpotifyAuthToken(firestoreClient firestore.Client) (*firebase.FirestoreSpotifyAuthToken, error) {
	// Go to Firebase and see if spotifyAuthToken exists
	snapshot, snapshotErr := firestoreClient.Collection("internal_tokens").Doc("spotifyAuthToken").Get(context.Background())
	if status.Code(snapshotErr) == codes.NotFound {
		log.Println("[GetSpotifyMedia] SpotifyAuthToken not found in DB. Need to create.")
		return nil, nil
	}

	if snapshotErr != nil {
		return nil, fmt.Errorf(snapshotErr.Error())
	}

	var spotifyAuthToken firebase.FirestoreSpotifyAuthToken
	dataToErr := snapshot.DataTo(&spotifyAuthToken)
	if dataToErr != nil {
		return nil, fmt.Errorf(dataToErr.Error())
	}

	log.Println("Checking to see if token needs to be refreshed")

	latestCreds, refreshAuthTokenErr := refreshSpotifyAuthToken(spotifyAuthToken, firestoreClient)
	if refreshAuthTokenErr != nil {
		return nil, refreshAuthTokenErr
	}

	if latestCreds != nil {
		return latestCreds, nil
	}

	// If we are here, no auth token was found
	newAuthToken, newAuthTokenErr := generateAuthToken()
	if newAuthTokenErr != nil {
		return nil, newAuthTokenErr
	}

	// Store new token in DB
	writeSpotifyAuthErr := writeSpotifyAuthtoken(*newAuthToken, firestoreClient)
	if writeSpotifyAuthErr != nil {
		return nil, writeSpotifyAuthErr
	}

	// TheYagich01: "Gejnerated" (08/11/20)
	log.Println("Generated auth token")
	return newAuthToken, nil
}

// refreshSpotifyAuthToken will check to see if Spotify AuthToken needs to be refreshed
func refreshSpotifyAuthToken(authToken firebase.FirestoreSpotifyAuthToken, firestoreClient firestore.Client) (*firebase.FirestoreSpotifyAuthToken, error) {
	// Get current time
	var currentTime = time.Now()

	fmt.Printf("Expires In: %d seconds\n", authToken.ExpiresIn)
	fmt.Printf("Expires At: %s seconds\n", authToken.ExpiredAt)

	expiredAtTime, expiredAtTimeErr := time.Parse(time.RFC3339, authToken.ExpiredAt)
	if expiredAtTimeErr != nil {
		return nil, fmt.Errorf(expiredAtTimeErr.Error())
	}

	if currentTime.After(expiredAtTime) {
		fmt.Println("spotifyAuthToken is expired. Refreshing...")
		refreshToken, tokenActionErr := generateAuthToken()
		if tokenActionErr != nil {
			return nil, fmt.Errorf(tokenActionErr.Error())
		}

		// Set new token data in database
		writeTokenErr := writeSpotifyAuthtoken(*refreshToken, firestoreClient)
		if writeTokenErr != nil {
			return nil, fmt.Errorf(writeTokenErr.Error())
		}

		return refreshToken, nil
	}

	// Nothing is expired just return original token
	return &authToken, nil
}

// generateAuthToken will call Spotify creds service to get authToken for making requests
func generateAuthToken() (*firebase.FirestoreSpotifyAuthToken, error) {
	// If not there generate new one and store
	authStr := os.Getenv("SPOTIFY_CLIENTID") + ":" + os.Getenv("SPOTIFY_SECRET")

	// Create Request
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	accessTokenReq, accessTokenReqErr := http.NewRequest("POST", spotifyAccessTokenURI,
		strings.NewReader(data.Encode()))
	if accessTokenReqErr != nil {
		return nil, fmt.Errorf(accessTokenReqErr.Error())
	}

	accessTokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	accessTokenReq.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(authStr)))
	accessTokenResp, accessTokenRespErr := httpClient.Do(accessTokenReq)
	if accessTokenRespErr != nil {
		return nil, fmt.Errorf(accessTokenRespErr.Error())
	}

	// Decode the token
	var spotifyClientCredsAuthResp social.SpotifyClientCredsAuthResp
	refreshTokenDecodeErr := json.NewDecoder(accessTokenResp.Body).Decode(&spotifyClientCredsAuthResp)
	if refreshTokenDecodeErr != nil {
		return nil, fmt.Errorf(refreshTokenDecodeErr.Error())
	}

	// Generate authToken for DB
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(time.Second * time.Duration(spotifyClientCredsAuthResp.ExpiresIn))
	authToken := firebase.FirestoreSpotifyAuthToken{
		IssuedAt:  issuedAt.Format(time.RFC3339),
		ExpiresIn: spotifyClientCredsAuthResp.ExpiresIn,
		ExpiredAt: expiresAt.Format(time.RFC3339),
		Token:     spotifyClientCredsAuthResp.AccessToken,
	}

	return &authToken, nil
}

// writeSpotifyAuthtoken will take spotifyAuthToken and write it to the internal_tokens collection
func writeSpotifyAuthtoken(authToken firebase.FirestoreSpotifyAuthToken, firestoreClient firestore.Client) error {
	spotifyAuthTokenDoc := firestoreClient.Collection("internal_tokens").Doc("spotifyAuthToken")
	if spotifyAuthTokenDoc == nil {
		return fmt.Errorf("spotifyAuthTokenDoc could not be found")
	}

	// MergeAll doesn't allow custom Go types so we need to create a map
	authMap := map[string]interface{}{
		"expiredAt": authToken.ExpiredAt,
		"expiresIn": authToken.ExpiresIn,
		"issuedAt":  authToken.IssuedAt,
		"token":     authToken.Token,
	}

	_, writeErr := spotifyAuthTokenDoc.Set(context.Background(), authMap, firestore.MergeAll)
	if writeErr != nil {
		return fmt.Errorf(writeErr.Error())
	}

	return nil
}
