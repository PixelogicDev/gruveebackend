package getapplemusicmedia

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/mediahelpers"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

var httpClient *http.Client
var firestoreClient *firestore.Client

// -- Apple Music Endpoints --/

var catalogHostname = "https://api.music.apple.com/v1/catalog"

func init() {
	log.Println("GetAppleMusicMedia Initialized")
}

// GetAppleMusicMedia will take in Apple media data and get the exact media from Apple Music API
func GetAppleMusicMedia(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		log.Printf("GetAppleMusicMedia [initWithEnv]: %v", initWithEnvErr)
		return
	}

	// Decode Request body to get media data
	var appleMusicMediaReq social.GetMediaReq
	appleMusicMediaReqErr := json.NewDecoder(request.Body).Decode(&appleMusicMediaReq)
	if appleMusicMediaReqErr != nil {
		http.Error(writer, appleMusicMediaReqErr.Error(), http.StatusInternalServerError)
		log.Printf("GetAppleMusicMedia [Request Decoder]: %v", appleMusicMediaReqErr)
		return
	}

	// Check to see if media is already part of collection, if so, just return that
	mediaData, mediaDataErr := mediahelpers.GetMediaFromFirestore(*firestoreClient, appleMusicMediaReq.Provider, appleMusicMediaReq.MediaID)
	if mediaDataErr != nil {
		http.Error(writer, mediaDataErr.Error(), http.StatusInternalServerError)
		log.Printf("[GetAppleMusicMedia]: %v", mediaDataErr)
		return
	}

	// MediaData exists, return it to the client
	if mediaData != nil {
		log.Printf("Media already exists, returning")
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(mediaData)
		return
	}

	// MediaData does not exist, call Apple Music Endpoint
	// We need to get the developer token from firebase
	appleDevToken, appleDevTokeErr := firebase.GetAppleDeveloperToken()
	if appleDevTokeErr != nil {
		http.Error(writer, appleDevTokeErr.Error(), http.StatusInternalServerError)
		log.Printf("[GetAppleMusicMedia]: %v", appleDevTokeErr)
		return
	}

	// Time to make our request to Apple Music API
	switch appleMusicMediaReq.MediaType {
	case "track":
		// Call track API
		firestoreMediaTrackData, firestoreMediaDataErr := getAppleMusicTrack(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetSpotifyMedia [GetSpotifyTrack Switch]: %v", firestoreMediaDataErr)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(firestoreMediaTrackData)
		return
	case "playlist":
		// Call playlist API
		firestoreMediaPlaylistData, firestoreMediaDataErr := getAppleMusicPlaylist(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetSpotifyMedia [GetSpotifyPlaylist Switch]: %v", firestoreMediaDataErr)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(firestoreMediaPlaylistData)
		return
		// case "album":
		// 	// Call album API
		// 	firestoreMediaAlbumData, firestoreMediaDataErr := getSpotifyAlbum(spotifyMediaReq.MediaID, creds.AccessToken)
		// 	if firestoreMediaDataErr != nil {
		// 		http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
		// 		log.Printf("GetSpotifyMedia [GetSpotifyPlaylist Switch]: %v", firestoreMediaDataErr)
		// 		return
		// 	}

		// 	writer.WriteHeader(http.StatusOK)
		// 	writer.Header().Set("Content-Type", "application/json")
		// 	json.NewEncoder(writer).Encode(firestoreMediaAlbumData)
		// 	return
		// default:
		// 	http.Error(writer, spotifyMediaReq.MediaType+" media type does not exist", http.StatusInternalServerError)
		// 	log.Printf("GetSpotifyMedia [MediaTypeSwitch]: %v media type does not exist", spotifyMediaReq.MediaType)
		// 	return
	}

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

// generateAppleMusicReq created the request object to call Apple Music API
func generateAppleMusicReq(uri string, method string, devToken string) (*http.Request, error) {
	// Generate request
	appleMusicReq, appleMusicReqErr := http.NewRequest(method, uri, nil)
	if appleMusicReqErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMeda [http.NewRequest]: %v", appleMusicReqErr)
	}

	// Add headers
	appleMusicReq.Header.Add("Authorization", "Bearer "+devToken)

	return appleMusicReq, nil
}

// getAppleMusicTrack will call Apple Music Song API to get the metadata for a song
func getAppleMusicTrack(trackID string, storefront string, appleDevToken firebase.FirestoreAppleDevJWT) (*firebase.FirestoreMedia, error) {
	// Generate request
	appleMusicGetTrackReq, appleMusicGetTrackReqErr := generateAppleMusicReq(catalogHostname+"/"+storefront+"/songs/"+trackID, "GET", appleDevToken.Token)
	if appleMusicGetTrackReqErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMeda [http.NewRequest]: %v", appleMusicGetTrackReqErr)
	}

	// Make request
	getTrackResp, getTrackRespErr := httpClient.Do(appleMusicGetTrackReq)
	if getTrackRespErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [client.Do]: %v", getTrackRespErr)
	}

	// Check to see if request was valid
	if getTrackResp.StatusCode != http.StatusOK {
		// Convert Apple Music Error Object
		var appleMusicErrorObj social.AppleMusicRequestError

		err := json.NewDecoder(getTrackResp.Body).Decode(&appleMusicErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Request Decoder]: %v", err)
		}

		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Track Request]: %v", appleMusicErrorObj)
	}

	var appleMusicTrackData appleMusicTrackResp
	respDecodeErr := json.NewDecoder(getTrackResp.Body).Decode(&appleMusicTrackData)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Response Decoder]: %v", respDecodeErr)
	}

	log.Println(appleMusicTrackData)

	// Check for length to make sure we found a match
	if len(appleMusicTrackData.Data) == 0 {
		return nil, fmt.Errorf("GetAppleMusicMedia [Length Check]: No results in data for id %s", trackID)
	}

	// Setup FirestoreMeida object
	track := appleMusicTrackData.Data[0]
	firestoreMedia := firebase.FirestoreMedia{
		Name:    track.Attributes.TrackName,
		Album:   track.Attributes.AlbumName,
		Type:    "track",
		Creator: track.Attributes.ArtistName,
		Apple: firebase.FirestoreMediaPlatformData{
			ID:     trackID,
			URL:    track.Attributes.ExternalURL,
			Images: track.Attributes.Artwork,
		},
	}

	return &firestoreMedia, nil
}

