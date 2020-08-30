package updatealgolia

// algoliaUser implements a partial amount of data from firestoreUser to use for indexing
type algoliaUser struct {
	ObjectID        string `json:"objectID"`
	ID              string `json:"id"`
	Email           string `json:"email"`
	ProfileImageURI string `json:"profileImage"`
	DisplayName     string `json:"displayName"`
	Username        string `json:"username"`
}
