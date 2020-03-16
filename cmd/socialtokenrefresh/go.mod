module github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh

go 1.13

require (
	cloud.google.com/go/firestore v1.1.1
	github.com/pixelogicdev/gruveebackend/pkg/firebase v0.0.0-20200308213401-073e9c1ba1b9
)

replace github.com/pixelogicdev/gruveebackend/cmd/socialplatform => ../cmd/socialplatform

replace github.com/pixelogicdev/gruveebackend/cmd/spotifyauth => ../cmd/spotifyauth

replace github.com/pixelogicdev/gruveebackend/cmd/createuser => ../cmd/createuser

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../pkg/firebase

replace github.com/pixelogicdev/gruveebackend/cmd/tokengen => ../cmd/tokengen
