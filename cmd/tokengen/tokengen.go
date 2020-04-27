package tokengen

// Remaiten - “Maybe now is the time to switch to Vue Native Kappa” (02/22/20)
// BackeyM - "It's go time 🙃" (02/22/20)
// TheDkbay - "It's double go time" (02/22/20)
// LilCazza - "I can't see what's going on here, it may be because I can only hear" (02/24/20)
// Tensei_c - "rust > go :P" (02/24/20)
// pahnev - "swift > rust" (02/24/20)
// jackconceprio - "I like pineapple juice any line" (02/25/20)
// OnePocketPimp - "TODO: Create Dog Treat API to properly reward all good doggos on streaming platforms" (03/08/20)
import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/pixelogicdev/gruveebackend/pkg/social"
)

var client *auth.Client

func init() {
	// Init Firebase App
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Printf("firebase.NewApp: %v", err)
		return
	}
	log.Println("Firebase NewApp initialized")

	// Init Fireabase Auth Admin
	client, err = app.Auth(context.Background())
	if err != nil {
		log.Printf("auth.Client: %v", err)
		return
	}
	log.Println("Firebase AuthAdmin initialized")
}

// GenerateCustomToken generates a CustomToken for Firebase Login
func GenerateCustomToken(writer http.ResponseWriter, request *http.Request) {
	// If this is getting called, we have already authorized the user by verifying their API token is valid and pulls back their data
	var tokenRequest social.GenerateTokenRequest

	// Decode json from request
	err := json.NewDecoder(request.Body).Decode(&tokenRequest)
	if err != nil {
		log.Printf("json.NewDecoder: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// Garahorn - "We need to generate the quantum GUID once the flux capacitor reaches terminal velocity." (02/24/20)
	token, err := client.CustomToken(context.Background(), tokenRequest.UID)
	if err != nil {
		log.Printf("client.CustomToken: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// Create reponse object and pass it along
	tokenResponse := social.GenerateTokenResponse{Token: token}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(tokenResponse)
}
