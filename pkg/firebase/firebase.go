package firebase

// no_neon_one - "go to GO or no to GO" (03/01/20)
// no_neon_one - "I think Microsoft named .Net so it wouldn’t show up in a Unix directory listing (by Oktal )." (03/08/20)
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	firestore "cloud.google.com/go/firestore"
)

// -- FIRESTORE -- //

// FirestoreUser respresents the data stored in Firestore for an user
type FirestoreUser struct {
	Email                   string                   `firestore:"email" json:"email"`
	ID                      string                   `firestore:"id" json:"id"`
	Playlists               []*firestore.DocumentRef `firestore:"playlists" json:"playlists"`
	PreferredSocialPlatform *firestore.DocumentRef   `firestore:"preferredSocialPlatform" json:"preferredSocialPlatform"`
	ProfileImage            *SpotifyImage            `firestore:"profileImage" json:"profileImage"`
	SocialPlatforms         []*firestore.DocumentRef `firestore:"socialPlatforms" json:"socialPlatforms"`
	DisplayName             string                   `firestore:"displayName" json:"displayName"`
	Username                string                   `firestore:"username" json:"username"`
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

// FirestoreMedia represents the "song" item stored in the songs collection. We use media because techincally this doens't have to be just a song
type FirestoreMedia struct {
	ID      string                     `firestore:"id" json:"id"`
	Name    string                     `firestore:"name" json:"name"`
	Album   string                     `firestore:"album,omitempty" json:"album,omitempty"`
	Type    string                     `firestore:"type" json:"type"`
	Creator string                     `firestore:"creator" json:"creator"`
	Apple   FirestoreMediaPlatformData `firestore:"apple,omitempty" json:"apple,omitempty"`
	Spotify FirestoreMediaPlatformData `firestore:"spotify,omitempty" json:"spotify,omitempty"`
	YouTube FirestoreMediaPlatformData `firestore:"youtube,omitempty" json:"youtube,omitempty"`
}

// FirestoreMediaPlatformData includes the specific data for the current media based on platform
type FirestoreMediaPlatformData struct {
	ID     string      `firestore:"id" json:"id"`
	URL    string      `firestore:"url" json:"url"`
	Images interface{} `firestore:"images" json:"images"`
}

// FirestoreAppleDevJWT rerpresents the JWT object that is needed for calling Apple Music API stuff
type FirestoreAppleDevJWT struct {
	ExpiresAt int64  `json:"expiresAt"`
	IssuedAt  int64  `json:"issuedAt"`
	Token     string `json:"token"`
}

// FirestoreSpotifyAuthToken represents the auth token object needed for calling Spotify APIs
type FirestoreSpotifyAuthToken struct {
	ExpiredAt string `json:"expiredAt"`
	ExpiresIn int    `json:"expiresIn"`
	IssuedAt  string `json:"issuedAt"`
	Token     string `json:"token"`
}

// FirestorePlaylist represents the data for a playlist store in Firestore
type FirestorePlaylist struct {
	ID        string                   `firestore:"id" json:"id"`
	Name      string                   `firestore:"name" json:"name"`
	CreatedBy *firestore.DocumentRef   `firestore:"createdBy" json:"createdBy"`
	Members   []*firestore.DocumentRef `firestore:"members" json:"members"`
	Songs     PlaylistSongs            `firestore:"songs" json:"songs"`
	Comments  interface{}              `firestore:"comments" json:"comments"` // This will actually need be an object with key:value pair of songId:[Comments]
	CoverArt  string                   `firestore:"coverArt" json:"coverArt"`
}

// PlaylistSongs represents the songs object in a playlist which includes an AddedBy object and AllSongs DocRef Array
type PlaylistSongs struct {
	AddedBy  map[string][]string      `firestore:"addedBy" json:"addedBy"`
	AllSongs []*firestore.DocumentRef `firestore:"allSongs" json:"allSongs"`
}

// FirestoreEvent implements the Firestore event from a trigger function
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue implements the values that come from a Firestore event
type FirestoreValue struct {
	CreateTime time.Time          `json:"createTime"`
	Fields     FirestoreEventUser `json:"fields"`
	Name       string             `json:"name"`
	UpdateTime time.Time          `json:"updateTime"`
}

// FirestoreEventUser implements the values that are recevied from a firestore trigger
type FirestoreEventUser struct {
	ID           stringValue          `json:"id"`
	Email        stringValue          `json:"email"`
	ProfileImage profileImageMapValue `json:"profileImage"`
	DisplayName  stringValue          `json:"displayName"`
	Username     stringValue          `json:"username"`
}

// FirestoreEventProfileImage implements data needed for firestore image
type FirestoreEventProfileImage struct {
	Width  integerValue `json:"width,omitempty"`
	Height integerValue `json:"height,omitempty"`
	URL    stringValue  `json:"url,omitempty"`
}

// profileImageMapValue implements the profileImage nested object value type for Firestore Event
type profileImageMapValue struct {
	MapValue struct {
		Fields FirestoreEventProfileImage `json:"fields"`
	} `json:"mapValue"`
}

// stringintegerValue implements the integer value type for Firestore Events
type integerValue struct {
	IntegerValue int `json:"integerValue"`
}

// stringValue implements the string value type for Firestore Events
type stringValue struct {
	StringValue string `json:"stringValue"`
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

// -- HELPER FUNCTIONS -- //

// GetAppleDeveloperToken will go to Firestore and grab developer token if it exists and will refresh it if expired
func GetAppleDeveloperToken() (*FirestoreAppleDevJWT, error) {
	var hostname string
	httpClient := &http.Client{}
	if os.Getenv("ENVIRONMENT") == "DEV" {
		hostname = os.Getenv("HOSTNAME_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		hostname = os.Getenv("HOSTNAME_PROD")
	}

	// Create URI
	createAppleDevURI := hostname + "/createAppleDevToken"
	appleDevTokenReq, appleDevTokenReqErr := http.NewRequest("GET", createAppleDevURI, nil)
	if appleDevTokenReqErr != nil {
		return nil, fmt.Errorf("[GetAppleDeveloperToken]: %v", appleDevTokenReqErr)
	}

	// Call endpoint
	appleDevTokenResp, appleDevTokenRespErr := httpClient.Do(appleDevTokenReq)
	if appleDevTokenRespErr != nil {
		return nil, fmt.Errorf("[GetAppleDeveloperToken]: %v", appleDevTokenRespErr)
	}

	// Decode response to AppleDevJWT
	var appleDevToken FirestoreAppleDevJWT
	appleDevTokenDecodeErr := json.NewDecoder(appleDevTokenResp.Body).Decode(&appleDevToken)
	if appleDevTokenDecodeErr != nil {
		return nil, fmt.Errorf("[GetAppleDeveloperToken]: %v", appleDevTokenDecodeErr)
	}

	// Return JWT
	return &appleDevToken, nil
}
