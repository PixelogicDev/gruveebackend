module github.com/pixelogicdev/gruveebackend/cmd/createappledevtoken

go 1.13

require (
	cloud.google.com/go/firestore v1.2.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/pixelogicdev/gruveebackend/pkg/firebase v1.0.0-beta.11
	google.golang.org/grpc v1.28.0
)

replace github.com/pixelogicdev/gruveebackend/cmd/socialplatform => ../cmd/socialplatform

replace github.com/pixelogicdev/gruveebackend/cmd/spotifyauth => ../cmd/spotifyauth

replace github.com/pixelogicdev/gruveebackend/cmd/createuser => ../cmd/createuser

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../pkg/firebase

replace github.com/pixelogicdev/gruveebackend/cmd/tokengen => ../cmd/tokengen

replace github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh => ../cmd/socialtokenrefresh

replace github.com/pixelogicdev/gruveebackend/pkg/social => ../../pkg/social

replace github.com/pixelogicdev/gruveebackend/cmd/createsocialplaylist => ../cmd/createsocialplaylist

replace github.com/pixelogicdev/gruveebackend/cmd/updatealgolia => ../cmd/updatealgolia

replace github.com/pixelogicdev/gruveebackend/cmd/getspotifymedia => ../cmd/getspotifymedia

replace github.com/pixelogicdev/gruveebackend/cmd/appleauth => ../cmd/appleauth
