package getspotifymedia

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	logger.Log("GetCreds", "Starting...")

	// Check to see if we have env variables
	if os.Getenv("SPOTIFY_CLIENTID") == "" || os.Getenv("SPOTIFY_SECRET") == "" {
		return nil, fmt.Errorf("getSpotifyMedia [Check Env Props]: PROPS NOT HERE")
	}

	logger.Log("GetCreds", "Found Spotify ClientID in env variables")

	// Check for Auth token in DB
	authToken, authTokenErr := fetchToken()
	if authTokenErr != nil {
		return nil, authTokenErr
	}

	logger.Log("GetCreds", "Received token.")

	// Check to see if auth token exists && needs to be refreshed
	if authToken != nil {
		logger.Log("GetCreds", "Checking to see if token needs to be refreshed")

		latestCreds, refreshAuthTokenErr := refreshAuthToken(*authToken)
		if refreshAuthTokenErr != nil {
			return nil, refreshAuthTokenErr
		}

		logger.Log("GetCreds", "Auth token refresh complete.")

		return latestCreds, nil
	}

	logger.Log("GetCreds", "No auth token found.")

	// If we are here, no auth token was found
	newAuthToken, newAuthTokenErr := generateAuthToken()
	if newAuthTokenErr != nil {
		return nil, newAuthTokenErr
	}

	logger.Log("GetCreds", "Successfully generated new auth token.")

	// Store new token in DB
	writeSpotifyAuthErr := writeSpotifyAuthtoken(*newAuthToken)
	if writeSpotifyAuthErr != nil {
		return nil, writeSpotifyAuthErr
	}

	logger.Log("GetCreds", "Successfully stored new auth token in DB.")

	// TheYagich01: "Gejnerated" (08/11/20)
	return newAuthToken, nil
}

// fetchToken will grab the Apple Developer Token from DB
func fetchToken() (*firebase.FirestoreSpotifyAuthToken, error) {
	logger.Log("FetchToken", "Starting...")

	// Go to Firebase and see if spotifyAuthToken exists
	snapshot, snapshotErr := firestoreClient.Collection("internal_tokens").Doc("spotifyAuthToken").Get(context.Background())
	if status.Code(snapshotErr) == codes.NotFound {
		logger.Log("FetchToken", "SpotifyAuthToken not found in DB. Need to create.")
		return nil, nil
	}

	if snapshotErr != nil {
		return nil, fmt.Errorf(snapshotErr.Error())
	}

	logger.Log("FetchToken", "Successfully got snapshot.")

	var spotifyAuthToken firebase.FirestoreSpotifyAuthToken
	dataToErr := snapshot.DataTo(&spotifyAuthToken)
	if dataToErr != nil {
		return nil, fmt.Errorf(dataToErr.Error())
	}

	logger.Log("FetchToken", "Successfully decoded auth token.")

	return &spotifyAuthToken, nil
}

// generateAuthToken will call Spotify creds service to get authToken for making requests
func generateAuthToken() (*firebase.FirestoreSpotifyAuthToken, error) {
	logger.Log("GenerateAuthToken", "Starting...")

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

	logger.Log("GenerateAuthToken", "Request generated.")

	accessTokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	accessTokenReq.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(authStr)))
	accessTokenResp, accessTokenRespErr := httpClient.Do(accessTokenReq)
	if accessTokenRespErr != nil {
		return nil, fmt.Errorf(accessTokenRespErr.Error())
	}

	logger.Log("GenerateAuthToken", "Response received.")

	// Decode the token
	var spotifyClientCredsAuthResp social.SpotifyClientCredsAuthResp
	refreshTokenDecodeErr := json.NewDecoder(accessTokenResp.Body).Decode(&spotifyClientCredsAuthResp)
	if refreshTokenDecodeErr != nil {
		return nil, fmt.Errorf(refreshTokenDecodeErr.Error())
	}

	logger.Log("GenerateAuthToken", "Successfully decoded token.")

	// Generate authToken for DB
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(time.Second * time.Duration(spotifyClientCredsAuthResp.ExpiresIn))
	authToken := firebase.FirestoreSpotifyAuthToken{
		IssuedAt:  issuedAt.Format(time.RFC3339),
		ExpiresIn: spotifyClientCredsAuthResp.ExpiresIn,
		ExpiredAt: expiresAt.Format(time.RFC3339),
		Token:     spotifyClientCredsAuthResp.AccessToken,
	}

	logger.Log("GenerateAuthToken", "Successfully generated token to store in DB.")

	return &authToken, nil
}

