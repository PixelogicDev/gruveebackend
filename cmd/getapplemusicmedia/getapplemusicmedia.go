package getapplemusicmedia

import (
	"encoding/json"
	"log"
	"net/http"

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

	var (
		firestoreMediaData    interface{}
		firestoreMediaDataErr error
	)

	// Time to make our request to Apple Music API
	switch appleMusicMediaReq.MediaType {
	case "track":
		firestoreMediaData, firestoreMediaDataErr = getAppleMusicTrack(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetAppleMusicMedia [GetAppleMusicTrack Switch]: %v", firestoreMediaDataErr)
			return
		}
	case "playlist":
		firestoreMediaData, firestoreMediaDataErr = getAppleMusicPlaylist(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetAppleMusicMedia [GetAppleMusicPlaylist Switch]: %v", firestoreMediaDataErr)
			return
		}
	case "album":
		firestoreMediaData, firestoreMediaDataErr = getAppleMusicAlbum(appleMusicMediaReq.MediaID, appleMusicMediaReq.Storefront, *appleDevToken)
		if firestoreMediaDataErr != nil {
			http.Error(writer, firestoreMediaDataErr.Error(), http.StatusInternalServerError)
			log.Printf("GetAppleMusicMedia [GetAppleMusicAlbum Switch]: %v", firestoreMediaDataErr)
			return
		}
	default:
		http.Error(writer, appleMusicMediaReq.MediaType+" media type does not exist", http.StatusInternalServerError)
		log.Printf("GetAppleMusicMedia [MediaTypeSwitch]: %v media type does not exist", appleMusicMediaReq.MediaType)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(firestoreMediaData)
}
