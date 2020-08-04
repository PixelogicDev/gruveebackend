package getapplemusicmedia

import (
	"time"
)

// -- RESPONSES -- //

// appleMusicTrackResp defines the data returned and needed from the Apple Music Get Track API
type appleMusicTrackResp struct {
	Data []appleMusicTrackData `json:"data"`
}

// appleMusicPlaylistResp defines the data returned and needed from the Apple Music Get Playlist API
type appleMusicPlaylistResp struct {
	Data []appleMusicPlaylistData `json:"data"`
}

// appleMusicAlbumResp defines the data retuned and needed from the Apple Music Get Album API
type appleMusicAlbumResp struct {
	Data []appleMusicAlbumData `json:"data"`
}

// -- DATA -- //

// appleMusicTrackData defines the track data inside the data array response
type appleMusicTrackData struct {
	Attributes struct {
		AlbumName   string            `json:"albumName"`
		ArtistName  string            `json:"artistName"`
		Artwork     appleMusicArtwork `json:"artwork"`
		TrackName   string            `json:"name"`
		ExternalURL string            `json:"url"`
	} `json:"attributes"`
}

// appleMusicPlaylistData defines the playlist data inside the data array response
type appleMusicPlaylistData struct {
	Attributes struct {
		Artwork          appleMusicArtwork        `json:"artwork"`
		CuratorName      string                   `json:"curatorName"`
		Description      appleMusicEditorialNotes `json:"description"`
		LastModifiedDate time.Time                `json:"lastModifiedDate"`
		Name             string                   `json:"name"`
		PlaylistType     appleMusicPlaylistType   `json:"playlistType"`
		URL              string                   `json:"url"`
	} `json:"attributes"`
	Relationships struct {
		Tracks appleMusicTrackRelationship `json:"tracks"`
	} `json:"relationships"`
	Type string `json:"type"`
}

// appleMusicAlbumData defines the album data inside the data array response
type appleMusicAlbumData struct {
	Attributes struct {
		AlbumName           string                   `json:"albumName"`
		ArtistName          string                   `json:"artistName"`
		Artwork             appleMusicArtwork        `json:"artwork"`
		ContentRating       string                   `json:"contentRating"`
		Copyright           string                   `json:"copyright"`
		EditorialNotes      appleMusicEditorialNotes `json:"editorialNotes"`
		GenreNames          []string                 `json:"genreNames"`
		IsComplete          bool                     `json:"isComplete"`
		IsSingle            bool                     `json:"isSingle"`
		Name                string                   `json:"name"`
		RecordLabel         string                   `json:"recordLabel"`
		ReleaseDate         string                   `json:"releaseDate"`
		TrackCount          int                      `json:"trackCount"`
		URL                 string                   `json:"url"`
		IsMasteredForItunes bool                     `json:"isMasteredForItunes"`
	} `json:"attributes"`

	Relationships struct {
		Tracks appleMusicTrackRelationship `json:"tracks"`
	} `json:"relationships"`
}

// appleMusicVideoData defines the music video data
type appleMusicVideoData struct {
	Attributes struct {
		AlbumName      string                   `json:"albumName"`
		ArtistName     string                   `json:"artistName"`
		Artwork        appleMusicArtwork        `json:"artwork"`
		ContentRating  string                   `json:"contentRating"`
		DurationInMS   float64                  `json:"durationInMillis"`
		EditorialNotes appleMusicEditorialNotes `json:"editorialNotes"`
		GenreNames     []string                 `json:"genreNames"`
		ISRC           string                   `json:"isrc"`
		Name           string                   `json:"name"`
		ReleaseDate    time.Time                `json:"releaseDate"`
		TrackNumber    int                      `json:"trackNumber"`
		URL            string                   `json:"url"`
		VideoSubType   string                   `json:"videoSubType"`
		HasHDR         bool                     `json:"hadHDR"`
		Has4k          bool                     `json:"has4k"`
	} `json:"attributes"`
}

// -- DATA TYPES -- //

// appleMusicPlaylistType defines the different possible playlist types
type appleMusicPlaylistType string

// appleMusicArtwork defines the artwork properties of Apple Music media
type appleMusicArtwork struct {
	BGColor    string `json:"bgColor" firestore:"bgColor"`
	Height     int    `json:"height" firestore:"height"`
	TextColor1 string `json:"textColor1,omitempty" firestore:"textColor1,omitempty"`
	TextColor2 string `json:"textColor2,omitempty" firestore:"textColor2,omitempty"`
	TextColor3 string `json:"textColor3,omitempty" firestore:"textColor3,omitempty"`
	TextColor4 string `json:"textColor4,omitempty" firestore:"textColor4,omitempty"`
	URL        string `json:"url" firestore:"url"`
	Width      int    `json:"width" firestore:"width"`
}

// appleMusicEditorialNotes defines the description of a playlist
type appleMusicEditorialNotes struct {
	Short    string `json:"short"`
	Standard string `json:"standard"`
}

// appleMusicTrackRelationship defines the data the content of a playlist
// Data here can be appleMusicTrackData or appleMusicVideoData
type appleMusicTrackRelationship struct {
	Data []interface{} `json:"data"`
}

// -- CONSTANTS -- //
// The different options possible for a PlaylistType
const (
	userShared  appleMusicPlaylistType = "user-shared"
	editorial   appleMusicPlaylistType = "editorial"
	external    appleMusicPlaylistType = "external"
	personalMix appleMusicPlaylistType = "personal-mix"
)
