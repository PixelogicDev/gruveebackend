package spotifyauth

// InukApp - "Todo: add Plex auth support" (03/22/20)
// DaedTTV - "32 Font Size Kinda THICC" (03/23/20)
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
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var spotifyMeURI = "https://api.spotify.com/v1/me"
var httpClient *http.Client
var firestoreClient *firestore.Client
var logger sawmill.Logger
var hostname string

func init() {
	// Initialize client
	httpClient = &http.Client{}
	log.Println("AuthorizeWithSpotify initialized")
}

// AuthorizeWithSpotify will verify Spotify creds are valid and return any associated Firebase user or create a new Firebase user
func AuthorizeWithSpotify(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr(initWithEnvErr, "initWithEnvErr", nil)
		return
	}

	var spotifyAuthRequest social.SpotifyAuthRequest

	authResponseErr := json.NewDecoder(request.Body).Decode(&spotifyAuthRequest)
	if authResponseErr != nil {
		http.Error(writer, authResponseErr.Error(), http.StatusInternalServerError)
		logger.LogErr(authResponseErr, "spotifyAuthRequest Decoder", request)
		return
	}

	if len(spotifyAuthRequest.APIToken) == 0 {
		http.Error(writer, "AuthorizeWithSpotify: ApiToken was empty.", http.StatusBadRequest)
		logger.LogErr(fmt.Errorf("ApiToken was empty"), "spotifyAuthRequest Decoder", request)
		return
	}

	spotifyMeReq, spotifyMeReqErr := http.NewRequest("GET", spotifyMeURI, nil)
	if spotifyMeReqErr != nil {
		http.Error(writer, spotifyMeReqErr.Error(), http.StatusInternalServerError)
		logger.LogErr(spotifyMeReqErr, "http.NewRequest", spotifyMeReq)
		return
	}

	// pheonix_d123 - "Client's gotta do what the Client's gotta do!" (02/26/20)
	spotifyMeReq.Header.Add("Authorization", "Bearer "+spotifyAuthRequest.APIToken)
	resp, httpErr := httpClient.Do(spotifyMeReq)
	if httpErr != nil {
		http.Error(writer, httpErr.Error(), http.StatusBadRequest)
		logger.LogErr(httpErr, "client.Do", spotifyMeReq)
		return
	}

	// Check to see if request was valid
	if resp.StatusCode != http.StatusOK {
		// Convert Spotify Error Object
		var spotifyErrorObj social.SpotifyRequestError

		err := json.NewDecoder(resp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			logger.LogErr(err, "Spotify Request Decoder", spotifyMeReq)
			return
		}

		http.Error(writer, spotifyErrorObj.Error.Message, spotifyErrorObj.Error.Status)
		logger.LogErr(fmt.Errorf(spotifyErrorObj.Error.Message), "Spotify Request Decoder", spotifyMeReq)
		return
	}

	var spotifyMeResponse social.SpotifyMeResponse

	// syszen - "wait that it? #easyGo"(02/27/20)
	// LilCazza - "Why the fuck doesn't this shit work" (02/27/20)
	respDecodeErr := json.NewDecoder(resp.Body).Decode(&spotifyMeResponse)
	if respDecodeErr != nil {
		http.Error(writer, respDecodeErr.Error(), http.StatusBadRequest)
		logger.LogErr(respDecodeErr, "Spotify Request Decoder", spotifyMeReq)
		return
	}

	// Check DB for user, if there return user object
	authorizeWithSpotifyResp, userErr := getUser(spotifyMeResponse.ID)
	if userErr != nil {
		http.Error(writer, userErr.Error(), http.StatusBadRequest)
		logger.LogErr(userErr, "getUser", nil)
		return
	}

	// We do not have our user
	if authorizeWithSpotifyResp == nil && userErr == nil {
		// First, generate & write social platform object
		socialPlatDocRef, socialPlatData, socialPlatErr := createSocialPlatform(spotifyMeResponse, spotifyAuthRequest)
		if socialPlatErr != nil {
			http.Error(writer, socialPlatErr.Error(), http.StatusBadRequest)
			logger.LogErr(socialPlatErr, "createSocialPlatform", nil)
			return
		}

		// Then, generate & write Firestore User object
		var firestoreUser, firestoreUserErr = createUser(spotifyMeResponse, socialPlatDocRef)
		if firestoreUserErr != nil {
			http.Error(writer, firestoreUserErr.Error(), http.StatusBadRequest)
			log.Printf("AuthorizeWithSpotify [createUser]: %v", firestoreUserErr)
			return
		}

		// Finally, get custom JWT
		var customToken, customTokenErr = getCustomToken(firestoreUser.ID)
		if customTokenErr != nil {
			http.Error(writer, customTokenErr.Error(), http.StatusBadRequest)
			logger.LogErr(customTokenErr, "customToken", nil)
			return
		}

		// sillyonly: "path.addLine(to: CGPoint(x: rect.width, y: rect.height))" (03/13/20)
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		var spoitfyAuthResp = social.AuthorizeWithSpotifyResponse{
			Email:                   firestoreUser.Email,
			ID:                      firestoreUser.ID,
			Playlists:               []firebase.FirestorePlaylist{},
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
		// Still need to get our custom token here
		var customToken, customTokenErr = getCustomToken(authorizeWithSpotifyResp.ID)
		if customTokenErr != nil {
			http.Error(writer, customTokenErr.Error(), http.StatusBadRequest)
			logger.LogErr(customTokenErr, "customToken", nil)
			return
		}
		authorizeWithSpotifyResp.JWT = customToken.Token

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(authorizeWithSpotifyResp)
	}

	return
}

// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
		hostname = os.Getenv("HOSTNAME_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
		hostname = os.Getenv("HOSTNAME_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("AuthorizeWithSpotify [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), "NOT DEV", "AuthorizeWithSpotify")
	if err != nil {
		log.Printf("AuthorizeWithSpotify [Init Sawmill]: %v", err)
	}

	log.Println(currentProject)
	log.Println(hostname)

	logger = sawmillLogger
	firestoreClient = client
	return nil
}

// sillyonly - "So 140 char? is this twitter or a coding stream!" (03/02/20)
func getUser(uid string) (*social.AuthorizeWithSpotifyResponse, error) {
	// Go to firestore and check for uid
	fbID := "spotify:" + uid
	userRef := firestoreClient.Doc("users/" + fbID)
	if userRef == nil {
		return nil, fmt.Errorf("doesUserExist: users/%s is an odd path", fbID)
	}

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
	socialPlatformSnaps, socialPlatformSnapsErr := fetchSnapshots(firestoreUser.SocialPlatforms)
	if socialPlatformSnapsErr != nil {
		return nil, fmt.Errorf("FetchSnapshots: %v", socialPlatformSnapsErr)
	}

	// Conver socialPlatforms to data
	socialPlatforms, preferredPlatform := snapsToSocialPlatformData(socialPlatformSnaps)

	// Get references from playlists
	playlistsSnaps, playlistSnapsErr := fetchSnapshots(firestoreUser.Playlists)
	if playlistSnapsErr != nil {
		return nil, fmt.Errorf("FetchSnapshots: %v", playlistSnapsErr)
	}

	// Convert playlists to data
	playlists := snapsToPlaylistData(playlistsSnaps)

	// Convert user to response object
	authorizeWithSpotifyResponse := social.AuthorizeWithSpotifyResponse{
		Email:                   firestoreUser.Email,
		ID:                      firestoreUser.ID,
		Playlists:               playlists,
		PreferredSocialPlatform: preferredPlatform,
		SocialPlatforms:         socialPlatforms,
		Username:                firestoreUser.Username,
	}

	return &authorizeWithSpotifyResponse, nil
}

