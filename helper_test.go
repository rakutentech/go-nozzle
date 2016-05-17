package nozzle

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

func NewDopplerServer(t *testing.T, inputCh <-chan []byte, authToken string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != authToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		upgrader := websocket.Upgrader{
			// Accept all origin
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// Should not reach here
			t.Fatalf("err: %s", err)
		}
		defer ws.Close()
		defer ws.WriteControl(websocket.CloseMessage, []byte(""), time.Time{})

		for input := range inputCh {
			if err := ws.WriteMessage(websocket.BinaryMessage, input); err != nil {
				// Should not reach here
				t.Fatalf("err: %s", err)
			}
		}
	}))
}

func NewEvent(message string, timestamp int64) ([]byte, error) {
	logMessage := &events.LogMessage{
		Message:     []byte(message),
		MessageType: events.LogMessage_OUT.Enum(),
		AppId:       proto.String("my-app-guid"),
		SourceType:  proto.String("DEA"),
		Timestamp:   proto.Int64(timestamp),
	}

	return proto.Marshal(&events.Envelope{
		LogMessage: logMessage,
		EventType:  events.Envelope_LogMessage.Enum(),
		Origin:     proto.String("fake-origin-1"),
		Timestamp:  proto.Int64(timestamp),
	})

}
