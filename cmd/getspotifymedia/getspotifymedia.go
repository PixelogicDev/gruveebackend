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

// -- Types -- //

// getSpotifyMediaReq takes in the data needed to request the media data from Spotify
type getSpotifyMediaReq struct {
	Provider  string `json:"provider"`
	MediaID   string `json:"mediaId"`
	MediaType string `json:"mediaType"`
}

// spotifyTrackResp defines the data returned and needed from the Spotify Get Track API
type spotifyTrackResp struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Artists      []spotifyArtist `json:"artists"`
	Type         string          `json:"type"`
	Album        spotifyAlbum    `json:"album"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// spotifyPlaylistResp defines the data returned and needed from the Spotify Get Playlist API
type spotifyPlaylistResp struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Type         string                  `json:"type"`
	Images       []firebase.SpotifyImage `json:"images"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// spotifyAlbum defines the data returned and needed from the Spotify Get Track API
type spotifyAlbum struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Type         string                  `json:"type"`
	Artists      []spotifyArtist         `json:"artists"`
	Images       []firebase.SpotifyImage `json:"images"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// spotifyArtist defines the data returned and needed from the Spotify Get Track API
type spotifyArtist struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

var httpClient *http.Client
var firestoreClient *firestore.Client
var spotifyAccessTokenURI = "https://accounts.spotify.com/api/token"

// Get Media Endpoints
var spotifyGetTrackURI = "https://api.spotify.com/v1/tracks"
var spotifyGetPlaylistURI = "https://api.spotify.com/v1/playlists"
var spotifyGetAlbumURI = "https://api.spotify.com/v1/albums"

// Draco401 - "Draco401 was here." (04/17/20)
func init() {
	client, err := firestore.NewClient(context.Background(), "gruvee-3b7c4")
	if err != nil {
		log.Printf("GetSpotifyMedia [Init Firestore]: %v", err)
		return
	}
	firestoreClient = client

	httpClient = &http.Client{}

	log.Println("GetSpotifyMedia Initialized")
}

// GetSpotifyMedia will take in Spotify media data and get the exact media from Spotify API
func GetSpotifyMedia(writer http.ResponseWriter, request *http.Request) {
	// Get Spotify access token
	creds, credErr := getCreds()
	if credErr != nil {
		http.Error(writer, credErr.Error(), http.StatusInternalServerError)
		log.Printf("GetSpotifyMedia [getCreds]: %v", credErr)
		return
	}

	// Decode Request body to get track data
	var spotifyMediaReq getSpotifyMediaReq
	spotifyReqDecodeErr := json.NewDecoder(request.Body).Decode(&spotifyMediaReq)
	if spotifyReqDecodeErr != nil {
		http.Error(writer, spotifyReqDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("GetSpotifyMedia [Request Decoder]: %v", spotifyReqDecodeErr)
		return
	}

	log.Println(spotifyMediaReq)

	// Setup and call Spotify search
	switch spotifyMediaReq.MediaType {
	case "track":
		// Call track API
		firestoreMediaData, firestoreMediaDataErr := getSpotifyTrack(spotifyMediaReq.MediaID, creds.AccessToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetSpotifyMedia [GetSpotifyTrack Switch]: %v", firestoreMediaDataErr)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(firestoreMediaData)
		return
	case "playlist":
		// Call playlist API
		spotifyPlaylistData, spotifyPlaylistDataErr := getSpotifyPlaylist(spotifyMediaReq.MediaID, creds.AccessToken)
		if spotifyPlaylistDataErr != nil {
			http.Error(writer, spotifyPlaylistDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetSpotifyMedia [GetSpotifyPlaylist Switch]: %v", spotifyPlaylistDataErr)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(spotifyPlaylistData)
		return
	case "album":
		// Call album API
		spotifyAlbumData, spotifyAlbumDataErr := getSpotifyAlbum(spotifyMediaReq.MediaID, creds.AccessToken)
		if spotifyAlbumDataErr != nil {
			http.Error(writer, spotifyAlbumDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetSpotifyMedia [GetSpotifyPlaylist Switch]: %v", spotifyAlbumDataErr)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(spotifyAlbumData)
		return
	default:
		http.Error(writer, spotifyMediaReq.MediaType+" media type does not exist", http.StatusInternalServerError)
		log.Printf("GetSpotifyMedia [MediaTypeSwitch]: %v media type does not exist", spotifyMediaReq.MediaType)
		return
	}
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

	// For now, let's just get the token directly from spotify
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

	// If multiple artists append to a string
	var creators string
	for index, artist := range spotifyTrackData.Artists {
		if index == 0 {
			creators = artist.Name
		} else {
			creators = creators + ", " + artist.Name
		}
	}

	// Setup FirestoreMeida object
	firestoreMedia := firebase.FirestoreMedia{
		ID:           spotifyTrackData.ID,
		Name:         spotifyTrackData.Name,
		Album:        spotifyTrackData.Album.Name,
		Type:         spotifyTrackData.Type,
		Creator:      creators,
		Images:       spotifyTrackData.Album.Images,
		ExternalURLs: map[string]string{"spotify": spotifyTrackData.ExternalURLs.Spotify},
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

	return nil, nil
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

	return nil, nil
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
