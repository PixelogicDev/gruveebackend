package social

import (
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
)

// -- SPOTIFY AUTH -- //

// SpotifyAuthRequest includes APIToken needed for Spotify API
type SpotifyAuthRequest struct {
	APIToken     string `json:"token"`
	ExpiresIn    int    `json:"expiresIn"`
	RefreshToken string `json:"refreshToken"`
}

// SpotifyClientCredsAuthResp includes the response for a client credentials flow from Spotify
type SpotifyClientCredsAuthResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// SpotifyMeResponse represents the response coming back from the /me endpoint
type SpotifyMeResponse struct {
	DisplayName string                  `json:"display_name"`
	Email       string                  `json:"email"`
	ID          string                  `json:"id"`
	Images      []firebase.SpotifyImage `json:"images"`
	Product     string                  `json:"product"`
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
	Email                   string                             `json:"email"`
	ID                      string                             `json:"id"`
	Playlists               []firebase.FirestorePlaylist       `json:"playlists"`
	PreferredSocialPlatform firebase.FirestoreSocialPlatform   `json:"preferredSocialPlatform"`
	SocialPlatforms         []firebase.FirestoreSocialPlatform `json:"socialPlatforms"`
	Username                string                             `json:"username"`
	JWT                     string                             `json:"jwt,omitempty"`
}

// -- APPLE MUSIC AUTH -- //

// AppleMusicRequestError represents the Apple Music Error Object
type AppleMusicRequestError struct {
	Errors []appleMusicError `json:"errors"`
}

// appleMusicError represents the Apple Music Error object data
type appleMusicError struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
	ID     string `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

// -- GENERATE TOKEN -- //

// GenerateTokenRequest represents the UID for the user that we want to create a custom token for
type GenerateTokenRequest struct {
	UID string
}

// GenerateTokenResponse represents what we will send back to the client
type GenerateTokenResponse struct {
	Token string `json:"token"`
}

// -- REFRESH TOKEN -- //

// TokenRefreshRequest includes uid to grab all social platforms for user
type TokenRefreshRequest struct {
	UID string `json:"uid"`
}

// RefreshTokensResponse contains a list of refreshed tokens for multiple social platforms
type RefreshTokensResponse struct {
	RefreshTokens map[string]firebase.APIToken `json:"refreshTokens"`
}

// RefreshToken contains the generic information for a refresh token for social platform
type RefreshToken struct {
	PlatformName string `json:"platformName"`
	APIToken     string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
}

// -- CREATE USER -- //

// CreateUserReq is the payload that includes the minimal amount of data to create a user
type CreateUserReq struct {
	Email              string                 `json:"email"`
	ID                 string                 `json:"id"`
	SocialPlatformPath string                 `json:"socialPlatformPath"`
	ProfileImage       *firebase.SpotifyImage `json:"profileImage"`
	DisplayName        string                 `json:"displayName"`
	Username           string                 `json:"username"`
}

// -- COMMON GET MEDIA -- //

// GetMediaReq is the payload that inclues the provider, mediaId, and mediaType for finding media service
type GetMediaReq struct {
	Provider  string `json:"provider"`
	MediaID   string `json:"mediaId"`
	MediaType string `json:"mediaType"`
	// This is only an Apple Music property so remove if not passed through
	Storefront string `json:"storefront,omitempty"`
}
