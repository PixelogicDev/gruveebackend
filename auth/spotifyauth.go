// thoastyk - "X O X" (02/26/20)
// thoastyk - "_ X O" (02/26/20)
// thoastyk - "X _ O" (02/26/20)
// creativenobu - "Have you flutter tried?" (02/26/20)
// TheDkbay - "If this were made in Flutter Alec would already be done but he loves to pain himself and us by using inferior technology maybe he will learn in the future." (03/02/20)
// OnePocketPimp - "Alec had an Idea at this moment in time 9:53 am 3-1-2020" (03/01/20)
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	firestore "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var httpClient *http.Client
var firestoreClient *firestore.Client

func init() {
	// Initialize client
	httpClient = &http.Client{}

	// Initialize Firestore
	// TODO: Get way to use env variable for project id
	client, err := firestore.NewClient(context.Background(), "gruvee-3b7c4")
	if err != nil {
		log.Printf("AuthorizeWithSpotify [Init Firestore]: %v", err)
		return
	}

	firestoreClient = client
	log.Println("Firestore initialized")
}

// AuthorizeWithSpotify will verify spotify creds are valid and return that user or create a new user if the creds valid
func AuthorizeWithSpotify(writer http.ResponseWriter, request *http.Request) {
	var spotifyAuthRequest SpotifyAuthRequest

	authResponseErr := json.NewDecoder(request.Body).Decode(&spotifyAuthRequest)
	if authResponseErr != nil {
		http.Error(writer, authResponseErr.Error(), http.StatusInternalServerError)
		log.Printf("AuthorizeWithSpotify [spotifyAuthRequest Decoder]: %v", authResponseErr)
		return
	}

	if len(spotifyAuthRequest.APIToken) > 0 {
		req, newReqErr := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
		if newReqErr != nil {
			http.Error(writer, newReqErr.Error(), http.StatusInternalServerError)
			log.Printf("AuthorizeWithSpotify [http.NewRequest]: %v", newReqErr)
			return
		}

		req.Header.Add("Authorization", "Bearer "+spotifyAuthRequest.APIToken)

		// pheonix_d123 - "Client's gotta do what the Client's gotta do!" (02/26/20)
		resp, httpErr := httpClient.Do(req)
		if httpErr != nil {
			http.Error(writer, httpErr.Error(), http.StatusBadRequest)
			log.Printf("AuthorizeWithSpotify [client.Do]: %v", httpErr)
			return
		}

		// Check to see if request was valid
		if resp.StatusCode != http.StatusOK {
			// Conver Spotify Error Object
			var spotifyErrorObj SpotifyRequestError

			err := json.NewDecoder(resp.Body).Decode(&spotifyErrorObj)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				log.Printf("AuthorizeWithSpotify [Spotify Request Decoder]: %v", err)
				return
			}

			http.Error(writer, spotifyErrorObj.Error.Message, spotifyErrorObj.Error.Status)
			log.Printf("AuthorizeWithSpotify [Spotify Request Decoder]: %v", spotifyErrorObj.Error.Message)
			return
		}

		var spotifyMeResponse SpotifyMeResponse
		// syszen - "wait that it? #easyGo"(02/27/20)
		// LilCazza - "Why the fuck doesn't this shit work" (02/27/20)
		responseErr := json.NewDecoder(resp.Body).Decode(&spotifyMeResponse)
		if responseErr != nil {
			http.Error(writer, responseErr.Error(), http.StatusBadRequest)
			log.Printf("AuthorizeWithSpotify [Spotify Request Decoder]: %v", responseErr)
			return
		}

		// Check DB for user, if there return user object else return nil
		authorizeWithSpotifyResp, userErr := getUser(spotifyMeResponse.ID)
		if userErr != nil {
			http.Error(writer, userErr.Error(), http.StatusBadRequest)
			log.Printf("AuthorizeWithSpotify: %v", userErr)
			return
		}

		// We have our user
		if authorizeWithSpotifyResp != nil {
			writer.WriteHeader(http.StatusOK)
			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(authorizeWithSpotifyResp)
		}

		// Create new user if nil
		return
	}

	http.Error(writer, "AuthorizeWithSpotify: ApiToken was empty.", http.StatusBadRequest)
	log.Println("AuthorizeWithSpotify: ApiToken was empty.")
}

