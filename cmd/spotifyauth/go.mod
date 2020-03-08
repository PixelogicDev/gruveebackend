module github.com/pixelogicdev/gruveebackend/cmd/spotifyauth

go 1.13

require (
	cloud.google.com/go/firestore v1.1.1
	google.golang.org/grpc v1.27.1
)

// TODO: Will need to remove this once deplyed in Github
replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ./pkg/firebase
