// Remaiten - “Maybe now is the time to switch to Vue Native Kappa” (02/22/20)
// BackeyM - "It's go time 🙃" (02/22/20)
// TheDkbay - "It's double go time" (02/22/20)
// LilCazza - "I can't see what's going on here, it may be because I can only hear" (02/24/20)
// Tensei_c - "rust > go :P" (02/24/20)
// pahnev - "swift > rust" (02/24/20)
// jackconceprio - "I like pineapple juice any line" (02/25/20)

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	auth "firebase.google.com/go/auth"
)

// GenerateTokenRequest includes the UID for the user that we want to create a custom token for
type GenerateTokenRequest struct {
	UID string
}

// GenerateTokenResponse includes what we will send back to the client
type GenerateTokenResponse struct {
	Token string `json:"token"`
}

var client *auth.Client

func init() {
	// Init Firebase App
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
		return
	}

	fmt.Println("Firebase app initialized.")

	// Init Fireabase Auth Admin
	client, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("auth.Client: %v", err)
		return
	}

	fmt.Println("Auth client initialized.")
}

// GenerateCustomToken generates a CustomToken for Firebase Login
func GenerateCustomToken(writer http.ResponseWriter, request *http.Request) {
	// Thinking about sending 3rd Party token here
	// then call 3rd party endpoint to get user data
	// Then we also know that this person has just authed with 3rd Party platform
	// Then here we call our write to DB
	// Then we can return JWT to that client
	var tokenRequest GenerateTokenRequest

	// Decode json from request
	err := json.NewDecoder(request.Body).Decode(&tokenRequest)
	if err != nil {
		log.Fatalf("json.NewDecoder: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// Is this person authorized to mint a token?
	// IE - Have they authenticated with a third party service such as spotify

	// Garahorn - "We need to generate the quantum GUID once the flux capacitor reaches terminal velocity." (02/24/20)
	// Get Custom Firebase Token
	token, err := client.CustomToken(context.Background(), tokenRequest.UID)
	if err != nil {
		log.Fatalf("client.CustomToken: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// Create reponse object and pass it along
	tokenResponse := GenerateTokenResponse{Token: token}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(tokenResponse)
}
