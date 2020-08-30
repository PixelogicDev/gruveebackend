package createappledevtoken

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/dgrijalva/jwt-go"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Helpers
// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
		applePrivateKeyPath = os.Getenv("APPLE_PRIVATE_KEY_PATH_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
		applePrivateKeyPath = os.Getenv("APPLE_PRIVATE_KEY_PATH_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("CreateAppleDevToken [Init Firestore]: %v", err)
	}

	// Initialize Sawmill
	sawmillLogger, err := sawmill.InitClient(currentProject, os.Getenv("GCLOUD_CONFIG"), os.Getenv("ENVIRONMENT"), "CreateAppleDevToken")
	if err != nil {
		log.Printf("CreateAppleDevToken [Init Sawmill]: %v", err)
	}

	firestoreClient = client
	logger = sawmillLogger
	return nil
}

// fetchToken will grab the Apple Developer Token from DB
func fetchToken() (*firebase.FirestoreAppleDevJWT, error) {
	logger.Log("FetchToken", "Starting...")

	// Go to Firebase and see if appleDevToken exists
	snapshot, snapshotErr := firestoreClient.Collection("internal_tokens").Doc("appleDevToken").Get(context.Background())
	if status.Code(snapshotErr) == codes.NotFound {
		logger.Log("FetchToken", "AppleDevToken not found in DB. Need to create.")
		return nil, nil
	}

	if snapshotErr != nil {
		return nil, fmt.Errorf(snapshotErr.Error())
	}

	logger.Log("FetchToken", "Snapshot found.")

	var appleDevToken firebase.FirestoreAppleDevJWT
	dataToErr := snapshot.DataTo(&appleDevToken)
	if dataToErr != nil {
		return nil, fmt.Errorf(dataToErr.Error())
	}

	logger.Log("FetchToken", "Decoded AppleDevToken successfully.")

	return &appleDevToken, nil
}

// isTokenExpired will check to see if the Apple Developer Token is expired
func isTokenExpired(token *firebase.FirestoreAppleDevJWT) bool {
	logger.Log("IsTokenExpired", "Starting...")

	// Get current time
	var currentTime = time.Now()

	logger.Log("IsTokenExpired", fmt.Sprintf("Issued At: %d seconds", token.IssuedAt))
	logger.Log("IsTokenExpired", fmt.Sprintf("Expires At: %d seconds\n", token.ExpiresAt))

	if currentTime.After(time.Unix(token.ExpiresAt, 0)) {
		logger.Log("IsTokenExpired", "Token is expired. Need to generate a new one.")
		return true
	}

	logger.Log("IsTokenExpired", "Token is not expired.")
	return false
}

// generateJWT will create a new Apple Developer Token and store in DB
func generateJWT() (*firebase.FirestoreAppleDevJWT, error) {
	logger.Log("GenerateJWT", "Starting...")

	// Env Props
	appleTeamKey := os.Getenv("APPLE_TEAM_ID")
	appleKID := os.Getenv("APPLE_KID")

	// Download .p8 file
	signKeyByte, signKeyByteErr := ioutil.ReadFile(applePrivateKeyPath)
	if signKeyByteErr != nil {
		return nil, fmt.Errorf("Could not read .p8 file: %v", signKeyByteErr)
	}

	logger.Log("GenerateJWT", ".p8 file read successfully.")

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

	logger.Log("GenerateJWT", "JWT claims setup successfully.")

	// Generate and sign JWT
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header = map[string]interface{}{
		"alg": "ES256",
		"kid": appleKID,
	}

	logger.Log("GenerateJWT", "Generated new JWT successfully.")

	// Decode block
	block, _ := pem.Decode(signKeyByte)

	logger.Log("GenerateJWT", "Created JWT block successfully.")

	// Create proper key
	signingKey, signingKeyError := x509.ParsePKCS8PrivateKey(block.Bytes)
	if signingKeyError != nil {
		return nil, fmt.Errorf("Could not parse Private key: %v", signingKeyError)
	}

	logger.Log("GenerateJWT", "Created JWT signing key successfully.")

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot sign token: %v", err)
	}

	logger.Log("GenerateJWT", "Created signed string successfully.")

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

	logger.Log("GenerateJWT", "Dev token successfully written to Firestore.")

	return &appleDevToken, nil
}

// writeAppleDevToken writes the newly created JWT to our database
func writeAppleDevToken(JWT firebase.FirestoreAppleDevJWT) error {
	logger.Log("WriteAppleDevToken", "Starting...")

	appleJWTDoc := firestoreClient.Collection("internal_tokens").Doc("appleDevToken")
	if appleJWTDoc == nil {
		return fmt.Errorf("appleDevToken could not be found")
	}

	logger.Log("WriteAppleDevToken", "Received AppleDevtoken from Firestore Collection")

	jwtInterface := map[string]interface{}{
		"expiresAt": JWT.ExpiresAt,
		"issuedAt":  JWT.IssuedAt,
		"token":     JWT.Token,
	}

	_, writeErr := appleJWTDoc.Set(context.Background(), jwtInterface, firestore.MergeAll)
	if writeErr != nil {
		return fmt.Errorf(writeErr.Error())
	}

	logger.Log("WriteAppleDevToken", "Set Apple DevToken in Firestore.")

	return nil
}
