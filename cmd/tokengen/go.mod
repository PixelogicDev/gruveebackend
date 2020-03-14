module github.com/pixelogicdev/gruveebackend/cmd/tokengen

go 1.13

require (
	cloud.google.com/go/storage v1.6.0 // indirect
	firebase.google.com/go v3.12.0+incompatible
	github.com/pixelogicdev/gruveebackend/pkg/firebase v0.0.0-20200308213401-073e9c1ba1b9
	google.golang.org/api v0.20.0 // indirect
)

// ENABLE WHEN IN DEBUG
replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../pkg/firebase

replace github.com/pixelogicdev/gruveebackend/cmd/socialplatform => ../cmd/socialplatform

replace github.com/pixelogicdev/gruveebackend/cmd/spotifyauth => ../cmd/spotifyauth

replace github.com/pixelogicdev/gruveebackend/cmd/createuser => ../cmd/createuser
