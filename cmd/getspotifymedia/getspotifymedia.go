package getspotifymedia

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/mediahelpers"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

// -- Types -- //

// getSpotifyMediaReq takes in the data needed to request the media data from Spotify
type getSpotifyMediaReq struct {
	Provider  string `json:"provider"`
	MediaID   string `json:"mediaId"`
	MediaType string `json:"mediaType"`
}

var (
	httpClient            *http.Client
	firestoreClient       *firestore.Client
	logger                sawmill.Logger
	spotifyAccessTokenURI = "https://accounts.spotify.com/api/token"
	spotifyGetTrackURI    = "https://api.spotify.com/v1/tracks"
	spotifyGetPlaylistURI = "https://api.spotify.com/v1/playlists"
	spotifyGetAlbumURI    = "https://api.spotify.com/v1/albums"
)

// Draco401 - "Draco401 was here." (04/17/20)
func init() {
	log.Println("GetSpotifyMedia Initialized")
}

// GetSpotifyMedia will take in Spotify media data and get the exact media from Spotify API
func GetSpotifyMedia(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initWithEnvErr, nil)
		return
	}

	// Decode Request body to get track data
	var spotifyMediaReq social.GetMediaReq
	spotifyReqDecodeErr := json.NewDecoder(request.Body).Decode(&spotifyMediaReq)
	if spotifyReqDecodeErr != nil {
		http.Error(writer, spotifyReqDecodeErr.Error(), http.StatusInternalServerError)
		logger.LogErr("Request Decoder", spotifyReqDecodeErr, request)
		return
	}

	// Check to see if media is already part of collection, if so, just return that
	mediaData, mediaDataErr := mediahelpers.GetMediaFromFirestore(*firestoreClient, spotifyMediaReq.Provider, spotifyMediaReq.MediaID)
	if mediaDataErr != nil {
		http.Error(writer, mediaDataErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GetMediaFromFirestore", mediaDataErr, request)
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

	// Get Spotify access token (currently getting access token of user)
	creds, credErr := getCreds()
	if credErr != nil {
		http.Error(writer, credErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GetCreds", credErr, nil)
		return
	}

	var (
		firestoreMediaData    interface{}
		firestoreMediaDataErr error
	)

	// Setup and call Spotify search
	switch spotifyMediaReq.MediaType {
	case "track":
		firestoreMediaData, firestoreMediaDataErr = getSpotifyTrack(spotifyMediaReq.MediaID, creds.Token)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			logger.LogErr("GetSpotifyAlbum Switch", firestoreMediaDataErr, request)
			return
		}
	case "playlist":
		firestoreMediaData, firestoreMediaDataErr = getSpotifyPlaylist(spotifyMediaReq.MediaID, creds.Token)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			logger.LogErr("GetSpotifyAlbum Switch", firestoreMediaDataErr, request)
			return
		}
	case "album":
		firestoreMediaData, firestoreMediaDataErr = getSpotifyAlbum(spotifyMediaReq.MediaID, creds.Token)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			logger.LogErr("GetSpotifyAlbum Switch", firestoreMediaDataErr, request)
			return
		}
	default:
		http.Error(writer, spotifyMediaReq.MediaType+" media type does not exist", http.StatusInternalServerError)
		log.Printf("GetSpotifyMedia [MediaTypeSwitch]: %v media type does not exist", spotifyMediaReq.MediaType)
		logger.LogErr("MediaTypeSwitch", fmt.Errorf("%v media type does not exist", spotifyMediaReq.MediaType), request)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(firestoreMediaData)
}
