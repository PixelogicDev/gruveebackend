package appleauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/unrolled/render"
)

// zebcode - "Zebcode Rules 🦸‍♂️" (04/29/20)
type appleDevTokenResp struct {
	Token string
}

var firestoreClient *firestore.Client
var appleDevToken firebase.FirestoreAppleDevJWT
var httpClient *http.Client
var hostname string
var templatePath string

func init() {
	httpClient = &http.Client{}
	log.Println("AuthorizeWithApple initialized.")
}

// AuthorizeWithApple will render a HTML page to get the AppleMusic credentials for user
func AuthorizeWithApple(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		log.Printf("[AuthorizeWithApple] [initWithEnv]: %v", initWithEnvErr)
		return
	}

	// Grab Apple Developer token from DB or create a new one
	createAppleDevURI := hostname + "/createAppleDevToken"
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
		log.Printf("AppleAuth [appleDevToken Decoder]: %v", appleDevTokenDecodeErr)
		return
	}

	// Render template
	render := render.New(render.Options{
		Directory: templatePath,
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
	if os.Getenv("APPLE_TEAM_ID") == "" {
		return fmt.Errorf("authorizeWithApple - APPLE_TEAM_ID does not exist")
	}

	// Get paths
	var currentProject string
	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
		hostname = os.Getenv("HOSTNAME_DEV")
		templatePath = os.Getenv("APPLE_AUTH_TEMPLATE_PATH_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
		hostname = os.Getenv("HOSTNAME_PROD")
		templatePath = os.Getenv("APPLE_AUTH_TEMPLATE_PATH_PROD")
	}

	// Initialize Firestore
	client, err := firestore.NewClient(context.Background(), currentProject)
	if err != nil {
		return fmt.Errorf("AuthorizeWithApple [Init Firestore]: %v", err)
	}

	firestoreClient = client
	return nil
}