// getSpotifyPlaylist will call Apple Music Playlist API to get the metadata for a playlist
func getAppleMusicPlaylist(playlistID string, storefront string, appleDevToken firebase.FirestoreAppleDevJWT) (*firebase.FirestoreMedia, error) {
	// Generate request
	appleMusicGetPlaylistReq, appleMusicGetTrackReqErr := generateAppleMusicReq(catalogHostname+"/"+storefront+"/playlists/"+playlistID, "GET", appleDevToken.Token)
	if appleMusicGetTrackReqErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMeda [http.NewRequest]: %v", appleMusicGetTrackReqErr)
	}

	getPlaylistResp, getPlaylistRespErr := httpClient.Do(appleMusicGetPlaylistReq)
	if getPlaylistRespErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [client.Do]: %v", getPlaylistRespErr)
	}

	// Check to see if request was valid
	if getPlaylistResp.StatusCode != http.StatusOK {
		// Convert Apple Music Error Object
		var appleMusicErrorObj social.AppleMusicRequestError

		err := json.NewDecoder(getPlaylistResp.Body).Decode(&appleMusicErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Request Decoder]: %v", err)
		}

		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Track Request]: %v", appleMusicErrorObj)
	}

	var appleMusicPlaylistResp appleMusicPlaylistResp
	respDecodeErr := json.NewDecoder(getPlaylistResp.Body).Decode(&appleMusicPlaylistResp)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Response Decoder]: %v", respDecodeErr)
	}

	log.Println(appleMusicPlaylistResp)

	// Check for length to make sure we found a match
	if len(appleMusicPlaylistResp.Data) == 0 {
		return nil, fmt.Errorf("GetAppleMusicMedia [Length Check]: No results in data for id %s", playlistID)
	}

	// Setup FirestoreMedia object
	playlist := appleMusicPlaylistResp.Data[0]
	firestoreMedia := firebase.FirestoreMedia{
		Name:    playlist.Attrbutes.Name,
		Album:   "Playlist",
		Type:    "playlist",
		Creator: playlist.Attrbutes.CuratorName,
		Apple: firebase.FirestoreMediaPlatformData{
			ID:     playlistID,
			URL:    playlist.Attrbutes.URL,
			Images: playlist.Attrbutes.Artwork,
		},
	}

	return &firestoreMedia, nil
}
