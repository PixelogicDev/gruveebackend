package fetchallmedia

import "github.com/pixelogicdev/gruveebackend/pkg/mediahelpers"

// spotifyQueryResp includes the data from a search request to Spotify
type spotifyQueryResp struct {
	Albums spotifyQueryRespAlbums `json:"albums,omitempty"`
	Tracks spotifyQueryRespTracks `json:"tracks,omitempty"`
}

// spotifyQueryRespTracks includes the list of track data
type spotifyQueryRespTracks struct {
	Items  []mediahelpers.SpotifyTrackData `json:"items"`
	Limit  int                             `json:"limit"`
	Offset int                             `json:"offset"`
	Total  int                             `json:"total"`
}

// spotifyQueryRespAlbums includes the list of album data
type spotifyQueryRespAlbums struct {
	Items  []mediahelpers.SpotifyAlbumData `json:"items"`
	Limit  int                             `json:"limit"`
	Offset int                             `json:"offset"`
	Total  int                             `json:"total"`
}
