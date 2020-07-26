package sawmill

// If you're wondering why this package is named sawmill, I do too: https://clips.twitch.tv/PlacidOutstandingPelicanCharlieBitMe
import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/logging"
	"google.golang.org/api/option"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

// Logger holds the newly generated error logging client, so that it can be used for our own, more simple custom logging functions instead of the ones the Google provides
type Logger struct {
	// The Logger
	GoogleLogger *logging.Logger
	// If their was an error with the initialization, it won't log to the cloud
	InitError bool
	// The service that logging was created for
	ServiceName string
}

// InitClient creates a logging client
func InitClient(projectID string, credentialsJSON string, environment string, serviceName string) (Logger, error) {
	// The client takes the credentials in a byte array, so we do that conversion here
	credentialsByte := []byte(credentialsJSON)
	ctx := context.Background()

	// Initializes an Google logging client
	loggingClient, err := logging.NewClient(ctx, projectID, option.WithCredentialsJSON(credentialsByte)) // WithCredentialsJSON takes a raw JSON string in a byte array

	if err != nil {
		// If there is an error with initialization, it returns a Client object containing an empty ErrorClient, so nothing will get logged to the cloud, only the console
		return Logger{nil, true, ""}, err
	}

	loggingClient.OnError = func(err error) {
		log.Printf("Logging error in "+serviceName+": %v", err)
	}

	// Creates the actual logger
	logger := loggingClient.Logger(serviceName)

	// If we are in a dev environment, we don't want to log to the cloud, so this variable is used as the value for InitError, and if is set to true, nothing will be logged to the cloud
	var isDev bool = false

	// Here the aforementioned variable is set to true if we are in the DEV environment
	if environment == "DEV" {
		isDev = true
	}

	return Logger{logger, isDev, serviceName}, nil
}

// This is an empty Entry struct, which is used to substitute in for any of the optional arguments that are not passed
var emptyEntry logging.Entry

// LogErr logs an error
func (c Logger) LogErr(err error, operation string, req *http.Request) {
	// Logs the error to the terminal
	log.Printf("%s ["+operation+"]: %v ", c.ServiceName, err)

	// Checks if there was an error with the initialization of the Cloud Logging Client
	if c.InitError {
		return
	}

	// requestStruct holds the value for the HTTPRequest, if req isn't passed, it will stay empty
	requestStruct := emptyEntry.HTTPRequest

	// If req is passed, this sets requestStruct to a structure containing req
	if req != nil {
		requestStruct = &logging.HTTPRequest{
			Request: req,
		}
	}

	// Reports the error
	c.GoogleLogger.Log(logging.Entry{
		Payload: err.Error(),
		Operation: &logpb.LogEntryOperation{
			Id: operation,
		},
		HTTPRequest: requestStruct,
	})

	// Sends the log to the cloud
	c.GoogleLogger.Flush()
}
