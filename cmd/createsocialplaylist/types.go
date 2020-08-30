package createsocialplaylist

import "github.com/pixelogicdev/gruveebackend/pkg/firebase"

// createSocialPlaylistRequest includes the socialPlatform and playlist that will be added
type createSocialPlaylistRequest struct {
	SocialPlatform firebase.FirestoreSocialPlatform `json:"socialPlatform"`
	PlaylistName   string                           `json:"playlistName"`
}

// createSocialPlaylistResponse includes the refreshToken for the platform if there is one
type createSocialPlaylistResponse struct {
	PlatformName string            `json:"platformName"`
	RefreshToken firebase.APIToken `json:"refreshToken"`
}

// appleMusicPlaylistRequest includes the payload needed to create an Apple Music Playlist
type appleMusicPlaylistRequest struct {
	Attributes struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"attributes"`
}

// spotifyPlaylistRequest includes the payload needed to create a Spotify Playlist
type spotifyPlaylistRequest struct {
	Name          string `json:"name"`
	Public        bool   `json:"public"`
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
}
