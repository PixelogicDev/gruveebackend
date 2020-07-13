package errlog

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/errorreporting"
	"google.golang.org/api/option"
)

// Client holds the newly generated error logging client, so that it can be used for our own, more simple custom logging functions instead of the ones the Google provides
type Client struct {
	ErrorClient *errorreporting.Client // The client generated by Google

	// If their was an error with the initialization, WasError will be true, so when an error logging function is called, nothing will happen
	// If the initialization fails, it will just print the same error that we already log when the Client is initialized
	// Basically WasError just reduces log clutter
	WasError bool
}

// InitErrClient accesses the enviroment variables on its own, and creates an error logging client
func InitErrClient(service string) (Client, error) {
	// Gets correct enviroment variable
	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
	}

	// Returns the Client struct
	return initClient(currentProject, os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), service)
}

// InitErrClientWithEnv uses the envirorment variable passed to it to create an error logging client
func InitErrClientWithEnv(currentProject string, credentials string, service string) (Client, error) {
	// Returns the Client struct
	return initClient(currentProject, credentials, service)
}

// initClient takes all the arguments needed, and initializes the Google logging client, and then uses that to create an instance of our own Client
func initClient(currentProject string, credentials string, service string) (Client, error) {
	ctx := context.Background()

	// initializes an Google logging client
	errorClient, err := errorreporting.NewClient(ctx, currentProject, errorreporting.Config{
		ServiceName: service,
		// This is called on a logging error
		OnError: func(err error) {
			log.Printf(service+" [errlog.LogErr]: %v", err)
		},
	}, option.WithCredentialsFile(credentials))
	if err != nil {
		// If there is an error with initialization, it returns a Client object containing an empty ErrorClient, and sets WasError to be true, so nothing gets reported, which would trigger an error message
		return Client{nil, true}, err
	}

	// If there wasn't an error, it returns a Client object with an ErrorClient, and sets WasError to be false
	return Client{errorClient, false}, nil
}

// LogErr logs an error with Google Cloud, included a request object
func (c Client) LogErr(err error) {
	// Checks if there was an error with initialization
	if c.WasError {
		return
	}

	// Reports the error
	c.ErrorClient.Report(errorreporting.Entry{
		Error: err,
	})

	// As every single error log is followed by a return statement, no more errors get logged after the first, so we close the client
	c.ErrorClient.Close()
}

// LogErrReq logs an error with Google Cloud, included a request object
func (c Client) LogErrReq(err error, req *http.Request) {
	// Checks if there was an error with initialization
	if c.WasError {
		return
	}

	// Reports the error
	c.ErrorClient.Report(errorreporting.Entry{
		Error: err,
		Req:   req,
	})

	// As every single error log is followed by a return statement, no more errors get logged after the first, so we close the client
	c.ErrorClient.Close()
}

// LogErrWithUser logs an error with Google Cloud, containing information about the user
func (c Client) LogErrWithUser(err error, req *http.Request, user string) {
	// Checks if there was an error with initialization
	if c.WasError {
		return
	}

	// Reports the error
	c.ErrorClient.Report(errorreporting.Entry{
		Error: err,
		Req:   req,
		User:  user,
	})

	// As every single error log is followed by a return statement, no more errors get logged after the first, so we close the client
	c.ErrorClient.Close()
}
