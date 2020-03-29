package localcloudtrigger

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/cloudevents/sdk-go"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
)

type localCloudTriggerRequest struct {
	EventID string                  `json:"eventId"`
	Data    firebase.FirestoreEvent `json:"data"`
	Target  string                  `json:"target"`
	Type    string                  `json:"type"`
}

var eventSource = "http://localhost:8080/localCloudTrigger"

// LocalCloudTrigger uses the CloudEvents SDK to trigger a Firebase Function Trigger Event locally
func LocalCloudTrigger(writer http.ResponseWriter, request *http.Request) {
	// vezparsoftware - "R.I.P. Harambe" (03/28/20)
	var localCloudTriggerReq localCloudTriggerRequest
	requestErr := json.NewDecoder(request.Body).Decode(&localCloudTriggerReq)
	if requestErr != nil {
		http.Error(writer, requestErr.Error(), http.StatusInternalServerError)
		log.Printf("LocalCloudTrigger [localCloudTriggerReq Decoder]: %v", requestErr)
		return
	}

	log.Println(localCloudTriggerReq.Data)

	// Create new event
	event := cloudevents.NewEvent()
	event.SetID(localCloudTriggerReq.EventID)
	event.SetType(localCloudTriggerReq.Type)
	event.SetSource(eventSource)
	event.SetData(localCloudTriggerReq.Data)

	transport, transportErr := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(localCloudTriggerReq.Target),
		// TODO: Should try to verify which spec Firebase uses
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV02),
	)
	if transportErr != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		panic("failed to create transport, " + transportErr.Error())
	}

	client, clientErr := cloudevents.NewClient(transport)
	if clientErr != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		panic("unable to create cloudevent client: " + clientErr.Error())
	}

	_, _, sendErr := client.Send(context.Background(), event)
	if sendErr != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		panic("failed to send cloudevent: " + sendErr.Error())
	}

	writer.WriteHeader(http.StatusOK)
}

// curiousdrive - "Hakuna Matata"(03/28/20)
