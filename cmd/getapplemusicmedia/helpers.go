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

	// Initialize HttpClient
	httpClient = &http.Client{}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("CreateUser [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "FetchAllMedia")
	if err != nil {
		log.Printf("FetchAllMedia [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}

// generateAppleMusicReq created the request object to call Apple Music API
func generateAppleMusicReq(uri string, method string, devToken string) (*http.Request, error) {
	logger.Log("GenerateAppleMusicReq", "Starting...")

	// Generate request
	appleMusicReq, appleMusicReqErr := http.NewRequest(method, uri, nil)
	if appleMusicReqErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMeda [http.NewRequest]: %v", appleMusicReqErr)
	}

	logger.Log("GenerateAppleMusicReq", "Generarted AppleMusicReq")

	// Add headers
	appleMusicReq.Header.Add("Authorization", "Bearer "+devToken)

	logger.Log("GenerateAppleMusicReq", "Created request successfully.")

	return appleMusicReq, nil
}

// getAppleMusicTrack will call Apple Music Song API to get the metadata for a song
func getAppleMusicTrack(trackID string, storefront string, appleDevToken firebase.FirestoreAppleDevJWT) (*firebase.FirestoreMedia, error) {
	logger.Log("GetAppleMusicTrack", "Starting...")

	// Generate request
	appleMusicGetTrackReq, appleMusicGetTrackReqErr := generateAppleMusicReq(catalogHostname+"/"+storefront+"/songs/"+trackID, "GET", appleDevToken.Token)
	if appleMusicGetTrackReqErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMeda [http.NewRequest]: %v", appleMusicGetTrackReqErr)
	}

	logger.Log("GetAppleMusicTrack", "Generated Apple Music Request.")

	// Make request
	getTrackResp, getTrackRespErr := httpClient.Do(appleMusicGetTrackReq)
	if getTrackRespErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [client.Do]: %v", getTrackRespErr)
	}

	logger.Log("GetAppleMusicTrack", "Received Track Response.")

	// Check to see if request was valid
	if getTrackResp.StatusCode != http.StatusOK {
		logger.Log("GetAppleMusicTrack", "Error received. Decoding body...")

		// Convert Apple Music Error Object
		var appleMusicErrorObj social.AppleMusicRequestError

		err := json.NewDecoder(getTrackResp.Body).Decode(&appleMusicErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Request Decoder]: %v", err)
		}

		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Track Request]: %v", appleMusicErrorObj)
	}

	logger.Log("GetAppleMusicTrack", "Successful response.")

	var appleMusicTrackData appleMusicTrackResp
	respDecodeErr := json.NewDecoder(getTrackResp.Body).Decode(&appleMusicTrackData)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Response Decoder]: %v", respDecodeErr)
	}

	logger.Log("GetAppleMusicTrack", "Decoded Apple Musc Track Data.")

	// Check for length to make sure we found a match
	if len(appleMusicTrackData.Data) == 0 {
		logger.Log("GetAppleMusicTrack", "No matches found.")
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

	logger.Log("GetAppleMusicTrack", "Successfully returning track.")

	return &firestoreMedia, nil
}

