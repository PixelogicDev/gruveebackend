# Grüvee Functions Documentation

### This file contains a list of all the functions we have, what they do, and their requests and responses.

## Functions
### **AuthorizeWithApple** - */authorizeWithApple*
This function has a server side generated template that is used to authenticate a user with their Apple Music account. Once authenticated it will store the Apple Music User Token as part of the user’s document in Firebase. We use that token to do things in the user’s Apple Music account such as, creating a new playlist, adding, etc.

#### Request
Nothing included

#### Response
`200` -> Template rendered properly

`500` -> Error rendering template



### **CreateAppleDevToken** - */createAppleDevToken*
This function is used to generate a new Apple Developer Token. This token is used to authenticate any of our calls to Apple Music API. It will check to see if the token exists and if it needs to be refreshed. Any changes to the token will be written back into the `internal_tokens` collection. 

#### Request
Nothing included

#### Response
`200` 
```go
struct {
    ExpiresAt int64  `json:"expiresAt"`
    IssuedAt  int64  `json:"issuedAt"`
    Token     string `json:"token"`
}
```

`500` -> Any other errors


### **CreateProviderUser** - */createProviderUser*
CreateProviderUser will check to see if the newly created user needs to be added to the providers_users collection. This function is only called for platforms that are supported natively through Firebase. This platforms like Spotify do not have a way to generate UIDs directly from Firebase so we are able to give the Firebase user any UID we want. While on the other hand, built in Firebase auth platforms give you a UID that you are not able to change. This collection links the Firebase UID to the actual platform UID. For example `FirebaseId{1234567890} -> ApplePlatform{apple:000000000}` 

#### Request
```go
type createProviderUserReq struct {
    FirebaseProviderUID string `json:"firebaseProviderUID"`
    PlatformProviderUID string `json:"platformProviderUID"`
}
```

#### Response
`200` -> Successful write to Firestore

`500` -> Any server errors



### **CreateSocialPlaylist** - */createSocialPlaylist*
CreateSocialPlaylist will take in a SocialPlatform and will go create a playlist on the social account itself. For example, if I am logged in as a Spotify user and create a new playlist in Grüvee, this function will then fire off to create that same playlist in the user’s Spotify account.

#### Request
```go
type createSocialPlaylistRequest struct {
    SocialPlatform firebase.FirestoreSocialPlatform `json:"socialPlatform"`
    PlaylistName   string                           `json:"playlistName"`
}
```

#### Response
`200`
```go
type createSocialPlaylistResponse struct {
    PlatformName string            `json:"platformName"`
    RefreshToken firebase.APIToken `json:"refreshToken"`
}
```

`204` -> Successful playlist creation & no token refresh

`500` -> Any server errors


### **CreateUser** */createUser*
CreateUser will write a new Firebase user to Firestore. This function is used to create the actual Firestore Document.

#### Request
```go
type CreateUserReq struct {
    Email              string                 `json:"email"`
    ID                 string                 `json:"id"`
    SocialPlatformPath string                 `json:"socialPlatformPath"`
    ProfileImage       *firebase.SpotifyImage `json:"profileImage"`
    DisplayName        string                 `json:"displayName"`
    Username           string                 `json:"username"`
}
```


#### Response
`200`
```go
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
```

`500` -> Any server errors


### **DoesUserDocExist** - */doesUserDocExist*
DoesUserDocExist checks to see if there is already a Firebase user document for someone right before they sign in. This was created to pull back a user document if they already have an account with Grüvee

#### Request
```go
type doesUserDocExistReq struct {
    UID string `json:"uid"`
}
```

#### Response
`200`
```go
type doesUserDocExistResp struct {
    Result bool `json:"result"`
}
```

`500` -> Any server errors

### **FetchAllMedia** - */FetchAllMedia*
FetchAllMedia queries all music providers for songs data when a new document is added. In the background, this thing will fire off so the song document will include all the information for all supported platforms.

#### Request
Song document

#### Response
Returns `Error` or `nil` 

### **GetAppleMusicMedia** - */getAppleMusicMedia*
GetAppleMusicMedia will take in Apple media data and get the exact media from Apple Music API. We need to utilize the AppleDeveloperToken in the header to make the proper call. We can get a track, album, or playlist from Apple Music

