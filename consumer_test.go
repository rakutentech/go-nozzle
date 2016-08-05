package nozzle

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

type testRawConsumer struct{}

func (c *testRawConsumer) Consume() (<-chan *events.Envelope, <-chan error) {
	eventCh, errCh := make(chan *events.Envelope), make(chan error)
	return eventCh, errCh
}

func (c *testRawConsumer) Close() error {
	return nil
}

func TestConsumer_implement(t *testing.T) {
	var _ Consumer = &consumer{}
}

func TestRawConsumer_implement(t *testing.T) {
	// Test rawConsumer implements consumer
	var _ rawConsumer = &rawDefaultConsumer{}
}

func TestRawConsumer_consume(t *testing.T) {
	t.Parallel()

	// inputCh is used to send message from test web socket server
	inputCh := make(chan []byte)

	// authToken is valid auth token used for authorizing web socket connection
	authToken := "n98ubNOIUog9gOPUbvqiur"

	// Setup web socket server
	ts := NewDopplerServer(t, inputCh, authToken)
	defer ts.Close()

	consumer := &rawDefaultConsumer{
		dopplerAddr:    strings.Replace(ts.URL, "http:", "ws:", 1),
		token:          authToken,
		subscriptionID: "test-go-nozzle-A",
		insecure:       true,
		logger:         log.New(ioutil.Discard, "", log.LstdFlags),
	}
	eventCh, _ := consumer.Consume()

	// Create test message send from web socket.
	// It will be encoded to protocol buffer.
	timestamp := time.Now().UnixNano()
	message := "Hello from fake loggregator"

	eventBytes, err := NewEvent(message, timestamp)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Send event to websocket
	inputCh <- eventBytes

	// Receive event from websocket(via raw consumer)
	event := <-eventCh
	got := string(event.GetLogMessage().Message)
	if got != message {
		t.Fatalf("expect %q to be eq %q", got, message)
	}
}

func TestRawConsumerClose_no_connection(t *testing.T) {
	consumer := &rawDefaultConsumer{
		logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}
	err := consumer.Close()
	if err == nil {
		t.Fatalf("expects to be failed")
	}
}

func TestRawConsumer_validate(t *testing.T) {
	tests := []struct {
		in      *rawDefaultConsumer
		success bool
	}{
		{
			in: &rawDefaultConsumer{
				dopplerAddr:    "wss://doppler.cloudfoundry.com",
				token:          "POrr7uofS1TOqaGCpH0skk=",
				subscriptionID: "go-nozzle-A",
			},
			success: true,
		},

		{
			in: &rawDefaultConsumer{
				dopplerAddr:    "wss://doppler.cloudfoundry.com",
				subscriptionID: "go-nozzle-A",
			},
			success: false,
		},

		{
			in:      &rawDefaultConsumer{},
			success: false,
		},
	}

	for i, tt := range tests {
		err := tt.in.validate()
		if tt.success {
			if err == nil {
				// ok
				continue
			}
			t.Fatalf("#%d expects '%v' to be nil", i, err)
		}

		if !tt.success && err != nil {
			// ok
			continue
		}

		t.Errorf("#%d expects err not to be nil", i)
	}
}
