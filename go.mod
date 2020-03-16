module github.com/pixelogicdev/gruveebackend

go 1.13

require (
	cloud.google.com/go/firestore v1.1.1
	github.com/GoogleCloudPlatform/functions-framework-go v1.0.0
	github.com/pixelogicdev/gruveebackend/cmd/createuser v0.0.0-00010101000000-000000000000
	github.com/pixelogicdev/gruveebackend/cmd/socialplatform v0.0.0
	github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh v0.0.0-00010101000000-000000000000
	github.com/pixelogicdev/gruveebackend/cmd/spotifyauth v0.0.0-20200308212314-0462fa42269c
	github.com/pixelogicdev/gruveebackend/cmd/tokengen v0.0.0-20200308212314-0462fa42269c
	github.com/pixelogicdev/gruveebackend/pkg/firebase v0.0.0-20200308213401-073e9c1ba1b9
)

// ENABLE WHEN IN DEBUG
replace github.com/pixelogicdev/gruveebackend/cmd/spotifyauth => ./cmd/spotifyauth

replace github.com/pixelogicdev/gruveebackend/cmd/tokengen => ./cmd/tokengen

replace github.com/pixelogicdev/gruveebackend/cmd/socialplatform => ./cmd/socialplatform

replace github.com/pixelogicdev/gruveebackend/cmd/createuser => ./cmd/createuser

replace github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh => ./cmd/socialtokenrefresh

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ./pkg/firebase
