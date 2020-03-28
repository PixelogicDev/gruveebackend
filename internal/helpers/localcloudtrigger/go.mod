module github.com/pixelogicdev/gruveebackend/internal/helpers/localcloudtrigger

go 1.13

require (
	github.com/cloudevents/sdk-go v1.1.2
	github.com/pixelogicdev/gruveebackend/pkg/firebase v1.0.0-beta.3
)

replace github.com/pixelogicdev/gruveebackend/pkg/firebase => ../../../pkg/firebase
