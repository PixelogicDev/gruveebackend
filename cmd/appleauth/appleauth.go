package appleauth

import (
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
	"github.com/pixelogicdev/gruveebackend/pkg/sawmill"
	"github.com/unrolled/render"
)

// zebcode - "Zebcode Rules 🦸‍♂️" (04/29/20)
var (
	firestoreClient *firestore.Client
	logger          sawmill.Logger
	appleDevToken   firebase.FirestoreAppleDevJWT
	httpClient      *http.Client
	hostname        string
	templatePath    string
)

func init() {
	log.Println("AuthorizeWithApple initialized.")
}

// AuthorizeWithApple will render a HTML page to get the AppleMusic credentials for user
func AuthorizeWithApple(writer http.ResponseWriter, request *http.Request) {
	// Initialize
	initWithEnvErr := initWithEnv()
	if initWithEnvErr != nil {
		http.Error(writer, initWithEnvErr.Error(), http.StatusInternalServerError)
		logger.LogErr("InitWithEnv", initWithEnvErr, nil)
		return
	}

	// DR_DinoMight: Dammmmn, Apple Really?!?!?! (08/11/20)
	appleDevToken, appleDevTokenErr := firebase.GetAppleDeveloperToken()
	if appleDevTokenErr != nil {
		http.Error(writer, appleDevTokenErr.Error(), http.StatusInternalServerError)
		logger.LogErr("GetAppleMusicMedia", appleDevTokenErr, nil)
		return
	}

	logger.Log("GetAppleDeveloperToken", "AppleDevToken recieved.")

	// Render template
	render := render.New(render.Options{
		Directory: templatePath,
	})
	renderErr := render.HTML(writer, http.StatusOK, "auth", appleDevToken)
	if renderErr != nil {
		http.Error(writer, renderErr.Error(), http.StatusInternalServerError)
		logger.LogErr("Render", renderErr, nil)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
