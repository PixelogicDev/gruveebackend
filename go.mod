module github.com/pixelogicdev/gruveebackend

go 1.13

require (
	cloud.google.com/go/firestore v1.2.0
	github.com/GoogleCloudPlatform/functions-framework-go v1.0.0
	github.com/cloudevents/sdk-go v1.1.2 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/joho/godotenv v1.3.0
	github.com/pixelogicdev/gruveebackend/cmd/appleauth v1.0.0-beta.1
	github.com/pixelogicdev/gruveebackend/cmd/createappledevtoken v1.0.0-beta.1
	github.com/pixelogicdev/gruveebackend/cmd/createsocialplaylist v1.0.0-beta.6
	github.com/pixelogicdev/gruveebackend/cmd/createuser v1.0.0-beta.2
	github.com/pixelogicdev/gruveebackend/cmd/getspotifymedia v1.0.0-beta.2
	github.com/pixelogicdev/gruveebackend/cmd/socialplatform v1.0.0-beta.2
	github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh v1.0.0-beta.4
	github.com/pixelogicdev/gruveebackend/cmd/spotifyauth v1.0.0-beta.5
	github.com/pixelogicdev/gruveebackend/cmd/tokengen v1.0.0-beta.3
	github.com/pixelogicdev/gruveebackend/cmd/updatealgolia v1.0.0-beta.8
	github.com/pixelogicdev/gruveebackend/pkg/firebase v1.0.0-beta.12
	github.com/pixelogicdev/gruveebackend/pkg/social v1.0.0-beta.5
	google.golang.org/grpc v1.28.0
)

replace github.com/pixelogicdev/gruveebackend/cmd/appleauth => ./cmd/appleauth

replace github.com/pixelogicdev/gruveebackend/cmd/spotifyauth => ./cmd/spotifyauth

replace github.com/pixelogicdev/gruveebackend/cmd/tokengen => ./cmd/tokengen

replace github.com/pixelogicdev/gruveebackend/cmd/socialplatform => ./cmd/socialplatform

replace github.com/pixelogicdev/gruveebackend/cmd/createuser => ./cmd/createuser

replace github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh => ./cmd/socialtokenrefresh

replace github.com/pixelogicdev/gruveebackend/cmd/createsocialplaylist => ./cmd/createsocialplaylist

replace github.com/pixelogicdev/gruveebackend/cmd/updatealgolia => ./cmd/updatealgolia

replace github.com/pixelogicdev/gruveebackend/cmd/getspotifymedia => ./cmd/getspotifymedia

replace github.com/pixelogicdev/gruveebackend/cmd/createappledevtoken => ./cmd/createappledevtoken

// WIP
replace github.com/pixelogicdev/gruveebackend/internal/helpers/localcloudtrigger => ./internal/helpers/localcloudtrigger

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ./pkg/firebase

replace github.com/pixelogicdev/gruveebackend/pkg/social => ./pkg/social
