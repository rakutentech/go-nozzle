package nozzle

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
)

var (
	// Truncated target values
	TR_Origin    = "doppler"
	TR_EventType = events.Envelope_CounterEvent
	TR_EventName = "TruncatingBuffer.DroppedMessages"
)

func TestDefaultSlowDetector_implement(t *testing.T) {
	var _ slowDetector = &defaultSlowDetector{}
}

func TestDefaultSlowDetectorClose(t *testing.T) {
	detector := &defaultSlowDetector{
		logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}
	if err := detector.Stop(); err == nil {
		t.Fatalf("expects to be failed")
	}
}

func TestDefaultDetect_eventCh(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  *events.Envelope
		Expect bool
	}{
		{
			Input: &events.Envelope{
				Origin:    &TR_Origin,
				EventType: &TR_EventType,
				CounterEvent: &events.CounterEvent{
					Name: &TR_EventName,
				},
			},
			Expect: true,
		},

		{
			Input:  &events.Envelope{},
			Expect: false,
		},
	}

	testDetector := &defaultSlowDetector{
		logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}

	eventCh := make(chan *events.Envelope)
	errCh := make(chan error)
	_, _, detectCh := testDetector.Detect(context.Background(), eventCh, errCh)

	for _, tc := range cases {
		// Send the events
		go func() {
			eventCh <- tc.Input
		}()

		select {
		case <-detectCh:
			if !tc.Expect {
				t.Fatalf("expect not to be detected")
			}
		case <-time.After(1 * time.Second):
			if tc.Expect {
				t.Fatalf("expect to be detected")
			}
		}
	}
}

func TestDefaultDetect_errCh(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  error
		Expect bool
	}{
		{
			Input: &websocket.CloseError{
				Code: websocket.ClosePolicyViolation,
			},
			Expect: true,
		},

		{
			Input:  errors.New(""),
			Expect: false,
		},
	}

	testDetector := &defaultSlowDetector{
		logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}

	eventCh := make(chan *events.Envelope)
	errCh := make(chan error)
	_, _, detectCh := testDetector.Detect(context.Background(), eventCh, errCh)

	for _, tc := range cases {
		// Send the events
		go func() {
			errCh <- tc.Input
		}()

		select {
		case <-detectCh:
			if !tc.Expect {
				t.Fatalf("expect not to be detected")
			}
		case <-time.After(1 * time.Second):
			if tc.Expect {
				t.Fatalf("expect to be detected")
			}
		}
	}
}

func TestDefaultDetect_context(t *testing.T) {
	t.Parallel()

	testDetector := &defaultSlowDetector{
		logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}

	eventCh := make(chan *events.Envelope)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	eventCh_, errCh_, detectCh := testDetector.Detect(ctx, eventCh, errCh)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		_, chOpen := <-eventCh_
		if !chOpen {
			wg.Done()
		}
	}()

	wg.Add(1)
	go func() {
		_, chOpen := <-detectCh
		if !chOpen {
			wg.Done()
		}
	}()

	wg.Add(1)
	go func() {
		for {
			_, chOpen := <-errCh_
			if !chOpen {
				wg.Done()
				return
			}
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
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for 3 channels")
	}
}

func TestIsTruncated(t *testing.T) {
	cases := []struct {
		Input  *events.Envelope
		Expect bool
	}{
		{
			Input: &events.Envelope{
				Origin:    &TR_Origin,
				EventType: &TR_EventType,
				CounterEvent: &events.CounterEvent{
					Name: &TR_EventName,
				},
			},
			Expect: true,
		},

		{
			Input:  &events.Envelope{},
			Expect: false,
		},
	}

	for i, tc := range cases {
		output := isTruncated(tc.Input)
		if output != tc.Expect {
			t.Fatalf("#%d expects %v to be eq %v", i, output, tc.Expect)
		}
	}
}
