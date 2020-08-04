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

// getAppleMusicPlaylist will call Apple Music Playlist API to get the metadata for a playlist
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

	return &firestoreMedia, nil
}

// getAppleMusicAlbum will call Apple Music Album API to get the metadata for an album
func getAppleMusicAlbum(albumID string, storefront string, appleDevToken firebase.FirestoreAppleDevJWT) (*firebase.FirestoreMedia, error) {
	// Generate request
	appleMusicGetPlaylistReq, appleMusicGetTrackReqErr := generateAppleMusicReq(catalogHostname+"/"+storefront+"/albums/"+albumID, "GET", appleDevToken.Token)
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

	var appleMusicAlbumResp appleMusicAlbumResp
	respDecodeErr := json.NewDecoder(getPlaylistResp.Body).Decode(&appleMusicAlbumResp)
	if respDecodeErr != nil {
		return nil, fmt.Errorf("GetAppleMusicMedia [Apple Music Response Decoder]: %v", respDecodeErr)
	}

	log.Println(appleMusicAlbumResp)

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

	return &firestoreMedia, nil
}
