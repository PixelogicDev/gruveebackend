// Remaiten - “Maybe now is the time to switch to Vue Native Kappa” (02/22/20)
// BackeyM - "It's go time 🙃" (02/22/20)
// TheDkbay - "It's double go time" (02/22/20)
package tokengen

import (
	"context"
	"net/http"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func init() {
	ctx := context.Background()
	conf := option.WithCredentialsFile("../../config/gruveeAdminPrivateKey.json")

	// Initialize firebase admin
	_, err := firebase.NewApp(ctx, nil, conf)
	
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}

	fmt.Println("We made it.")
}

// GenerateToken does a thing (THIS IS A TEST)
func GenerateToken(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "HI TOKEN GENERATOR!\n")
	writer.Write([]byte("token"))
}