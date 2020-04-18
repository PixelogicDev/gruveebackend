module github.com/pixelogicdev/gruveebackend/cmd/getspotifymedia

go 1.13

replace github.com/pixelogicdev/gruveebackend/cmd/socialplatform => ../cmd/socialplatform

replace github.com/pixelogicdev/gruveebackend/cmd/spotifyauth => ../cmd/spotifyauth

replace github.com/pixelogicdev/gruveebackend/cmd/createuser => ../cmd/createuser

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../pkg/firebase

replace github.com/pixelogicdev/gruveebackend/cmd/tokengen => ../cmd/tokengen

replace github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh => ../cmd/socialtokenrefresh

replace github.com/pixelogicdev/gruveebackend/pkg/social => ../../pkg/social

replace github.com/pixelogicdev/gruveebackend/cmd/createsocialplaylist => ../cmd/createsocialplaylist

replace github.com/pixelogicdev/gruveebackend/cmd/updatealgolia => ../cmd/updatealgolia

require (
	cloud.google.com/go/firestore v1.1.1
	github.com/pixelogicdev/gruveebackend/pkg/firebase v1.0.0-beta.2
	github.com/pixelogicdev/gruveebackend/pkg/social v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.21.1
)