// fetchSnapshots takes in an array for Firestore Documents references and return their DocumentSnapshots
func fetchSnapshots(refs []*firestore.DocumentRef) ([]*firestore.DocumentSnapshot, error) {
	docsnaps, err := firestoreClient.GetAll(context.Background(), refs)
	if err != nil {
		return nil, fmt.Errorf("fetchChildRefs: %v", err)
	}

	return docsnaps, nil
}

// snapsToPlaylistData takes in array of Firestore DocumentSnapshots and retursn array of FirestorePlaylists
func snapsToPlaylistData(snaps []*firestore.DocumentSnapshot) []firebase.FirestorePlaylist {
	var playlists []firebase.FirestorePlaylist

	for _, playlistSnap := range snaps {
		var playlist firebase.FirestorePlaylist

		dataErr := playlistSnap.DataTo(&playlist)
		if dataErr != nil {
			log.Printf("Encountered error while parsing playlist snapshot.")
			log.Printf("%v", dataErr)
			continue
		}

		playlists = append(playlists, playlist)
	}

	return playlists
}

// snapsToSocialPlatformData takes in array of Firestore DocumentSnapshots and retursn array of FirestoreSocialPlatforms & PreferredPlatform
func snapsToSocialPlatformData(snaps []*firestore.DocumentSnapshot) ([]firebase.FirestoreSocialPlatform, firebase.FirestoreSocialPlatform) {
	var socialPlatforms []firebase.FirestoreSocialPlatform
	var preferredService firebase.FirestoreSocialPlatform

	for _, socialSnaps := range snaps {
		var socialPlatform firebase.FirestoreSocialPlatform

		dataErr := socialSnaps.DataTo(&socialPlatform)
		if dataErr != nil {
			log.Printf("Encountered error while parsing socialSnaps.")
			log.Printf("%v", dataErr)
			continue
		}

		socialPlatforms = append(socialPlatforms, socialPlatform)

		if socialPlatform.IsPreferredService {
			preferredService = socialPlatform
		}
	}

	return socialPlatforms, preferredService
}

// createUser takes in the spotify response and returns a new firebase user
func createUser(spotifyResp social.SpotifyMeResponse,
	socialPlatDocRef *firestore.DocumentRef) (*firebase.FirestoreUser, error) {
	var createUserURI = hostname + "/createUser"

	// Get profile image
	var profileImage firebase.SpotifyImage
	if len(spotifyResp.Images) > 0 {
		profileImage = spotifyResp.Images[0]
	} else {
		profileImage = firebase.SpotifyImage{}
	}

	log.Println(socialPlatDocRef)

	// Create, CreateUser Request object
	var createUserReq = social.CreateUserReq{
		Email:              spotifyResp.Email,
		ID:                 "spotify:" + spotifyResp.ID,
		SocialPlatformPath: "social_platforms/" + socialPlatDocRef.ID,
		ProfileImage:       &profileImage,
		Username:           spotifyResp.DisplayName,
	}

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(createUserReq)
	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	// Create Request
	createUser, createUserErr := http.NewRequest("POST", createUserURI, bytes.NewBuffer(jsonPlatform))
	if createUserErr != nil {
		return nil, fmt.Errorf(createUserErr.Error())
	}

	createUser.Header.Add("Content-Type", "application/json")
	createUserResp, httpErr := httpClient.Do(createUser)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	if createUserResp.StatusCode != http.StatusOK {
		// Get error from body
		var body []byte
		body, _ = ioutil.ReadAll(createUserResp.Body)
		return nil, fmt.Errorf((string(body)))
	}

	var firestoreUser firebase.FirestoreUser
	respDecodeErr := json.NewDecoder(createUserResp.Body).Decode(&firestoreUser)
	if respDecodeErr != nil {
		return nil, fmt.Errorf(respDecodeErr.Error())
	}

	return &firestoreUser, nil
}

