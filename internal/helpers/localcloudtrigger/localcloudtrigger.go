package localcloudtrigger

import (
	"context"
	"net/http"

	"github.com/cloudevents/sdk-go"
	"github.com/pixelogicdev/gruveebackend/pkg/firebase"
)

type localCloudTriggerRequest struct {
	Data   firebase.FirestoreEvent `json:"data"`
	Source string                  `json:"source"`
	Type   string                  `json:"type"`
}

// LocalCloudTrigger uses the CloudEvents SDK to trigger a Firebase Function Trigger Event locally
func LocalCloudTrigger(writer http.ResponseWriter, request *http.Request) {
	// vezparsoftware - "R.I.P. Harambe" (03/28/20)
	/*
		Source: The Local Trigger path we want to call
		Data: What we expect to send to our trigger
	*/

	// var localCloudTriggerReq localCloudTriggerRequest

	// requestErr := json.NewDecoder(request.Body).Decode(&localCloudTriggerReq)
	// if requestErr != nil {
	// 	http.Error(writer, requestErr.Error(), http.StatusInternalServerError)
	// 	log.Printf("LocalCloudTrigger [localCloudTriggerReq Decoder]: %v", requestErr)
	// 	return
	// }

	event := cloudevents.NewEvent()
	event.SetID("Gruvee123")
	event.SetType("providers/cloud.firestore/eventTypes/document.create")
	event.SetSource("http://localhost:8080")
	event.SetData(firebase.FirestoreEvent{})

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget("http://localhost:8080/updateAlgolia"),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV02),
	)
	if err != nil {
		panic("failed to create transport, " + err.Error())
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		panic("unable to create cloudevent client: " + err.Error())
	}
	if _, _, err := c.Send(context.Background(), event); err != nil {
		panic("failed to send cloudevent: " + err.Error())
	}

	// "providers/cloud.firestore/eventTypes/document.create"
	// event.SetType(localCloudTriggerReq.Type)
	// event.SetSource(localCloudTriggerReq.Source)
	// event.SetData(localCloudTriggerReq.Data)
}

// curiousdrive - "Hakuna Matata"(03/28/20)