func getUser(uid string) (*AuthorizeWithSpotifyResponse, error) {
	// Go to firestore and check for uid
	fbID := "spotify:" + uid

	user := firestoreClient.Doc("users/" + fbID)
	if user == nil {
		return nil, fmt.Errorf("doesUserExist: users/%s is an odd path", fbID)
	}

	// If uid does not exist return nil
	userSnap, err := user.Get(context.Background())
	if status.Code(err) == codes.NotFound {
		return nil, fmt.Errorf("doesUserExist: %v uid does not exist", fbID)
	}

	// UID does exist, return firestore user
	var firestoreUser FirestoreUser
	dataErr := userSnap.DataTo(&firestoreUser)
	if dataErr != nil {
		return nil, fmt.Errorf("doesUserExist: %v", dataErr)
	}

	// Get references from socialPlatforms
	socialPlatforms, preferredPlatform, socialErr := fetchChildRefs(firestoreUser.SocialPlatforms)
	if socialErr != nil {
		return nil, fmt.Errorf("doesUserExist: %v", socialErr)
	}

	// Convert user to response object
	authorizeWithSpotifyResponse := AuthorizeWithSpotifyResponse{
		Email:                   firestoreUser.Email,
		ID:                      firestoreUser.ID,
		Playlists:               []string{},
		PreferredSocialPlatform: *preferredPlatform,
		SocialPlatforms:         *socialPlatforms,
		Username:                firestoreUser.Username,
	}

	return &authorizeWithSpotifyResponse, nil
}

func fetchChildRefs(refs []*firestore.DocumentRef) (*[]FirestoreSocialPlatform, *FirestoreSocialPlatform, error) {
	docsnaps, err := firestoreClient.GetAll(context.Background(), refs)
	if err != nil {
		return nil, nil, fmt.Errorf("fetchChildRefs: %v", err)
	}

	var platforms []FirestoreSocialPlatform
	var preferredService FirestoreSocialPlatform

	for _, userSnap := range docsnaps {
		var socialPlatform FirestoreSocialPlatform

		dataErr := userSnap.DataTo(&socialPlatform)
		if dataErr != nil {
			fmt.Println(dataErr)
		}

		if socialPlatform.IsPreferredService {
			preferredService = socialPlatform
		}

		platforms = append(platforms, socialPlatform)
	}

	return &platforms, &preferredService, nil
}

// no_neon_one - "BACKEND as a service" (02/29/20)
// sillyonly - "still waiting on alecc to give me a discount" (02/29/20)
// jackconceprio - "Baby lock the door and turn the lights down low Put some music on that's soft and slow
// Baby we ain't got no place to go I hope you understand I've been thinking 'bout this all day long Never
// felt a feeling quite this strong I can't believe how much it turns me on Just to be your man There's
// no hurry Don't you worry We can take our time Come a little closer Lets go over What I had in mind
// Baby lock the door and turn the lights down low Put some music on that's soft and slow Baby we ain't
// got no place to go I hope you understand I've been thinking 'bout this all day long Never felt a
// feeling quite this strong I can't believe how much it turns me on Just to be your man Ain't nobody ever
// love nobody The way that I love you We're alone now You don't know how Long I've wanted to Lock the door
// and turn the lights down low Put some music on that's soft and slow Baby we ain't got no place to go I
// hope you understand I've been thinking 'bout this all day long Never felt a feeling that was quite this
// strong I can't believe how much it turns me on Just to be your man I can't believe how much it turns me
// on Just to be your own" (03/01/20)
