package fetchallmedia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/mediahelpers"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

const (
	spotifyHostName = "https://api.spotify.com/v1/search"
)

var httpClient *http.Client
var firestoreClient *firestore.Client

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

// generateRequest creats the request object to call any query API
func generateRequest(uri string, method string, token string) (*http.Request, error) {
	// Generate request
	queryReq, queryReqErr := http.NewRequest(method, uri, nil)
	if queryReqErr != nil {
		return nil, fmt.Errorf("FetchAllMedia [http.NewRequest]: %v", queryReqErr)
	}

	// Add headers
	queryReq.Header.Add("Authorization", "Bearer "+token)
	return queryReq, nil
}

// querySpotifyMedia will use Spotify search API to pull back the data for the specified media
func querySpotifyMedia(mediaName string, mediaCreator string, mediaType string) (*io.ReadCloser, error) {
	var mediaNameQuery string

	switch mediaType {
	case "track":
		mediaNameQuery = "track:" + "\"" + mediaName + "\""
	case "album":
		mediaNameQuery = "album:" + "\"" + mediaName + "\""
	}

	// Add queries
	query := url.Values{}
	query.Add("q", mediaNameQuery+"+artist:"+"\""+mediaCreator+"\"")
	query.Add("type", mediaType)
	query.Add("limit", "5")

	// Fetch spotify API token
	authToken, authTokenErr := mediahelpers.FetchSpotifyAuthToken(*firestoreClient)
	if authTokenErr != nil {
		return nil, fmt.Errorf("Spotify fetch auth token error: %v", authTokenErr)
	}

	// Setup request
	queryReq, queryReqErr := generateRequest(spotifyHostName, "GET", authToken.Token)
	if queryReqErr != nil {
		return nil, fmt.Errorf("Spotify generate request error: %v", queryReqErr)
	}

	// Add queries to request
	queryReq.URL.RawQuery = query.Encode()

	// Call endpoint
	queryResp, queryRespErr := httpClient.Do(queryReq)
	if queryRespErr != nil {
		return nil, fmt.Errorf("Spotify query error: %v", queryRespErr)
	}

	// Check if request passed back error in the body
	if queryResp.StatusCode != http.StatusOK {
		// Convert Spotify Error Object
		var spotifyErrorObj social.SpotifyRequestError

		err := json.NewDecoder(queryResp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			return nil, fmt.Errorf("Spotify Request Decoder: %v", err)
		}

		return nil, fmt.Errorf("Spotify Track Request: %v", spotifyErrorObj.Error.Message)
	}

	return &queryResp.Body, nil
}

// writeData takes the song object data and writes it to Firestore
func writeData(dataBlob map[string]interface{}, path string) error {
	docPath := firestoreClient.Doc(path)
	if docPath == nil {
		return fmt.Errorf("DocPath does not exist: %s", path)
	}

	_, writeErr := docPath.Set(context.Background(), dataBlob, firestore.MergeAll)
	if writeErr != nil {
		return fmt.Errorf(writeErr.Error())
	}

	return nil
}
