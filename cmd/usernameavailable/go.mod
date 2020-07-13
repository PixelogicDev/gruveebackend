module github.com/pixelogicdev/gruveebackend/cmd/usernameavailable

go 1.13

<<<<<<< da3878c925ca6ef481f9b424c30ccaafe56a0d2e
require (
	cloud.google.com/go/firestore v1.2.0
	github.com/pixelogicdev/gruveebackend/pkg/sawmill v0.0.0-20200807181419-cd8cef80a5be
)
=======
replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../pkg/firebase

replace github.com/pixelogicdev/gruveebackend/pkg/social => ../../pkg/social

require cloud.google.com/go/firestore v1.2.0
>>>>>>> Adding new usernameavailable function
