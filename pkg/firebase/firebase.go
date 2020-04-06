package firebase

// no_neon_one - "go to GO or no to GO" (03/01/20)
// no_neon_one - "I think Microsoft named .Net so it wouldn’t show up in a Unix directory listing (by Oktal )." (03/08/20)
import (
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
	ProfileImage            SpotifyImage             `firestore:"profileImage" json:"profileImage"`
	SocialPlatforms         []*firestore.DocumentRef `firestore:"socialPlatforms" json:"socialPlatforms"`
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
