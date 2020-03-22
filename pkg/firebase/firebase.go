package firebase

// no_neon_one - "go to GO or no to GO" (03/01/20)
// no_neon_one - "I think Microsoft named .Net so it wouldn’t show up in a Unix directory listing (by Oktal )." (03/08/20)
import (
	firestore "cloud.google.com/go/firestore"
)

// -- TYPES -- //

// -- SPOTIFY AUTH -- //

// SpotifyAuthRequest includes APIToken needed for Spotify API
type SpotifyAuthRequest struct {
	APIToken     string `json:"token"`
	ExpiresIn    int    `json:"expiresIn"`
	RefreshToken string `json:"refreshToken"`
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
	Playlists               []FirestorePlaylist       `json:"playlists"`
	PreferredSocialPlatform FirestoreSocialPlatform   `json:"preferredSocialPlatform"`
	SocialPlatforms         []FirestoreSocialPlatform `json:"socialPlatforms"`
	Username                string                    `json:"username"`
	JWT                     string                    `json:"jwt,omitempty"`
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

// -- FIRESTORE -- //

// FirestoreUser respresents the data stored in Firestore for an user
type FirestoreUser struct {
	Email                   string                   `firestore:"email"`
	ID                      string                   `firestore:"id"`
	Playlists               []*firestore.DocumentRef `firestore:"playlists"`
	PreferredSocialPlatform *firestore.DocumentRef   `firestore:"preferredSocialPlatform"`
	SocialPlatforms         []*firestore.DocumentRef `firestore:"socialPlatforms"`
	Username                string                   `firestore:"username"`
}

// FirestoreSocialPlatform represents the data for a social platform stored in Firestore
type FirestoreSocialPlatform struct {
	APIToken           APIToken     `firestore:"apiToken" json:"apiToken"`
	RefreshToken       string       `firestore:"refreshToken" json:"refreshToken"`
	Email              string       `firestore:"email" json:"email"`
	ID                 string       `firestore:"id" json:"id"`
	IsPreferredService bool         `firestore:"isPreferredService" json:"isPreferredService"`
	IsPremium          bool         `firestore:"isPremium" json:"isPremium"`
	PlatformName       string       `firestore:"platformName" json:"platformName"`
	ProfileImage       SpotifyImage `firestore:"profileImage" json:"profileImage"`
	Username           string       `firestore:"username" json:"username"`
}

// FirestorePlaylist represents the data for a playlist store in Firestore
type FirestorePlaylist struct {
	ID        string      `firestore:"id" json:"id"`
	Name      string      `firestore:"name" json:"name"`
	CreatedBy string      `firestore:"createdBy" json:"createdBy"`
	Members   []string    `firestore:"members" json:"members"`
	Songs     []string    `firestore:"songs" json:"songs"`
	Comments  interface{} `firestore:"comments" json:"comments"` // This will actually need be an object with key:value pair of songId:[Comments]
	CoverArt  string      `firestore:"coverArt" json:"coverArt"`
}

// -- FIRESTORE TYPES -- //

// APIToken contains the access token and the time in which it expires
type APIToken struct {
	CreatedAt string `firestore:"createdAt" json:"createdAt"`
	ExpiredAt string `firestore:"expiredAt" json:"expiredAt"`
	Token     string `firestore:"token" json:"token"`
	ExpiresIn int    `firestore:"expiresIn" json:"expiresIn"`
}

// SpotifyImage includes data for any image returned in Spotify
type SpotifyImage struct {
	Height int    `firestore:"height,omitempty" json:"height,omitempty"`
	URL    string `firestore:"url" json:"url"`
	Width  int    `firestore:"width,omitempty" json:"width,omitempty"`
}