// getAppleMusicPlaylist will call Apple Music Playlist API to get the metadata for a playlist
func getAppleMusicPlaylist(playlistID string, storefront string, appleDevToken firebase.FirestoreAppleDevJWT) (*firebase.FirestoreMedia, error) {
	logger.Log("GetAppleMusicPlaylist", "Starting...")

	// Generate request
	appleMusicGetPlaylistReq, appleMusicGetTrackReqErr := generateAppleMusicReq(catalogHostname+"/"+storefront+"/playlists/"+playlistID, "GET", appleDevToken.Token)
	if appleMusicGetTrackReqErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMeda [http.NewRequest]: %v", appleMusicGetTrackReqErr)
	}

	logger.Log("GetAppleMusicPlaylist", "Generated request.")

	getPlaylistResp, getPlaylistRespErr := httpClient.Do(appleMusicGetPlaylistReq)
	if getPlaylistRespErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [client.Do]: %v", getPlaylistRespErr)
	}

	logger.Log("GetAppleMusicPlaylist", "Received playlist response")

	// Check to see if request was valid
	if getPlaylistResp.StatusCode != http.StatusOK {
		logger.Log("GetAppleMusicPlaylist", "Error received. Decoding body...")

		// Convert Apple Music Error Object
		var appleMusicErrorObj social.AppleMusicRequestError

		err := json.NewDecoder(getPlaylistResp.Body).Decode(&appleMusicErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Request Decoder]: %v", err)
		}

		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Track Request]: %v", appleMusicErrorObj)
	}

	logger.Log("GetAppleMusicPlaylist", "Successful response.")

	var appleMusicPlaylistResp appleMusicPlaylistResp
	respDecodeErr := json.NewDecoder(getPlaylistResp.Body).Decode(&appleMusicPlaylistResp)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Response Decoder]: %v", respDecodeErr)
	}

	logger.Log("GetAppleMusicPlaylist", "Successfully decoded response")

	// Check for length to make sure we found a match
	if len(appleMusicPlaylistResp.Data) == 0 {
		return nil, fmt.Errorf("GetAppleMusicMedia [Length Check]: No results in data for id %s", playlistID)
	}

	// Setup FirestoreMedia object
	playlist := appleMusicPlaylistResp.Data[0]
	firestoreMedia := firebase.FirestoreMedia{
		Name:    playlist.Attributes.Name,
		Album:   "Playlist",
		Type:    "playlist",
		Creator: playlist.Attributes.CuratorName,
		Apple: firebase.FirestoreMediaPlatformData{
			ID:     playlistID,
			URL:    playlist.Attributes.URL,
			Images: playlist.Attributes.Artwork,
		},
	}

	logger.Log("GetAppleMusicPlaylist", "Successfully created playlist data.")

	return &firestoreMedia, nil
}

// getAppleMusicAlbum will call Apple Music Album API to get the metadata for an album
func getAppleMusicAlbum(albumID string, storefront string, appleDevToken firebase.FirestoreAppleDevJWT) (*firebase.FirestoreMedia, error) {
	logger.Log("GetAppleMusicAlbum", "Starting...")

	// Generate request
	appleMusicGetAlbumReq, appleMusicGetTrackReqErr := generateAppleMusicReq(catalogHostname+"/"+storefront+"/albums/"+albumID, "GET", appleDevToken.Token)
	if appleMusicGetTrackReqErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMeda [http.NewRequest]: %v", appleMusicGetTrackReqErr)
	}

	logger.Log("GetAppleMusicAlbum", "Generated album request.")

	getAlbumResp, getPlaylistRespErr := httpClient.Do(appleMusicGetAlbumReq)
	if getPlaylistRespErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [client.Do]: %v", getPlaylistRespErr)
	}

	logger.Log("GetAppleMusicAlbum", "Received album")

	// Check to see if request was valid
	if getAlbumResp.StatusCode != http.StatusOK {
		logger.Log("GetAppleMusicAlbum", "Error received. Decoding body...")

		// Convert Apple Music Error Object
		var appleMusicErrorObj social.AppleMusicRequestError

		err := json.NewDecoder(getAlbumResp.Body).Decode(&appleMusicErrorObj)
		if err != nil {
			return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Request Decoder]: %v", err)
		}

		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Track Request]: %v", appleMusicErrorObj)
	}

	var appleMusicAlbumResp appleMusicAlbumResp
	respDecodeErr := json.NewDecoder(getAlbumResp.Body).Decode(&appleMusicAlbumResp)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Response Decoder]: %v", respDecodeErr)
	}

	logger.Log("GetAppleMusicAlbum", "Decoded response")

	// Check for length to make sure we found a match
	if len(appleMusicAlbumResp.Data) == 0 {
		return nil, fmt.Errorf("GetAppleMusicMedia [Length Check]: No results in data for id %s", albumID)
	}

	// Setup FirestoreMedia object
	album := appleMusicAlbumResp.Data[0]
	firestoreMedia := firebase.FirestoreMedia{
		Name:    album.Attributes.Name,
		Album:   "Album",
		Type:    "album",
		Creator: album.Attributes.ArtistName,
		Apple: firebase.FirestoreMediaPlatformData{
			ID:     albumID,
			URL:    album.Attributes.URL,
			Images: album.Attributes.Artwork,
		},
	}

	logger.Log("GetAppleMusicAlbum", "Successfully created Album response.")

	return &firestoreMedia, nil
}
