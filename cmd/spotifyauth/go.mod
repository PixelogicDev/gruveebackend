module github.com/pixelogicdev/gruveebackend/cmd/spotifyauth

go 1.13

require (
	cloud.google.com/go/firestore v1.1.1
	github.com/pixelogicdev/gruveebackend/pkg/firebase v0.0.0-20200308212314-0462fa42269c
	google.golang.org/grpc v1.27.1
)

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../pkg/firebase
