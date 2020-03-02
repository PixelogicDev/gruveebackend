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

// GenerateTokenRequest represents the UID for the user that we want to create a custom token for
type GenerateTokenRequest struct {
	UID string
}

// GenerateTokenResponse represents what we will send back to the client
type GenerateTokenResponse struct {
	Token string `json:"token"`
}

var client *auth.Client

func init() {
	// Init Firebase App
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Printf("firebase.NewApp: %v", err)
		return
	}

	fmt.Println("Firebase app initialized.")

	// Init Fireabase Auth Admin
	client, err = app.Auth(context.Background())
	if err != nil {
		log.Printf("auth.Client: %v", err)
		return
	}

	fmt.Println("Auth client initialized.")
}

// GenerateCustomToken generates a CustomToken for Firebase Login
func GenerateCustomToken(writer http.ResponseWriter, request *http.Request) {
	// If this is getting called, we have already authorized the user by verifying their API token is valid and pulls back their data
	var tokenRequest GenerateTokenRequest

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
	tokenResponse := GenerateTokenResponse{Token: token}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(tokenResponse)
}
