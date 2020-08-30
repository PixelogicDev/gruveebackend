package getapplemusicmedia

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/mediahelpers"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

const catalogHostname = "https://api.music.apple.com/v1/catalog"

var (
	httpClient      *http.Client
	firestoreClient *firestore.Client
	logger          sawmill.Logger
)

func init() {
	log.Println("GetAppleMusicMedia Initialized")
}

// GetAppleMusicMedia will take in Apple media data and get the exact media from Apple Music API
func GetAppleMusicMedia(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnvErr", initWithEnvErr, nil)
		return
	}

	// Decode Request body to get media data
	var appleMusicMediaReq social.GetMediaReq
	appleMusicMediaReqErr := json.NewDecoder(request.Body).Decode(&appleMusicMediaReq)
	if appleMusicMediaReqErr != nil {
		http.Error(writer, appleMusicMediaReqErr.Error(), http.StatusInternalServerError)
		logger.LogErr("Request Decoder", appleMusicMediaReqErr, nil)
		return
	}

	logger.Log("GetAppleMusicMedia", "AppleMusicMediaReq decoded successfully.")

	// Check to see if media is already part of collection, if so, just return that
	mediaData, mediaDataErr := mediahelpers.GetMediaFromFirestore(*firestoreClient, appleMusicMediaReq.Provider, appleMusicMediaReq.MediaID)
	if mediaDataErr != nil {
		http.Error(writer, mediaDataErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GetMediaFromFirestore", mediaDataErr, nil)
		return
	}

	// MediaData exists, return it to the client
	if mediaData != nil {
		logger.Log("GetAppleMusicMedia", "Media already exists, returning")
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
		logger.LogErr("GetAppleDeveloperToken", appleDevTokeErr, nil)
		return
	}

	logger.Log("GetAppleMusicMedia", "Received Apple Developer Token.")

	// We only declare this here if we need to write new data
	var (
		firestoreMediaData    interface{}
		firestoreMediaDataErr error
	)

	// Time to make our request to Apple Music API
	switch appleMusicMediaReq.MediaType {
	case "track":
		logger.Log("GetAppleMusicMedia", "Making track request")
		firestoreMediaData, firestoreMediaDataErr = getAppleMusicTrack(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			logger.LogErr("GetAppleMusicTrack", firestoreMediaDataErr, nil)
			return
		}
	case "playlist":
		logger.Log("GetAppleMusicMedia", "Making playlist request")
		firestoreMediaData, firestoreMediaDataErr = getAppleMusicPlaylist(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			logger.LogErr("GetAppleMusicPlaylist", firestoreMediaDataErr, nil)
			return
		}
	case "album":
		logger.Log("GetAppleMusicMedia", "Making album request")
		firestoreMediaData, firestoreMediaDataErr = getAppleMusicAlbum(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			logger.LogErr("GetAppleMusicAlbum", firestoreMediaDataErr, nil)
			return
		}
	default:
		http.Error(writer, appleMusicMediaReq.MediaType+" media type does not exist", http.StatusInternalServerError)
		logger.LogErr("GetAppleMusicDefault", fmt.Errorf("%v media type does not exist", appleMusicMediaReq.MediaType), nil)
		return
	}

	logger.Log("GetAppleMusicMedia", "Successfully got Apple Music Media.")

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(firestoreMediaData)
}