#### Request
```go
type GetMediaReq struct {
    Provider  string `json:"provider"`
    MediaID   string `json:"mediaId"`
    MediaType string `json:"mediaType"`
    // This is only an Apple Music property so remove if not passed through
    Storefront string `json:"storefront,omitempty"`
}
```


#### Response

`200` -> `interface{}` depends on the media that is returned

`500` -> Any errors in function


### **GetSpotifyMedia** - */getSpotifyMedia*
GetSpotifyMedia will take in Spotify media data and get the exact media from Spotify API. We need to utilize the SpotifyAuthToken in the header to make the proper call. We can get track, album, or playlist from Spotify

#### Request
```go
type GetMediaReq struct {
    Provider  string `json:"provider"`
    MediaID   string `json:"mediaId"`
    MediaType string `json:"mediaType"`
    // This is only an Apple Music property so remove if not passed through
    Storefront string `json:"storefront,omitempty"`
}
```

#### Response

`200` -> `interface{}` depends on the media that is returned

`500` -> Any errors in function


### **CreateSocialPlatform** - */createSocialPlatform*
CreateSocialPlatform will write a new social platform to firestore. This is used when a new user is created and we need to generate an object that holds all the platforms that the user has authenticated with.

#### Request
```go
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
```


#### Response
`200` -> Successfully created and wrote platform to Firestore

`500` -> Any function errors


### **SocialTokenRefresh** - */socialTokenRefresh*
SocialTokenRefresh checks to see if we need to refresh current API tokens for social platforms. This was currently created to check all social platforms for expired auth tokens, but currently only really works with Spotify. This is because Apple Music doesn’t have a refresh method currently and YouTube music is not currently implemented.

#### Request
```go
type TokenRefreshRequest struct {
    UID string `json:"uid"`
}
```

#### Response
`200`
```go
type RefreshTokensResponse struct {
    RefreshTokens map[string]firebase.APIToken `json:"refreshTokens"`
}
```

`500` -> Any function errors

### **AuthorizeWithSpotify** - */spotifyAuth*
AuthorizeWithSpotify will verify Spotify creds are valid and return any associated Firebase user or create a new Firebase user. This is a big function that handles all the authentication needed to connect a user’s Spotify account to Grüvee.

#### Request
```go
type SpotifyAuthRequest struct {
    APIToken     string `json:"token"`
    ExpiresIn    int    `json:"expiresIn"`
    RefreshToken string `json:"refreshToken"`
}
```

#### Response

`200`
```go
type AuthorizeWithSpotifyResponse struct {
    Email                   string                             `json:"email"`
    ID                      string                             `json:"id"`
    Playlists               []firebase.FirestorePlaylist       `json:"playlists"`
    PreferredSocialPlatform firebase.FirestoreSocialPlatform   `json:"preferredSocialPlatform"`
    SocialPlatforms         []firebase.FirestoreSocialPlatform `json:"socialPlatforms"`
    Username                string                             `json:"username"`
    JWT                     string                             `json:"jwt,omitempty"`
}
```

`500` -> Any function errors


### **GenerateCustomToken** - */generateCustomToken*
GenerateCustomToken generates a CustomToken for Firebase Login. This is needed when signing into platforms that we have to implement ourselves, such as Spotify.

#### Request
```go
type GenerateTokenRequest struct {
    UID string
}
```

#### Response
`200`
```go
type GenerateTokenResponse struct {
    Token string `json:"token"`
}
```

`500` -> Any function errors


### **UpdateAlgolia** - */UpdateAlgolia*
UpdateAlgolia sends new data to Algolia service for indexing. This trigger function is called when there is a new user document created. Algolia is used for quick searching via username for Grüvee users.

#### Request
User Document

#### Response
Error or nil


### **UsernameAvailable** - */usernameAvailable*
UsernameAvailable checks to see if the given username is available to use. When setting up Grüvee we give the user to change their username that people are able to see in the app. We use this function to check and see if the username they desire is available.

#### Request
```go
type usernameAvailableReq struct {
    Username string `json:"username"`
}
```

#### Response
`200`
```go
type usernameAvailableResp struct {
    Result bool `json:"result"`
}
```

`500` -> Any function errors