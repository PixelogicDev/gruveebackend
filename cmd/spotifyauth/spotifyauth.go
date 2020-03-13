package spotifyauth

// thoastyk 	"X O X" (02/26/20)
// thoastyk 	"_ X O" (02/26/20)
// pheonix_d123	"O O X I wanna interrupt the tic-tac-toe." (03/08/20)
// Belonix97  	"X O O I want to interrupt the interrupted tic-tac-toe line." (03/08/20)
// ItsAstrix  	"O O X I wanna interrupt the tic-tac-toe." (03/08/20)
// thoastyk 	"X _ O" (02/26/20)
// creativenobu - "Have you flutter tried?" (02/26/20)
// TheDkbay - "If this were made in Flutter Alec would already be done but he loves to pain himself and us by using inferior technology maybe he will learn in the future." (03/02/20)
// OnePocketPimp - "Alec had an Idea at this moment in time 9:53 am 3-1-2020" (03/01/20)
// ZenonLoL - "go mod vendor - it just works" (03/08/20)
// gamma7869 - "Maybe if I get Corona, I could finally get friends. Corona Friends?" (03/12/20)
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	firestore "cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
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
	log.Printf("Firestore initialized")
}

// AuthorizeWithSpotify will verify spotify creds are valid and return that user or create a new user if the creds valid
func AuthorizeWithSpotify(writer http.ResponseWriter, request *http.Request) {
	var spotifyAuthRequest firebase.SpotifyAuthRequest

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
			var spotifyErrorObj firebase.SpotifyRequestError

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

		var spotifyMeResponse firebase.SpotifyMeResponse
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

		// We do not have our user
		if authorizeWithSpotifyResp == nil && userErr == nil {
			// First, generate & write social platform object
			log.Printf("Calling createSocialPlatform")
			socialPlatDocRef, socialPlatData := createSocialPlatform(writer, spotifyMeResponse, spotifyAuthRequest)
			log.Printf("%v", socialPlatDocRef)

			// Then, generate & write Firestore User object
			log.Printf("About to call createUser")
			var firestoreUser = createUser(writer, spotifyMeResponse, socialPlatDocRef)
			if firestoreUser == nil {
				http.Error(writer, "firestoreUser could not be created", http.StatusBadRequest)
				log.Printf("AuthorizeWithSpotify: %v", userErr)
				return
			}

			// Finally, get custom JWT
			log.Println("Getting custom JWT")
			var customToken = getCustomToken(writer, firestoreUser.ID)
			if customToken == nil {
				http.Error(writer, "Could not get Custom JWT", http.StatusBadRequest)
				log.Printf("AuthorizeWithSpotify: %v", userErr)
				return
			}

			// Return jwt and user object
			log.Println(customToken)

			// sillyonly: "path.addLine(to: CGPoint(x: rect.width, y: rect.height))" (03/13/20)
			writer.WriteHeader(http.StatusOK)
			writer.Header().Set("Content-Type", "application/json")
			var spoitfyAuthResp = firebase.AuthorizeWithSpotifyResponse{
				Email:                   firestoreUser.Email,
				ID:                      firestoreUser.ID,
				Playlists:               firestoreUser.Playlists,
				PreferredSocialPlatform: *socialPlatData,
				SocialPlatforms:         []firebase.FirestoreSocialPlatform{*socialPlatData},
				Username:                firestoreUser.Username,
				JWT:                     customToken.Token,
			}

			json.NewEncoder(writer).Encode(spoitfyAuthResp)
			return
		}

		// We have our user
		if authorizeWithSpotifyResp != nil {
			writer.WriteHeader(http.StatusOK)
			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(authorizeWithSpotifyResp)
		}

		return
	}

	http.Error(writer, "AuthorizeWithSpotify: ApiToken was empty.", http.StatusBadRequest)
	log.Printf("AuthorizeWithSpotify: ApiToken was empty.")
}