// createSocialPlatform calls our CreateSocialPlatform Firebase Function to create & write new platform to DB
func createSocialPlatform(spotifyResp social.SpotifyMeResponse,
	authReq social.SpotifyAuthRequest) (*firestore.DocumentRef, *firebase.FirestoreSocialPlatform, error) {
	var createSocialPlatformURI = hostname + "/createSocialPlatform"

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

	// Adds the expiresIn time to current time
	var expiredAtStr = time.Now().Add(time.Second * time.Duration(authReq.ExpiresIn))

	var apiToken = firebase.APIToken{
		CreatedAt: time.Now().Format(time.RFC3339),
		ExpiredAt: expiredAtStr.Format(time.RFC3339),
		ExpiresIn: authReq.ExpiresIn,
		Token:     authReq.APIToken,
	}

	// Object that we will write to Firestore
	var platform = firebase.FirestoreSocialPlatform{
		APIToken:           apiToken,
		RefreshToken:       authReq.RefreshToken,
		Email:              spotifyResp.Email,
		ID:                 spotifyResp.ID,
		IsPreferredService: true, // If creating a new user, this is the first platform which should be the default
		IsPremium:          isPremium,
		PlatformName:       "spotify",
		ProfileImage:       profileImage,
		Username:           spotifyResp.DisplayName,
	}

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(platform)
	if jsonErr != nil {
		return nil, nil, fmt.Errorf(jsonErr.Error())
	}

	// Create Request
	socialPlatformReq, newReqErr := http.NewRequest("POST", createSocialPlatformURI, bytes.NewBuffer(jsonPlatform))
	if newReqErr != nil {
		return nil, nil, fmt.Errorf(newReqErr.Error())
	}

	// Run firebase function to write platform to database
	socialPlatformReq.Header.Add("Content-Type", "application/json")
	socialPlatformResp, httpErr := httpClient.Do(socialPlatformReq)
	if httpErr != nil {
		return nil, nil, fmt.Errorf(httpErr.Error())
	}

	if socialPlatformResp.StatusCode != http.StatusOK {
		// Get error from body
		var body, _ = ioutil.ReadAll(socialPlatformResp.Body)
		return nil, nil, fmt.Errorf(string(body))
	}

	// Get Document reference
	platformRef := firestoreClient.Doc("social_platforms/" + platform.ID)
	if platformRef == nil {
		return nil, nil, fmt.Errorf("Odd number of IDs or the ID was empty")
	}

	return platformRef, &platform, nil
}

// getCustomRoken calles our GenerateToken Firebase Function to create & return custom JWT
func getCustomToken(uid string) (*social.GenerateTokenResponse, error) {
	var generateTokenURI = hostname + "/generateCustomToken"
	var tokenRequest = social.GenerateTokenRequest{
		UID: uid,
	}

	// Create jsonBody
	jsonPlatform, jsonErr := json.Marshal(tokenRequest)
	if jsonErr != nil {
		return nil, fmt.Errorf(jsonErr.Error())
	}

	// Create Request
	customTokenReq, customTokenReqErr := http.NewRequest("POST", generateTokenURI, bytes.NewBuffer(jsonPlatform))
	if customTokenReqErr != nil {
		return nil, fmt.Errorf(customTokenReqErr.Error())
	}

	customTokenReq.Header.Add("Content-Type", "application/json")
	customTokenResp, httpErr := httpClient.Do(customTokenReq)
	if httpErr != nil {
		return nil, fmt.Errorf(httpErr.Error())
	}

	// Decode the token to send back
	var tokenResponse social.GenerateTokenResponse
	customTokenDecodeErr := json.NewDecoder(customTokenResp.Body).Decode(&tokenResponse)
	if customTokenDecodeErr != nil {
		return nil, fmt.Errorf(customTokenDecodeErr.Error())
	}

	return &tokenResponse, nil
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
