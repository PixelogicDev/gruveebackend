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
	Logger *logging.Logger // The Logger

	InitError bool // If their was an error with the initialization, it won't log to the cloud

	ServiceName string // The service that logging was created for
}

// InitClient creates a logging client
func InitClient(projectID string, credentials string, environment string, serviceName string) (Logger, error) {
	// Checks if were in the development environment, because if we are Cloud Logging is not needed
	if environment == "DEV" {
		return Logger{nil, false, serviceName}, nil
	}

	// The client takes the credentials in a byte array, so we do that conversion here
	credentialsByte := []byte(credentials)
	ctx := context.Background()

	// Initializes an Google logging client
	loggingClient, err := logging.NewClient(ctx, projectID, option.WithCredentialsJSON(credentialsByte))

	if err != nil {
		// If there is an error with initialization, it returns a Client object containing an empty ErrorClient, so nothing will get logged to the cloud, only the console
		return Logger{nil, true, ""}, err
	}

	loggingClient.OnError = func(err error) {
		log.Printf("Logging error in "+serviceName+": %v", err)
	}

	// Creates the actual logger
	logger := loggingClient.Logger(serviceName)

	return Logger{logger, false, serviceName}, nil
}

// This is an empty Entry struct, which is used to substitute in for any of the optional arguments that are not passed
var emptyEntry logging.Entry

// LogErr logs an error
func (c Logger) LogErr(err error, operation string, req *http.Request) {
	// Logs the error to the terminal
	log.Printf(c.ServiceName + " [" + operation + "]:", err)

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
	c.Logger.Log(logging.Entry{
		Payload: err.Error(),
		Operation: &logpb.LogEntryOperation{
			Id: operation,
		},
		HTTPRequest: requestStruct,
	})

	// Sends the log to the cloud
	c.Logger.Flush()
}