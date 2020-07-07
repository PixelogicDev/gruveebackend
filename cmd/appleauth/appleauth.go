package appleauth

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
	"github.com/unrolled/render"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// zebcode - "Zebcode Rules 🦸‍♂️" (04/29/20)
type appleDevTokenResp struct {
	Token string
}

var firestoreClient *firestore.Client
var appleDevToken firebase.FirestoreAppleDevJWT
var httpClient *http.Client
var currentP8Path string
var hostname string

func init() {
	httpClient = &http.Client{}
	log.Println("AuthorizeWithApple initialized.")
}

// AuthorizeWithApple will render a HTML page to get the AppleMusic credentials for user
func AuthorizeWithApple(writer http.ResponseWriter, request *http.Request) {
	if os.Getenv("APPLE_TEAM_ID") == "" {
		http.Error(writer, "[AuthorizeWithApple] APPLE_TEAM_ID does not exist!", http.StatusInternalServerError)
		log.Println("[AuthorizeWithApple] APPLE_TEAM_ID does not exist!")
		return
	}

	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		log.Printf("[AuthorizeWithApple] [initWithEnv]: %v", initWithEnvErr)
		return
	}

	// Grab token from DB && check expiration
	// Create Request
	createAppleDevURI := hostname + "/createAppleDevToken"
	log.Println(createAppleDevURI)
	appleDevTokenReq, appleDevTokenReqErr := http.NewRequest("GET", createAppleDevURI, nil)
	if appleDevTokenReqErr != nil {
		http.Error(writer, appleDevTokenReqErr.Error(), http.StatusInternalServerError)
		log.Println(appleDevTokenReqErr.Error())
		return
	}

	appleDevTokenResp, appleDevTokenRespErr := httpClient.Do(appleDevTokenReq)
	if appleDevTokenRespErr != nil {
		http.Error(writer, appleDevTokenRespErr.Error(), http.StatusInternalServerError)
		log.Println(appleDevTokenRespErr.Error())
		return
	}

	// Decode Token
	var appleDevToken firebase.FirestoreAppleDevJWT

	appleDevTokenDecodeErr := json.NewDecoder(appleDevTokenResp.Body).Decode(&appleDevToken)
	if appleDevTokenDecodeErr != nil {
		http.Error(writer, appleDevTokenDecodeErr.Error(), http.StatusInternalServerError)
		log.Printf("AuthorizeWithSpotify [spotifyAuthRequest Decoder]: %v", appleDevTokenDecodeErr)
		return
	}

	/* foundAppleDevToken, foundAppleDevTokenErr := getAppleDevToken()
	if foundAppleDevTokenErr != nil {
		http.Error(writer, foundAppleDevTokenErr.Error(), http.StatusInternalServerError)
		log.Printf("[AuthorizeWithApple] [getAppleDevToken]: %v", foundAppleDevTokenErr)
		return
	}

	// Check for developer token - Let's call our firebase function
	if foundAppleDevToken != nil {
		log.Println("[AuthorizeWithApple] Token already found. Sending back.")

		// TODO: We should check if it's expired and if so, create a new one

		appleDevToken = *foundAppleDevToken
	}

	if foundAppleDevToken == nil && foundAppleDevTokenErr == nil {
		log.Println("[AuthorizeWithApple] No AppleDevToken found in DB")
		createdAppleDevtoken, createdAppleDevtokenErr := createAppleDevToken()
		if createdAppleDevtokenErr != nil {
			http.Error(writer, createdAppleDevtokenErr.Error(), http.StatusInternalServerError)
			log.Printf("[AuthorizeWithApple] [createAppleDevToken]: %v", createdAppleDevtokenErr)
			return
		}

		appleDevToken = *createdAppleDevtoken
	} */

	// Render template
	render := render.New(render.Options{
		Directory: "cmd/appleauth/templates",
	})

	renderErr := render.HTML(writer, http.StatusOK, "auth", appleDevToken)
	if renderErr != nil {
		http.Error(writer, renderErr.Error(), http.StatusInternalServerError)
		log.Printf("[AuthorizeWithApple] [render]: %v", renderErr)
		return
	}
}

// Helpers
// initWithEnv takes our yaml env variables and maps them properly.
// Unfortunately, we had to do this is main because in init we weren't able to access env variables
func initWithEnv() error {
	// Get paths
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
		currentP8Path = os.Getenv("APPLE_P8_PATH_DEV")
		hostname = os.Getenv("HOSTNAME_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
		currentP8Path = os.Getenv("APPLE_P8_PATH_PROD")
		hostname = os.Getenv("HOSTNAME_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("SocialTokenRefresh [Init Firestore]: %v", err)
	}

	firestoreClient = client
	return nil
}

// getAppleDevToken will check our DB for appleDevJWT and return it if there
func getAppleDevToken() (*firebase.FirestoreAppleDevJWT, error) {
	// Go to Firebase and see if appleDevToken exists
	snapshot, snapshotErr := firestoreClient.Collection("internal_tokens").Doc("appleDevToken").Get(context.Background())
	if status.Code(snapshotErr) == codes.NotFound {
		log.Println("[AppleAuth] AppleDevToken not found in DB.")
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

// createAppleDevToken creates our actual JWT needed to make API requests
func createAppleDevToken() (*firebase.FirestoreAppleDevJWT, error) {
	// Env Props
	appleTeamKey := os.Getenv("APPLE_TEAM_ID")
	appleKID := os.Getenv("APPLE_KID")

	// Read .p8 file
	signKeyByte, signKeyByteErr := ioutil.ReadFile(currentP8Path)
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
		IssuedAt:  int64(issuedAt.Unix()),
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
