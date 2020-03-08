module github.com/pixelogicdev/gruveebackend

go 1.13

require (
	github.com/GoogleCloudPlatform/functions-framework-go v1.0.0
	github.com/pixelogicdev/gruveebackend/cmd/spotifyauth v0.0.0-00010101000000-000000000000
	github.com/pixelogicdev/gruveebackend/cmd/tokengen v0.0.0-00010101000000-000000000000
	github.com/pixelogicdev/gruveebackend/pkg/firebase v0.0.0-00010101000000-000000000000 // indirect
)

// Needed for local redirect
replace github.com/pixelogicdev/gruveebackend/cmd/spotifyauth => ./cmd/spotifyauth

replace github.com/pixelogicdev/gruveebackend/cmd/tokengen => ./cmd/tokengen

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ./pkg/firebase
