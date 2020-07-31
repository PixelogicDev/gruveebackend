// itshaydendev - "Firebase is bad but I use it anyway because I'm some kind of masochist" (04/26/20)
package main

// creativenobu - "compiled but feels interpreted (02/26/20)
// pheonix_d123 - "Felt Compiled. Might interpret later" (02/26/20)
// sillyonly - "YOU ALWAYS CLEAN MASTER BY FORCE PUSHING THE PERFECT CODE AND NOT THE CODE YOU WROTE" (02/23/20)
// sillyonly - "OR PUSH AFTER AN APPROVED PR" (02/23/20)
// no_neon_one - "have you tried Flutter?" (02/26/20)
// MrDemonWolf - "A Furry was here OwO" (03/08/20)
import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/joho/godotenv"
	"github.com/pixelogicdev/gruveebackend/cmd/appleauth"
	"github.com/pixelogicdev/gruveebackend/cmd/createappledevtoken"
	"github.com/pixelogicdev/gruveebackend/cmd/createprovideruser"
	"github.com/pixelogicdev/gruveebackend/cmd/createsocialplaylist"
	"github.com/pixelogicdev/gruveebackend/cmd/createuser"
	"github.com/pixelogicdev/gruveebackend/cmd/doesuserdocexist"
	"github.com/pixelogicdev/gruveebackend/cmd/getapplemusicmedia"
	"github.com/pixelogicdev/gruveebackend/cmd/getspotifymedia"
	"github.com/pixelogicdev/gruveebackend/cmd/socialplatform"
	"github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh"
	"github.com/pixelogicdev/gruveebackend/cmd/spotifyauth"
	"github.com/pixelogicdev/gruveebackend/cmd/tokengen"
	"github.com/pixelogicdev/gruveebackend/cmd/updatealgolia"
	"github.com/pixelogicdev/gruveebackend/cmd/usernameavailable"
)

func init() {
	// Load in ENV file
	goEnvErr := godotenv.Load("./internal/config.yaml")
	if goEnvErr != nil {
		log.Printf("Main [Load GoEnv]: %v\n", goEnvErr)
	}
	log.Println("Main environment variables loaded")
}

// InukApp - "Swift > Go" (03/15/20)
// Fr3fou - "i helped build this AYAYA, follow @fr3fou on twitter uwu" (04/07/20)
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// General Endpoints
	funcframework.RegisterHTTPFunction("/authorizeWithApple", appleauth.AuthorizeWithApple)
	funcframework.RegisterHTTPFunction("/authorizeWithSpotify", spotifyauth.AuthorizeWithSpotify)
	funcframework.RegisterHTTPFunction("/generateCustomToken", tokengen.GenerateCustomToken)
	funcframework.RegisterHTTPFunction("/createSocialPlatform", socialplatform.CreateSocialPlatform)
	funcframework.RegisterHTTPFunction("/createUser", createuser.CreateUser)
	funcframework.RegisterHTTPFunction("/socialTokenRefresh", socialtokenrefresh.SocialTokenRefresh)
	funcframework.RegisterHTTPFunction("/createSocialPlaylist", createsocialplaylist.CreateSocialPlaylist)
	funcframework.RegisterHTTPFunction("/createAppleDevToken", createappledevtoken.CreateAppleDevToken)
	funcframework.RegisterHTTPFunction("/doesUserDocExist", doesuserdocexist.DoesUserDocExist)
	funcframework.RegisterHTTPFunction("/usernameAvailable", usernameavailable.UsernameAvailable)
	funcframework.RegisterHTTPFunction("/createProviderUser", createprovideruser.CreateProviderUser)
	funcframework.RegisterEventFunction("/updateAlgolia", updatealgolia.UpdateAlgolia)

	// Get Media Endpoints
	funcframework.RegisterHTTPFunction("/getSpotifyMedia", getspotifymedia.GetSpotifyMedia)
	funcframework.RegisterHTTPFunction("/getAppleMusicMedia", getapplemusicmedia.GetAppleMusicMedia)

	// WIP: Local trigger endpoint for cloud event
	// funcframework.RegisterHTTPFunction("/localCloudTrigger", localcloudtrigger.LocalCloudTrigger)

	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
