package fetchallmedia

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

var (
	logger sawmill.Logger
)

// FetchAllMedia queries all music providers for songs data when a new document is added
func FetchAllMedia(ctx context.Context, event firebase.FirestoreEventSongs) error {
	logger.Log("FetchAllMedia", fmt.Sprintf("Event received: %v", event))

	initErr := initWithEnv()
	if initErr != nil {
		logger.LogErr("InitWithErr", initErr, nil)
		return initErr
	}

	// Init data
	var appleData *map[string]interface{}
	var spotifyData *map[string]interface{}
	var youtubeData *map[string]interface{}

	// Get media name & creator
	mediaName := event.Value.Fields.Name.StringValue
	mediaCreator := event.Value.Fields.Creator.StringValue
	mediaType := event.Value.Fields.Type.StringValue
	docPath := strings.Split(event.Value.Name, "documents/")

	logger.Log("FetchAllMedia", fmt.Sprintf("MediaName: %v", mediaName))
	logger.Log("FetchAllMedia", fmt.Sprintf("MediaCreator: %v", mediaCreator))
	logger.Log("FetchAllMedia", fmt.Sprintf("MediaType: %v", mediaType))
	logger.Log("FetchAllMedia", fmt.Sprintf("DocPath: %v", docPath))

	// Get media name and creator
	// This can be a album or playlist or song
	// We 1000% cannot get the playlist in another platform (until we start mapping the songs from a playlist)
	if mediaType == "playlist" {
		logger.Log("FetchAllMedia", "Media is a playlist, we don't need to run a check.")
		return nil
	}

	// Check each provider to see if it exists. If not, go query for that media
	if event.Value.Fields.Apple == (firebase.MediaMapValue{}) {
		// Call Apple Music Query with data
		logger.Log("FetchAllMedia", "Getting media for Apple Music...")
	}

	if event.Value.Fields.Spotify == (firebase.MediaMapValue{}) {
		// Call Spotify Query with data
		logger.Log("FetchAllMedia", "Getting media for Spotify...")
		data, queryErr := querySpotifyMedia(mediaName, mediaCreator, mediaType)
		if queryErr != nil {
			logger.LogErr("QuerySpotifyMedia", queryErr, nil)
			return queryErr
		}

		// Decode
		var spotifyQueryData spotifyQueryResp
		json.NewDecoder(*data).Decode(&spotifyQueryData)

		logger.Log("FetchAllMedia", fmt.Sprintf("SpotifyQueryData decoded: %v", spotifyQueryData))

		if len(spotifyQueryData.Tracks.Items) != 0 {
			logger.Log("FetchAllMedia", fmt.Sprintf("Found tracks: %v", spotifyQueryData.Tracks.Items[0]))
			// Grab first track item & create song object
			track := spotifyQueryData.Tracks.Items[0]
			spotifyData = &map[string]interface{}{
				"id":     track.ID,
				"images": track.Album.Images,
				"url":    track.ExternalURLs.Spotify,
			}
		} else if len(spotifyQueryData.Albums.Items) != 0 {
			logger.Log("FetchAllMedia", fmt.Sprintf("Found albums: %v", spotifyQueryData.Albums.Items[0]))
			// Grab first track item & create song object
			album := spotifyQueryData.Albums.Items[0]
			spotifyData = &map[string]interface{}{
				"id":     album.ID,
				"images": album.Images,
				"url":    album.ExternalURLs.Spotify,
			}
		}
	}

	if event.Value.Fields.YouTube == (firebase.MediaMapValue{}) {
		// Call Youtube Query with data
		logger.Log("FetchAllMedia", "Getting media for YouTube Music...")
	}

	// Write data to song document and check if it changed
	dataBlob := make(map[string]interface{})

	if appleData != nil {
		dataBlob["apple"] = appleData
	}

	if spotifyData != nil {
		dataBlob["spotify"] = spotifyData
	}

	if youtubeData != nil {
		dataBlob["youtube"] = youtubeData
	}

	if len(docPath) == 0 {
		error := fmt.Errorf("DocPath split was empty")
		logger.LogErr("DocPath Split", error, nil)
		return error
	}

	logger.Log("FetchAllMedia", fmt.Sprintf("Writing blob to Firestore: %v", dataBlob))

	// Write data to db
	writeDataErr := writeData(dataBlob, docPath[1])
	if writeDataErr != nil {
		logger.LogErr("WriteData", writeDataErr, nil)
		return writeDataErr
	}

	logger.Log("FetchAllMedia", "Data written successfully.")

	return nil
}