// sillyonly - "So 140 char? is this twitter or a coding stream!" (03/02/20)
func getUser(uid string) (*firebase.AuthorizeWithSpotifyResponse, error) {
	log.Printf("WE ARE HERE.")
	// Go to firestore and check for uid
	fbID := "spotify:" + uid

	log.Printf(fbID)

	userRef := firestoreClient.Doc("users/" + fbID)
	if userRef == nil {
		return nil, fmt.Errorf("doesUserExist: users/%s is an odd path", fbID)
	}

	log.Printf("WE GOT SOME USER.")

	// If uid does not exist return nil
	userSnap, err := userRef.Get(context.Background())
	if status.Code(err) == codes.NotFound {
		log.Printf("User with id %s was not found", fbID)
		return nil, nil
	}

	// UID does exist, return firestore user
	var firestoreUser firebase.FirestoreUser
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
	authorizeWithSpotifyResponse := firebase.AuthorizeWithSpotifyResponse{
		Email:                   firestoreUser.Email,
		ID:                      firestoreUser.ID,
		Playlists:               []string{},
		PreferredSocialPlatform: *preferredPlatform,
		SocialPlatforms:         *socialPlatforms,
		Username:                firestoreUser.Username,
	}

	return &authorizeWithSpotifyResponse, nil
}

func fetchChildRefs(refs []*firestore.DocumentRef) (*[]firebase.FirestoreSocialPlatform, *firebase.FirestoreSocialPlatform, error) {
	docsnaps, err := firestoreClient.GetAll(context.Background(), refs)
	if err != nil {
		return nil, nil, fmt.Errorf("fetchChildRefs: %v", err)
	}

	var platforms []firebase.FirestoreSocialPlatform
	var preferredService firebase.FirestoreSocialPlatform

	for _, userSnap := range docsnaps {
		var socialPlatform firebase.FirestoreSocialPlatform

		dataErr := userSnap.DataTo(&socialPlatform)
		if dataErr != nil {
			log.Printf("DATA ERROR")
			log.Printf("%v", dataErr)
		}

		if socialPlatform.IsPreferredService {
			preferredService = socialPlatform
		}

		platforms = append(platforms, socialPlatform)
	}

	return &platforms, &preferredService, nil
}

func createUser(writer http.ResponseWriter,
	spotifyResp firebase.SpotifyMeResponse,
	socialPlatDocRef *firestore.DocumentRef) *firebase.FirestoreUser {

	// Setup payload
	if socialPlatDocRef == nil {
		// Could not create social platform here
		log.Println("Could not write SocialPlatform Doc to DB")
		return nil
	}

	// Create firetoreUser Object
	var firestoreUser = firebase.FirestoreUser{
		Email:                   spotifyResp.Email,
		ID:                      "spotify:" + spotifyResp.ID,
		Playlists:               []string{},
		PreferredSocialPlatform: socialPlatDocRef,
		SocialPlatforms:         []*firestore.DocumentRef{socialPlatDocRef},
		Username:                spotifyResp.DisplayName,
	}

	log.Printf("Object built")

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(firestoreUser)
	if jsonErr != nil {
		http.Error(writer, jsonErr.Error(), http.StatusInternalServerError)
		log.Printf("createSocialPlatform [json.Marshal]: %v", jsonErr)
		return nil
	}

	log.Println("Calling createUser function.")

	// Create Request
	createUserReq, createUserReqErr := http.NewRequest("POST", "http://localhost:8080/createUser", bytes.NewBuffer(jsonPlatform))
	if createUserReqErr != nil {
		http.Error(writer, createUserReqErr.Error(), http.StatusInternalServerError)
		log.Printf("createUser [http.NewRequest]: %v", createUserReqErr)
		return nil
	}
	createUserReq.Header.Add("Content-Type", "application/json")

	createUserResp, httpErr := httpClient.Do(createUserReq)
	if httpErr != nil {
		http.Error(writer, httpErr.Error(), http.StatusBadRequest)
		log.Printf("createUser [client.Do]: %v", httpErr)
		return nil
	}

	if createUserResp.StatusCode != http.StatusOK {
		var body []byte
		// Get error from body
		body, _ = ioutil.ReadAll(createUserResp.Body)
		log.Printf(string(body))
		// TODO: Probably should return error
		return nil
	}

	return &firestoreUser
}

