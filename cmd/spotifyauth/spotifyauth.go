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
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

const spotifyMeURI = "https://api.spotify.com/v1/me"

var (
	httpClient      *http.Client
	firestoreClient *firestore.Client
	logger          sawmill.Logger
	hostname        string
)

func init() {
	log.Println("AuthorizeWithSpotify initialized")
}

// AuthorizeWithSpotify will verify Spotify creds are valid and return any associated Firebase user or create a new Firebase user
func AuthorizeWithSpotify(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnvErr", initWithEnvErr, nil)
		return
	}

	var spotifyAuthRequest social.SpotifyAuthRequest

	authResponseErr := json.NewDecoder(request.Body).Decode(&spotifyAuthRequest)
	if authResponseErr != nil {
		http.Error(writer, authResponseErr.Error(), http.StatusInternalServerError)
		logger.LogErr("SpotifyAuthRequest Decoder", authResponseErr, request)
		return
	}

	log.Printf("Decoded SpotifyAuthRequest: %v", spotifyAuthRequest)

	if len(spotifyAuthRequest.APIToken) == 0 {
		http.Error(writer, "AuthorizeWithSpotify: ApiToken was empty.", http.StatusBadRequest)
		logger.LogErr("SpotifyAuthRequest Decoder", fmt.Errorf("ApiToken was empty"), request)
		return
	}

	spotifyMeReq, spotifyMeReqErr := http.NewRequest("GET", spotifyMeURI, nil)
	if spotifyMeReqErr != nil {
		http.Error(writer, spotifyMeReqErr.Error(), http.StatusInternalServerError)
		logger.LogErr("Request", spotifyMeReqErr, spotifyMeReq)
		return
	}

	log.Printf("Created SpotifyMeRequest %v", spotifyMeReq)

	// pheonix_d123 - "Client's gotta do what the Client's gotta do!" (02/26/20)
	spotifyMeReq.Header.Add("Authorization", "Bearer "+spotifyAuthRequest.APIToken)

	log.Printf("Added Headers tp SpotifyMeRequest %v", spotifyMeReq.Header)

	resp, httpErr := httpClient.Do(spotifyMeReq)
	if httpErr != nil {
		http.Error(writer, httpErr.Error(), http.StatusBadRequest)
		logger.LogErr("GET Request", httpErr, spotifyMeReq)
		return
	}

	// Check to see if request was valid
	if resp.StatusCode != http.StatusOK {
		log.Printf("SpotifyMeReq came back with code %v", resp.StatusCode)

		// Convert Spotify Error Object
		var spotifyErrorObj social.SpotifyRequestError

		err := json.NewDecoder(resp.Body).Decode(&spotifyErrorObj)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			logger.LogErr("Spotify Request Decoder", err, spotifyMeReq)
			return
		}

		http.Error(writer, spotifyErrorObj.Error.Message, spotifyErrorObj.Error.Status)
		logger.LogErr("Spotify Request Decoder", fmt.Errorf(spotifyErrorObj.Error.Message), spotifyMeReq)
		return
	}

	var spotifyMeResponse social.SpotifyMeResponse

	// syszen - "wait that it? #easyGo"(02/27/20)
	// LilCazza - "Why the fuck doesn't this shit work" (02/27/20)
	respDecodeErr := json.NewDecoder(resp.Body).Decode(&spotifyMeResponse)
	if respDecodeErr != nil {
		http.Error(writer, respDecodeErr.Error(), http.StatusBadRequest)
		logger.LogErr("Spotify Request Decoder", respDecodeErr, spotifyMeReq)
		return
	}

	log.Printf("SpotifyMeReq was a success. Decoded response: %v", spotifyMeResponse)

	// Check DB for user, if there return user object
	authorizeWithSpotifyResp, userErr := getUser(spotifyMeResponse.ID)
	if userErr != nil {
		http.Error(writer, userErr.Error(), http.StatusBadRequest)
		logger.LogErr("GetUser", userErr, nil)
		return
	}

	log.Printf("Response from getUser check %v", authorizeWithSpotifyResp)

	// We do not have our user
	if authorizeWithSpotifyResp == nil && userErr == nil {
		log.Println("No user found. Need to create one.")

		// First, generate & write social platform object
		socialPlatDocRef, socialPlatData, socialPlatErr := createSocialPlatform(spotifyMeResponse, spotifyAuthRequest)
		if socialPlatErr != nil {
			http.Error(writer, socialPlatErr.Error(), http.StatusBadRequest)
			logger.LogErr("CreateSocialPlatform", socialPlatErr, nil)
			return
		}

		log.Printf("Social platform generated: %v", socialPlatDocRef)

		// Then, generate & write Firestore User object
		var firestoreUser, firestoreUserErr = createUser(spotifyMeResponse, socialPlatDocRef)
		if firestoreUserErr != nil {
			http.Error(writer, firestoreUserErr.Error(), http.StatusBadRequest)
			log.Printf("AuthorizeWithSpotify [createUser]: %v", firestoreUserErr)
			return
		}

		log.Printf("Firestore user generated: %v", firestoreUser)

		// Finally, get custom JWT
		var customToken, customTokenErr = getCustomToken(firestoreUser.ID)
		if customTokenErr != nil {
			http.Error(writer, customTokenErr.Error(), http.StatusBadRequest)
			logger.LogErr("CustomToken", customTokenErr, nil)
			return
		}

		log.Printf("Custom JWT Generated: %v", customToken)

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
		log.Println("User found!")
		// Still need to get our custom token here
		var customToken, customTokenErr = getCustomToken(authorizeWithSpotifyResp.ID)
		if customTokenErr != nil {
			http.Error(writer, customTokenErr.Error(), http.StatusBadRequest)
			logger.LogErr("CustomToken", customTokenErr, nil)
			return
		}
		authorizeWithSpotifyResp.JWT = customToken.Token

		log.Printf("Received token: %v", customToken.Token)

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(authorizeWithSpotifyResp)
	}

	return
}

// no_neon_one - "BACKEND as a service" (02/29/20)
// sillyonly - "still waiting on alecc to give me a discount" (02/29/20)
