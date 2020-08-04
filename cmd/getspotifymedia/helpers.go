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

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
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
func getCreds() (*social.SpotifyClientCredsAuthResp, error) {
	// Check to see if we have env variables
	if os.Getenv("SPOTIFY_CLIENTID") == "" || os.Getenv("SPOTIFY_SECRET") == "" {
		log.Fatalln("GetSpotifyMedia [Check Env Props]: PROPS NOT HERE.")
		return nil, fmt.Errorf("getSpotifyMedia [Check Env Props]: PROPS NOT HERE")
	}

	// Generate authStr for requests
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

	// Decode the token to send back
	var spotifyClientCredsAuthResp social.SpotifyClientCredsAuthResp
	refreshTokenDecodeErr := json.NewDecoder(accessTokenResp.Body).Decode(&spotifyClientCredsAuthResp)
	if refreshTokenDecodeErr != nil {
		return nil, fmt.Errorf(refreshTokenDecodeErr.Error())
	}

	return &spotifyClientCredsAuthResp, nil

	// TODO: This block of code checked for refresh of token
	// At this point we are getting a new token every time
	/* spotifyCredsRef := firestoreClient.Doc("platform_credentials/spotify")
	if spotifyCredsRef == nil {
		return nil, fmt.Errorf("spotify credentials do not exist")
	}

	// Grab token see if it is expired
	spotifyCredSnap, spotifyCredSnapErr := spotifyCredsRef.Get(context.Background())
	if status.Code(spotifyCredSnapErr) == codes.NotFound {
		return nil, fmt.Errorf("Spotify cred was not found")
	}

	var spotifyCreds platformCredentials
	dataErr := spotifyCredSnap.DataTo(&spotifyCreds)
	if dataErr != nil {
		return nil, fmt.Errorf("doesUserExist: %v", dataErr)
	}

	log.Println(spotifyCreds)
	return &spotifyCreds, nil */
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

// GenerateArtistsString takes in a list of artists and returns a comma separated string
func generateArtistsString(artists []spotifyArtist) string {
	var creators = []string{}

	for _, artist := range artists {
		creators = append(creators, artist.Name)
	}

	return strings.Join(creators, ", ")
}

// refreshToken will check for an expired token and call Spotify refresh if needed
/* func refreshToken(creds platformCredentials) (*firebase.APIToken, error) {
	// Get current time
	var currentTime = time.Now()

	fmt.Printf("Expires In: %d seconds\n", creds.APIToken.ExpiresIn)
	fmt.Printf("Expired At: %s\n", creds.APIToken.ExpiredAt)
	fmt.Printf("Created At: %s\n", creds.APIToken.CreatedAt)

	expiredAtTime, expiredAtTimeErr := time.Parse(time.RFC3339, creds.APIToken.ExpiredAt)
	if expiredAtTimeErr != nil {
		return nil, fmt.Errorf(expiredAtTimeErr.Error())
	}

	if currentTime.After(expiredAtTime) {
		// Call API refresh
		fmt.Printf("Access token is expired. Calling Refresh...\n")
		refreshToken, tokenActionErr := refreshTokenAction(platform)
		if tokenActionErr != nil {
			return nil, fmt.Errorf(tokenActionErr.Error())
		}

		var expiredAtStr = time.Now().Add(time.Second * time.Duration(refreshToken.ExpiresIn))
		var refreshedAPIToken = firebase.APIToken{
			CreatedAt: time.Now().Format(time.RFC3339),
			ExpiredAt: expiredAtStr.Format(time.RFC3339),
			ExpiresIn: refreshToken.ExpiresIn,
			Token:     refreshToken.AccessToken,
		}

		// Set new token data in database
		writeTokenErr := writeToken(platform.ID, refreshedAPIToken)
		if writeTokenErr != nil {
			fmt.Errorf(writeTokenErr.Error())
		}

		return &refreshedAPIToken, nil
	}

	// Nothing is expired just return original token
	return nil, nil
} */
