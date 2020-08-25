package getspotifymedia

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

	// Initialize HttpClient
	httpClient = &http.Client{}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("CreateUser [Init Firestore]: %v", err)
	}

	firestoreClient = client
	return nil
}

// getApiToken checks to see if we have an APIToken for client credential calls
func getCreds() (*firebase.FirestoreSpotifyAuthToken, error) {
	// Check to see if we have env variables
	if os.Getenv("SPOTIFY_CLIENTID") == "" || os.Getenv("SPOTIFY_SECRET") == "" {
		log.Fatalln("GetSpotifyMedia [Check Env Props]: PROPS NOT HERE.")
		return nil, fmt.Errorf("getSpotifyMedia [Check Env Props]: PROPS NOT HERE")
	}

	// Check for Auth token in DB
	authToken, authTokenErr := fetchToken()
	if authTokenErr != nil {
		return nil, authTokenErr
	}

	// Check to see if auth token exists && needs to be refreshed
	if authToken != nil {
		log.Println("Checking to see if token needs to be refreshed")

		latestCreds, refreshAuthTokenErr := refreshAuthToken(*authToken)
		if refreshAuthTokenErr != nil {
			return nil, refreshAuthTokenErr
		}

		return latestCreds, nil
	}

	// If we are here, no auth token was found
	newAuthToken, newAuthTokenErr := generateAuthToken()
	if newAuthTokenErr != nil {
		return nil, newAuthTokenErr
	}

	// Store new token in DB
	writeSpotifyAuthErr := writeSpotifyAuthtoken(*newAuthToken)
	if writeSpotifyAuthErr != nil {
		return nil, writeSpotifyAuthErr
	}

	// TheYagich01: "Gejnerated" (08/11/20)
	log.Println("Generated auth token")
	return newAuthToken, nil
}

// fetchToken will grab the Apple Developer Token from DB
func fetchToken() (*firebase.FirestoreSpotifyAuthToken, error) {
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

	return &spotifyAuthToken, nil
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

// refreshAuthToken will check to see if Spotify AuthToken needs to be refreshed
func refreshAuthToken(authToken firebase.FirestoreSpotifyAuthToken) (*firebase.FirestoreSpotifyAuthToken, error) {
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
		writeTokenErr := writeSpotifyAuthtoken(*refreshToken)
		if writeTokenErr != nil {
			return nil, fmt.Errorf(writeTokenErr.Error())
		}

		return refreshToken, nil
	}

	// Nothing is expired just return original token
	return &authToken, nil
}

// writeSpotifyAuthtoken will take spotifyAuthToken and write it to the internal_tokens collection
func writeSpotifyAuthtoken(authToken firebase.FirestoreSpotifyAuthToken) error {
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

// getSpotifyTrack calls Spotify GET track API and converts to Golang Type
func getSpotifyTrack(trackID string, accessToken string) (*firebase.FirestoreMedia, error) {
	// Generate URI
	spotifyGetURI := spotifyGetTrackURI + "/" + trackID

	// Generate request
	spotifyTrackReq, spotifyTrackReqErr := http.NewRequest("GET", spotifyGetURI, nil)
	if spotifyTrackReqErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [http.NewRequest]: %v", spotifyTrackReqErr)
	}

	// Add headers and call request
	spotifyTrackReq.Header.Add("Authorization", "Bearer "+accessToken)
	getTrackResp, spotifyTrackRespErr := httpClient.Do(spotifyTrackReq)
	if spotifyTrackRespErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [client.Do]: %v", spotifyTrackRespErr)
	}

	// Check to see if request was valid
	if getTrackResp.StatusCode != http.StatusOK {
		// Convert Spotify Error Object
		var spotifyErrorObj social.SpotifyRequestError

		err := json.NewDecoder(getTrackResp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetSpotifyMedia [Spotify Request Decoder]: %v", err)
		}
		return nil, fmt.Errorf("GetSpotifyMedia [Spotify Track Request]: %v", spotifyErrorObj.Error.Message)
	}

	var spotifyTrackData spotifyTrackResp

	// syszen - "wait that it? #easyGo"(02/27/20)
	// LilCazza - "Why the fuck doesn't this shit work" (02/27/20)
	respDecodeErr := json.NewDecoder(getTrackResp.Body).Decode(&spotifyTrackData)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [Spotify Response Decoder]: %v", respDecodeErr)
	}

	// Setup FirestoreMeida object
	firestoreMedia := firebase.FirestoreMedia{
		Name:    spotifyTrackData.Name,
		Album:   spotifyTrackData.Album.Name,
		Type:    spotifyTrackData.Type,
		Creator: generateArtistsString(spotifyTrackData.Artists),
		Spotify: firebase.FirestoreMediaPlatformData{
			ID:     trackID,
			URL:    spotifyTrackData.ExternalURLs.Spotify,
			Images: spotifyTrackData.Album.Images,
		},
	}

	return &firestoreMedia, nil
}