func createSocialPlatform(writer http.ResponseWriter,
	spotifyResp firebase.SpotifyMeResponse,
	authReq firebase.SpotifyAuthRequest) (*firestore.DocumentRef, *firebase.FirestoreSocialPlatform) {

	// Create request body
	var isPremium = false
	if spotifyResp.Product == "premium" {
		isPremium = true
	}

	var profileImage firebase.SpotifyImage
	if len(spotifyResp.Images) > 0 {
		profileImage = spotifyResp.Images[0]
	} else {
		profileImage = firebase.SpotifyImage{}
	}

	var platform = firebase.FirestoreSocialPlatform{
		APIToken:           authReq.APIToken,
		RefreshToken:       authReq.RefreshToken,
		Email:              spotifyResp.Email,
		ID:                 spotifyResp.ID,
		IsPreferredService: true, // If creating a new user, this is the first platform which should be the default
		IsPremium:          isPremium,
		PlatformName:       "spotify",
		ProfileImage:       profileImage,
		Username:           spotifyResp.DisplayName,
	}

	log.Printf("Object built")

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(platform)
	if jsonErr != nil {
		http.Error(writer, jsonErr.Error(), http.StatusInternalServerError)
		log.Printf("jsonErr was hit")
		log.Printf("createSocialPlatform [json.Marshal]: %v", jsonErr)
		return nil, nil
	}

	// Create Request
	socialPlatformReq, newReqErr := http.NewRequest("POST", "http://localhost:8080/createSocialPlatform", bytes.NewBuffer(jsonPlatform))
	if newReqErr != nil {
		http.Error(writer, newReqErr.Error(), http.StatusInternalServerError)
		log.Printf("newReqErr was hit")
		log.Printf("createSocialPlatform [http.NewRequest]: %v", newReqErr)
		return nil, nil
	}
	socialPlatformReq.Header.Add("Content-Type", "application/json")
	log.Printf("Req was built")

	// Run firebase function to write platform to database
	socialPlatformResp, httpErr := httpClient.Do(socialPlatformReq)
	if httpErr != nil {
		http.Error(writer, httpErr.Error(), http.StatusBadRequest)
		log.Printf("createUser [client.Do]: %v", httpErr)
		return nil, nil
	}

	if socialPlatformResp.StatusCode != http.StatusOK {
		// Get error from body
		var body, _ = ioutil.ReadAll(socialPlatformResp.Body)
		log.Printf(string(body))
		// TODO: Probably should return error
		return nil, nil
	}

	log.Printf("Document was written to DB successfully")
	// Get Document reference
	platformRef := firestoreClient.Doc("social_platforms/" + platform.ID)
	if platformRef == nil {
		return nil, nil
	}

	log.Printf("Made request to create SocialPlatform")
	return platformRef, &platform
}

func getCustomToken(writer http.ResponseWriter, uid string) *firebase.GenerateTokenResponse {
	var tokenRequest = firebase.GenerateTokenRequest{
		UID: uid,
	}

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(tokenRequest)
	if jsonErr != nil {
		http.Error(writer, jsonErr.Error(), http.StatusInternalServerError)
		log.Printf("getCustomToken [json.Marshal]: %v", jsonErr)
		return nil
	}

	// Create Request
	customTokenReq, customTokenReqErr := http.NewRequest("POST", "http://localhost:8080/generateToken", bytes.NewBuffer(jsonPlatform))
	if customTokenReqErr != nil {
		http.Error(writer, customTokenReqErr.Error(), http.StatusInternalServerError)
		log.Printf("getCustomToken [http.NewRequest]: %v", customTokenReqErr)
		return nil
	}
	customTokenReq.Header.Add("Content-Type", "application/json")

	customTokenResp, httpErr := httpClient.Do(customTokenReq)
	if httpErr != nil {
		http.Error(writer, httpErr.Error(), http.StatusBadRequest)
		log.Printf("getCustomToken [client.Do]: %v", httpErr)
		return nil
	}

	// Decode the token to send back
	var tokenResponse firebase.GenerateTokenResponse

	customTokenDecodeErr := json.NewDecoder(customTokenResp.Body).Decode(&tokenResponse)
	if customTokenDecodeErr != nil {
		http.Error(writer, customTokenDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("getCustomToken [tokenResponse Decoder]: %v", customTokenDecodeErr)
		return nil
	}

	return &tokenResponse
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
