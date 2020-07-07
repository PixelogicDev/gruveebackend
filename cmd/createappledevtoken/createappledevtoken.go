package createappledevtoken

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/dgrijalva/jwt-go"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DR_DinoMight - "Alec loves the song "Nelly's - Hot In Here" (05/05/20)
var firestoreClient *firestore.Client
var appleDevToken firebase.FirestoreAppleDevJWT
var p8FilePath string
var currentProject string

// CreateAppleDevToken will render a HTML page to get the AppleMusic credentials for user
func CreateAppleDevToken(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		log.Printf("[AuthorizeWithApple] [initWithEnv]: %v", initWithEnvErr)
		return
	}

	// Check for developer token in firebase
	devToken, devTokenErr := fetchToken()
	if devTokenErr != nil {
		http.Error(writer, devTokenErr.Error(), http.StatusInternalServerError)
		log.Printf("[CreateAppleDevToken] [getAppleDevToken]: %v", devTokenErr)
		return
	}

	if devToken != nil {
		appleDevToken = *devToken
		log.Println("[CreateAppleDevToken] Token found in DB. Checking for expiration")

		// If token expired, refresh, else continue
		if isTokenExpired(devToken) {
			if os.Getenv("APPLE_TEAM_ID") == "" {
				http.Error(writer, "[CreateAppleDevToken] APPLE_TEAM_ID does not exist!", http.StatusInternalServerError)
				log.Println("[CreateAppleDevToken] APPLE_TEAM_ID does not exist!")
				return
			}

			if os.Getenv("APPLE_KID") == "" {
				http.Error(writer, "[CreateAppleDevToken] APPLE_KID does not exist!", http.StatusInternalServerError)
				log.Println("[CreateAppleDevToken] APPLE_KID does not exist!")
				return
			}

			token, tokenErr := generateJWT()
			if tokenErr != nil {
				http.Error(writer, tokenErr.Error(), http.StatusInternalServerError)
				log.Printf("[CreateAppleDevToken]: %v", tokenErr)
				return
			}

			appleDevToken = *token
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(appleDevToken)
		return
	}

	token, tokenErr := generateJWT()
	if tokenErr != nil {
		http.Error(writer, tokenErr.Error(), http.StatusInternalServerError)
		log.Printf("CreateAppleDevToken]: %v", tokenErr)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(token)
}

// Helpers
// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
		p8FilePath = os.Getenv("APPLE_P8_PATH_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
		p8FilePath = os.Getenv("APPLE_P8_PATH_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("CreateAppleDevToken [Init Firestore]: %v", err)
	}

	firestoreClient = client
	return nil
}

// fetchToken will grab the Apple Developer Token from DB
func fetchToken() (*firebase.FirestoreAppleDevJWT, error) {
	// Go to Firebase and see if appleDevToken exists
	snapshot, snapshotErr := firestoreClient.Collection("internal_tokens").Doc("appleDevToken").Get(context.Background())
	if status.Code(snapshotErr) == codes.NotFound {
		log.Println("[CreateAppleDevToken] AppleDevToken not found in DB. Need to create.")
		return nil, nil
	}

	if snapshotErr != nil {
		return nil, fmt.Errorf(snapshotErr.Error())
	}

	var appleDevToken firebase.FirestoreAppleDevJWT
	dataToErr := snapshot.DataTo(&appleDevToken)
	if dataToErr != nil {
		return nil, fmt.Errorf(dataToErr.Error())
	}

	return &appleDevToken, nil
}

// isTokenExpired will check to see if the Apple Developer Token is expired
func isTokenExpired(token *firebase.FirestoreAppleDevJWT) bool {
	// Get current time
	var currentTime = time.Now()

	fmt.Printf("Issued At: %d seconds\n", token.IssuedAt)
	fmt.Printf("Expires At: %d seconds\n", token.ExpiresAt)

	if currentTime.After(time.Unix(token.ExpiresAt, 0)) {
		log.Println("Dev Token is expired. Generating a new one")

		return true
	}

	return false
}

// generateJWT will create a new Apple Developer Token and store in DB
func generateJWT() (*firebase.FirestoreAppleDevJWT, error) {
	// Env Props
	appleTeamKey := os.Getenv("APPLE_TEAM_ID")
	appleKID := os.Getenv("APPLE_KID")

	// Read .p8 file
	signKeyByte, signKeyByteErr := ioutil.ReadFile(p8FilePath)
	if signKeyByteErr != nil {
		return nil, fmt.Errorf("Could not read .p8 file: %v", signKeyByteErr)
	}

	// The issued at (iat) registered claim key, whose value indicates the time at which the token was generated, in terms of the number of seconds since Epoch, in UTC
	sixMonthsInSec := 15777000
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(time.Second * time.Duration(sixMonthsInSec))

	// Setup Claims
	claims := jwt.StandardClaims{
		Issuer:    appleTeamKey,
		ExpiresAt: int64(expiresAt.Unix()),
		// SagNurSchwitzer - "WHO WILL FIX ME NOW" (05/06/20)
		IssuedAt: int64(issuedAt.Unix()),
	}

	// Generate and sign JWT
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header = map[string]interface{}{
		"alg": "ES256",
		"kid": appleKID,
	}

	// Decode block
	block, _ := pem.Decode(signKeyByte)

	// Create proper key
	signingKey, signingKeyError := x509.ParsePKCS8PrivateKey(block.Bytes)
	if signingKeyError != nil {
		return nil, fmt.Errorf("Could not parse Private key: %v", signingKeyError)
	}

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot sign token: %v", err)
	}

	// Create object
	appleDevToken := firebase.FirestoreAppleDevJWT{
		ExpiresAt: claims.ExpiresAt,
		IssuedAt:  claims.IssuedAt,
		Token:     ss,
	}

	// Write dev token to DB
	writeError := writeAppleDevToken(appleDevToken)
	if writeError != nil {
		return nil, writeError
	}

	return &appleDevToken, nil
}

// writeAppleDevToken writes the newly created JWT to our database
func writeAppleDevToken(JWT firebase.FirestoreAppleDevJWT) error {
	appleJWTDoc := firestoreClient.Collection("internal_tokens").Doc("appleDevToken")
	if appleJWTDoc == nil {
		return fmt.Errorf("appleDevToken could not be found")
	}

	jwtInterface := map[string]interface{}{
		"expiresAt": JWT.ExpiresAt,
		"issuedAt":  JWT.IssuedAt,
		"token":     JWT.Token,
	}

	_, writeErr := appleJWTDoc.Set(context.Background(), jwtInterface, firestore.MergeAll)
	if writeErr != nil {
		return fmt.Errorf(writeErr.Error())
	}

	return nil
}