// refreshAuthToken will check to see if Spotify AuthToken needs to be refreshed
func refreshAuthToken(authToken firebase.FirestoreSpotifyAuthToken) (*firebase.FirestoreSpotifyAuthToken, error) {
	logger.Log("RefreshAuthToken", "Starting...")

	// Get current time
	var currentTime = time.Now()

	logger.Log("RefreshAuthToken", fmt.Sprintf("Expires In: %d seconds\n", authToken.ExpiresIn))
	logger.Log("RefreshAuthToken", fmt.Sprintf("Expires At: %s seconds\n", authToken.ExpiredAt))

	expiredAtTime, expiredAtTimeErr := time.Parse(time.RFC3339, authToken.ExpiredAt)
	if expiredAtTimeErr != nil {
		return nil, fmt.Errorf(expiredAtTimeErr.Error())
	}

	logger.Log("RefreshAuthToken", "Successfully parsed ExpiredAtTime.")

	if currentTime.After(expiredAtTime) {
		logger.Log("RefreshAuthToken", "SpotifyAuthToken is expired. Refreshing...")

		refreshToken, tokenActionErr := generateAuthToken()
		if tokenActionErr != nil {
			return nil, fmt.Errorf(tokenActionErr.Error())
		}

		logger.Log("RefreshAuthToken", "Successfully generated auth token.")

		// Set new token data in database
		writeTokenErr := writeSpotifyAuthtoken(*refreshToken)
		if writeTokenErr != nil {
			return nil, fmt.Errorf(writeTokenErr.Error())
		}

		logger.Log("RefreshAuthToken", "Successfully wrote token to Firestore")

		return refreshToken, nil
	}

	// Nothing is expired just return original token
	return &authToken, nil
}

// writeSpotifyAuthtoken will take spotifyAuthToken and write it to the internal_tokens collection
func writeSpotifyAuthtoken(authToken firebase.FirestoreSpotifyAuthToken) error {
	logger.Log("WriteSpotifyAuthtoken", "Starting...")

	spotifyAuthTokenDoc := firestoreClient.Collection("internal_tokens").Doc("spotifyAuthToken")
	if spotifyAuthTokenDoc == nil {
		return fmt.Errorf("spotifyAuthTokenDoc could not be found")
	}

	logger.Log("WriteSpotifyAuthtoken", "Successfully received SpotifyAuthTokenDoc.")

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

	logger.Log("WriteSpotifyAuthtoken", "Successfully set token in Firestore.")

	return nil
}

// getSpotifyTrack calls Spotify GET track API and converts to Golang Type
func getSpotifyTrack(trackID string, accessToken string) (*firebase.FirestoreMedia, error) {
	logger.Log("GetSpotifyTrack", "Starting...")

	// Generate URI
	spotifyGetURI := spotifyGetTrackURI + "/" + trackID

	// Generate request
	spotifyTrackReq, spotifyTrackReqErr := http.NewRequest("GET", spotifyGetURI, nil)
	if spotifyTrackReqErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [http.NewRequest]: %v", spotifyTrackReqErr)
	}

	logger.Log("GetSpotifyTrack", "Generated track request.")

	// Add headers and call request
	spotifyTrackReq.Header.Add("Authorization", "Bearer "+accessToken)
	getTrackResp, spotifyTrackRespErr := httpClient.Do(spotifyTrackReq)
	if spotifyTrackRespErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [client.Do]: %v", spotifyTrackRespErr)
	}

	logger.Log("GetSpotifyTrack", "Received track response.")

	// Check to see if request was valid
	if getTrackResp.StatusCode != http.StatusOK {
		logger.Log("GetSpotifyTrack", "Received error. Decoding...")

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

	logger.Log("GetSpotifyTrack", "Successfully decoded response data.")

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

	logger.Log("GetSpotifyTrack", "Successfully created firestore media object.")

	return &firestoreMedia, nil
}