// getSpotifyAlbum calls Spotify API to get playlist metadata
func getSpotifyPlaylist(playlistID string, accessToken string) (*firebase.FirestoreMedia, error) {
	// Generate URI
	spotifyGetURI := spotifyGetPlaylistURI + "/" + playlistID

	// Generate request
	spotifyPlaylistReq, spotifyPlaylistReqErr := http.NewRequest("GET", spotifyGetURI, nil)
	if spotifyPlaylistReqErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [http.NewRequest]: %v", spotifyPlaylistReqErr)
	}

	// Add headers and call request
	spotifyPlaylistReq.Header.Add("Authorization", "Bearer "+accessToken)
	getPlaylistResp, spotifyPlaylistRespErr := httpClient.Do(spotifyPlaylistReq)
	if spotifyPlaylistRespErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [client.Do]: %v", spotifyPlaylistRespErr)
	}

	// Check to see if request was valid
	if getPlaylistResp.StatusCode != http.StatusOK {
		// Convert Spotify Error Object
		var spotifyErrorObj social.SpotifyRequestError

		err := json.NewDecoder(getPlaylistResp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetSpotifyMedia [Spotify Request Decoder]: %v", err)
		}
		return nil, fmt.Errorf("GetSpotifyMedia [Spotify Track Request]: %v", spotifyErrorObj.Error.Message)
	}

	var playlistData spotifyPlaylistResp

	// syszen - "wait that it? #easyGo"(02/27/20)
	// LilCazza - "Why the fuck doesn't this shit work" (02/27/20)
	respDecodeErr := json.NewDecoder(getPlaylistResp.Body).Decode(&playlistData)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [Spotify Response Decoder]: %v", respDecodeErr)
	}

	// Setup FirestoreMeida object
	firestoreMedia := firebase.FirestoreMedia{
		Name:    playlistData.Name,
		Album:   "Playlist",
		Type:    "playlist",
		Creator: playlistData.Owner.DisplayName,
		Spotify: firebase.FirestoreMediaPlatformData{
			ID:     playlistData.ID,
			URL:    playlistData.ExternalURLs.Spotify,
			Images: playlistData.Images,
		},
	}

	return &firestoreMedia, nil
}

// getSpotifyAlbum calls Spotify API to get album metadata
func getSpotifyAlbum(albumID string, accessToken string) (*firebase.FirestoreMedia, error) {
	// Generate URI
	spotifyGetURI := spotifyGetAlbumURI + "/" + albumID

	// Generate request
	spotifyAlbumReq, spotifyAlbumReqErr := http.NewRequest("GET", spotifyGetURI, nil)
	if spotifyAlbumReqErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [http.NewRequest]: %v", spotifyAlbumReqErr)
	}

	// Add headers and call request
	spotifyAlbumReq.Header.Add("Authorization", "Bearer "+accessToken)
	spotifyAlbumResp, spotifyAlbumRespErr := httpClient.Do(spotifyAlbumReq)
	if spotifyAlbumRespErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [client.Do]: %v", spotifyAlbumRespErr)
	}

	// Check to see if request was valid
	if spotifyAlbumResp.StatusCode != http.StatusOK {
		// Convert Spotify Error Object
		var spotifyErrorObj social.SpotifyRequestError

		err := json.NewDecoder(spotifyAlbumResp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetSpotifyMedia [Spotify Request Decoder]: %v", err)
		}
		return nil, fmt.Errorf("GetSpotifyMedia [Spotify Track Request]: %v", spotifyErrorObj.Error.Message)
	}

	var albumData spotifyAlbum

	// syszen - "wait that it? #easyGo"(02/27/20)
	// LilCazza - "Why the fuck doesn't this shit work" (02/27/20)
	respDecodeErr := json.NewDecoder(spotifyAlbumResp.Body).Decode(&albumData)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [Spotify Response Decoder]: %v", respDecodeErr)
	}

	// Setup FirestoreMeida object
	firestoreMedia := firebase.FirestoreMedia{
		Name:    albumData.Name,
		Album:   albumData.Name,
		Type:    albumData.Type,
		Creator: generateArtistsString(albumData.Artists),
		Spotify: firebase.FirestoreMediaPlatformData{
			ID:     albumData.ID,
			URL:    albumData.ExternalURLs.Spotify,
			Images: albumData.Images,
		},
	}

	return &firestoreMedia, nil
}

// generateArtistsString takes in a list of artists and returns a comma separated string
func generateArtistsString(artists []spotifyArtist) string {
	var creators = []string{}

	for _, artist := range artists {
		creators = append(creators, artist.Name)
	}

	return strings.Join(creators, ", ")
}
