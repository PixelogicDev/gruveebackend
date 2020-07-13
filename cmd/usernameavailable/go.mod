module github.com/pixelogicdev/gruveebackend/cmd/usernameavailable

go 1.13

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../pkg/firebase

replace github.com/pixelogicdev/gruveebackend/pkg/social => ../../pkg/social

require cloud.google.com/go/firestore v1.2.0
