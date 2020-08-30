package createappledevtoken

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
)

// DR_DinoMight - "Alec loves the song "Nelly's - Hot In Here" (05/05/20)
var (
	currentProject      string
	firestoreClient     *firestore.Client
	logger              sawmill.Logger
	appleDevToken       firebase.FirestoreAppleDevJWT
	applePrivateKeyPath string
)

func init() {
	log.Println("CreateAppleDevToken initialized.")
}

// CreateAppleDevToken will render a HTML page to get the AppleMusic credentials for user
func CreateAppleDevToken(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initWithEnvErr, nil)
		return
	}

	// Check for developer token in firebase
	devToken, devTokenErr := fetchToken()
	if devTokenErr != nil {
		http.Error(writer, devTokenErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GetAppleDevToken", devTokenErr, nil)
		return
	}

	logger.Log("GetAppleDevToken", "DevToken received from Firebase")

	// If we have the token check to see if it needs to be refreshed
	if devToken != nil {
		appleDevToken = *devToken

		logger.Log("CreateAppleDevToken", "Token found in DB. Checking for expiration")

		// If token expired, refresh, else continue
		if isTokenExpired(devToken) {
			if os.Getenv("APPLE_TEAM_ID") == "" {
				http.Error(writer, "[CreateAppleDevToken] APPLE_TEAM_ID does not exist!", http.StatusInternalServerError)
				logger.LogErr("RefreshAppleDevToken", fmt.Errorf("APPLE_TEAM_ID does not exist"), nil)
				return
			}

			if os.Getenv("APPLE_KID") == "" {
				http.Error(writer, "[CreateAppleDevToken] APPLE_KID does not exist!", http.StatusInternalServerError)
				logger.LogErr("RefreshAppleDevToken", fmt.Errorf("APPLE_KID does not exist"), nil)
				return
			}

			logger.Log("CreateAppleDevToken", "Apple Team Id & Apple KID are here.")

			token, tokenErr := generateJWT()
			if tokenErr != nil {
				http.Error(writer, tokenErr.Error(), http.StatusInternalServerError)
				logger.LogErr("RefreshAppleDevToken", tokenErr, nil)
				return
			}

			logger.Log("GenerateJWT", "JWT generated successfully")

			appleDevToken = *token
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(appleDevToken)
		return
	}

	// If we do not have the token, we need to generate a new one
	token, tokenErr := generateJWT()
	if tokenErr != nil {
		http.Error(writer, tokenErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GenerateAppleDevToken", tokenErr, nil)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(token)
}
