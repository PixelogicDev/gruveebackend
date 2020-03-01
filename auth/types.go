// no_neon_one - "go to GO or no to GO" (03/01/20)
package auth

import (
	firestore "cloud.google.com/go/firestore"
)

// SpotifyAuthRequest includes APIToken needed for Spotify API
type SpotifyAuthRequest struct {
	APIToken string `json:"token"`
}

// SpotifyMeResponse represents the response coming back from the /me endpoint
type SpotifyMeResponse struct {
	DisplayName string         `json:"display_name"`
	Email       string         `json:"email"`
	ID          string         `json:"id"`
	Images      []SpotifyImage `json:"images"`
	Product     string         `json:"product"`
}

// SpotifyRequestError represents the Spotify Error Object
type SpotifyRequestError struct {
	Error spotifyRequestErrorDetails `json:"error"`
}

// SpotifyRequestErrorDetails represents the Spotify Error Object details
type spotifyRequestErrorDetails struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// AuthorizeWithSpotifyResponse represents the data to send back to the client for a user
type AuthorizeWithSpotifyResponse struct {
	Email                   string                    `json:"email"`
	ID                      string                    `json:"id"`
	Playlists               []string                  `json:"playlists"`
	PreferredSocialPlatform FirestoreSocialPlatform   `json:"preferredSocialPlatform"`
	SocialPlatforms         []FirestoreSocialPlatform `json:"socialPlatforms"`
	Username                string                    `json:"username"`
}

// FirestoreUser respresents the data stored in Firestore for an user
type FirestoreUser struct {
	Email                   string                   `firestore:"email"`
	ID                      string                   `firestore:"id"`
	Playlists               []string                 `firestre:"playlists"`
	PreferredSocialPlatform *firestore.DocumentRef   `firestore:"preferredSocialPlatform"`
	SocialPlatforms         []*firestore.DocumentRef `firestore:"socialPlatforms"`
	Username                string                   `firestore:"username"`
}

// FirestoreSocialPlatform represents the data for a social platform stored in Firestore
type FirestoreSocialPlatform struct {
	APIToken           string       `firestore:"apiToken" json:"apiToken"`
	RefreshToken       string       `firestore:"refreshToken" json:"refreshToken"`
	Email              string       `firestore:"email" json:"email"`
	ID                 string       `firestore:"id" json:"id"`
	IsPreferredService bool         `firestore:"isPreferredService" json:"isPreferredService"`
	IsPremium          bool         `firestore:"isPremium" json:"isPremium"`
	PlatformName       string       `firestore:"platformName" json:"platformName"`
	ProfileImage       SpotifyImage `firestore:"profileImage" json:"profileImage"`
	Username           string       `firestore:"username" json:"username"`
}

// SpotifyImage includes data for any image returned in Spotify
type SpotifyImage struct {
	Height int    `json:",omitempty"`
	URL    string `json:"url"`
	Width  int    `json:",omitempty"`
}