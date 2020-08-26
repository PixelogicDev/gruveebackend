package getspotifymedia

import "github.com/pixelogicdev/gruveebackend/pkg/firebase"

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
	ID     string                  `json:"id"`
	Name   string                  `json:"name"`
	Type   string                  `json:"type"`
	Images []firebase.SpotifyImage `json:"images"`
	Owner  struct {
		DisplayName string `json:"display_name"`
	} `json:"owner"`
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
