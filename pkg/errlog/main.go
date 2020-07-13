package errlog

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/errorreporting"
)

//ErrClient holds the newly generated error logging client, so that it can be used for our own, more simple custom logging functions instead of the ones the Google provides
type ErrClient struct {
	Client *errorreporting.Client
	Error  bool
}

//InitClient creates an error logging client
func InitClient(service string) (ErrClient, error) {
	ctx := context.Background()

	var currentProject string

	if os.Getenv("ENVIRONMENT") == "DEV" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_DEV")
	} else if os.Getenv("ENVIRONMENT") == "PROD" {
		currentProject = os.Getenv("FIREBASE_PROJECTID_PROD")
	}

	errorClient, err := errorreporting.NewClient(ctx, currentProject, errorreporting.Config{
		ServiceName: service,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		return ErrClient{nil, true}, err
	}

	return ErrClient{errorClient, false}, nil
}

//LogErr logs an err with Firebase
func (c ErrClient) LogErr(req *http.Request, err error) {
	if c.Error {
		return
	}
	c.Client.Report(errorreporting.Entry{
		Error: err,
		Req:   req,
	})
}
