package nozzle

import (
	"context"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

type testRawConsumer struct{}

func (c *testRawConsumer) Consume(ctx context.Context) (noaaEventsCh, <-chan error) {
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming
	eventCh, _ := consumer.Consume(ctx)

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

func TestRawConsumer_consume_cancel(t *testing.T) {
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
		subscriptionID: "test-go-nozzle-B",
		insecure:       true,
		logger:         log.New(ioutil.Discard, "", log.LstdFlags),
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start consuming
	eventCh, _ := consumer.Consume(ctx)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		_, chOpen := <-eventCh
		if !chOpen {
			wg.Done()
		}
	}()

	// Wait until all goroutines are closed
	doneCh := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		doneCh <- struct{}{}
	}()

	cancel()
	select {
	case <-doneCh:
		// success
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for 2 channels")
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