// getSpotifyAlbum calls Spotify API to get playlist metadata
func getSpotifyPlaylist(playlistID string, accessToken string) (*firebase.FirestoreMedia, error) {
	logger.Log("GetSpotifyPlaylist", "Starting...")

	// Generate URI
	spotifyGetURI := spotifyGetPlaylistURI + "/" + playlistID

	// Generate request
	spotifyPlaylistReq, spotifyPlaylistReqErr := http.NewRequest("GET", spotifyGetURI, nil)
	if spotifyPlaylistReqErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [http.NewRequest]: %v", spotifyPlaylistReqErr)
	}

	logger.Log("GetSpotifyPlaylist", "Generated playlist request.")

	// Add headers and call request
	spotifyPlaylistReq.Header.Add("Authorization", "Bearer "+accessToken)
	getPlaylistResp, spotifyPlaylistRespErr := httpClient.Do(spotifyPlaylistReq)
	if spotifyPlaylistRespErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [client.Do]: %v", spotifyPlaylistRespErr)
	}

	logger.Log("GetSpotifyPlaylist", "Received playlist response.")

	// Check to see if request was valid
	if getPlaylistResp.StatusCode != http.StatusOK {
		logger.Log("GetSpotifyPlaylist", "Received error. Decoding...")

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

	logger.Log("GetSpotifyPlaylist", "Successfully decoded response.")

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

	logger.Log("GetSpotifyPlaylist", "Successfully created firestore media object.")

	return &firestoreMedia, nil
}

// getSpotifyAlbum calls Spotify API to get album metadata
func getSpotifyAlbum(albumID string, accessToken string) (*firebase.FirestoreMedia, error) {
	logger.Log("GetSpotifyAlbum", "Starting...")

	// Generate URI
	spotifyGetURI := spotifyGetAlbumURI + "/" + albumID

	// Generate request
	spotifyAlbumReq, spotifyAlbumReqErr := http.NewRequest("GET", spotifyGetURI, nil)
	if spotifyAlbumReqErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [http.NewRequest]: %v", spotifyAlbumReqErr)
	}

	logger.Log("GetSpotifyAlbum", "Generated album request.")

	// Add headers and call request
	spotifyAlbumReq.Header.Add("Authorization", "Bearer "+accessToken)
	spotifyAlbumResp, spotifyAlbumRespErr := httpClient.Do(spotifyAlbumReq)
	if spotifyAlbumRespErr != nil {
		return nil, fmt.Errorf("GetSpotifyMedia [client.Do]: %v", spotifyAlbumRespErr)
	}

	logger.Log("GetSpotifyAlbum", "Received album response.")

	// Check to see if request was valid
	if spotifyAlbumResp.StatusCode != http.StatusOK {
		logger.Log("GetSpotifyAlbum", "Received error. Decoding...")

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

	logger.Log("GetSpotifyAlbum", "Successfully decoded album data.")

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

	logger.Log("GetSpotifyAlbum", "Successfully created firestore media object.")

	return &firestoreMedia, nil
}

// generateArtistsString takes in a list of artists and returns a comma separated string
func generateArtistsString(artists []spotifyArtist) string {
	logger.Log("GenerateArtistsString", "Starting...")

	var creators = []string{}

	for _, artist := range artists {
		creators = append(creators, artist.Name)
	}

	logger.Log("GenerateArtistsString", "Successfully appended artists.")

	return strings.Join(creators, ", ")
}
