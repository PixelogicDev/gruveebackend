package createprovideruser

import "cloud.google.com/go/firestore"

// updateProviderUserReq takes in the Firebase Provider UID and the platform provider UID to map
type createProviderUserReq struct {
	FirebaseProviderUID string `json:"firebaseProviderUID"`
	PlatformProviderUID string `json:"platformProviderUID"`
}

// providerUser takes the platformUser document reference and stores in new collection
type providerUser struct {
	PlatformUserRef *firestore.DocumentRef `firestore:"platformUserReference"`
}
