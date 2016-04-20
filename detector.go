package nozzle

import (
	"fmt"
	"log"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
)

// SlowDetectCh is channel used to send `slowConsumerAlert` event.
type SlowDetectCh chan struct{}

// SlowDetector defines the interface for detecting `slowConsumerAlert`
// event. By default, defaultSlowDetetor is used. It implements same detection
// logic with https://github.com/cloudfoundry-incubator/datadog-firehose-nozzle.
type SlowDetector interface {
	// Detect detects `slowConsumerAlert`. It receives upstream
	// events (RawConsumer) and inspects events it indicate
	// `slowConsumerAlert` and pass events to downstream.
	//
	// The events should be passed without modified.
	// It returns SlowDetectCh and notify there if `slowConsumerAlert` is detected.
	Detect(chan *events.Envelope, chan error) (chan *events.Envelope, chan error, SlowDetectCh)

	// Stop stops slow consumer detection. If any returns error.
	Stop() error
}

// defaultSlowDetector implements SlowDetector interface
type defaultSlowDetector struct {
	doneCh chan struct{}
	logger *log.Logger
}

// Detect start to detect `slowConsumerAlert` event.
func (sd *defaultSlowDetector) Detect(eventCh chan *events.Envelope, errCh chan error) (chan *events.Envelope, chan error, SlowDetectCh) {
	sd.logger.Println("[INFO] Start detecting slowConsumerAlert event")

	// Create new channel to pass producer
	eventCh_ := make(chan *events.Envelope)
	errCh_ := make(chan error)

	// doneCh is used to cancel sending data to
	// downstream process.
	sd.doneCh = make(chan struct{})

	// deteCh is used to send `slowConsumerAlert` event
	detectCh := make(SlowDetectCh)

	// Detect from from trafficcontroller event messages
	go func() {
		defer close(eventCh_)
		for event := range eventCh {
			// Check nozzle can catch up firehose outputs speed.
			if isTruncated(event) {
				detectCh <- struct{}{}
			}

			select {
			case eventCh_ <- event:
			case <-sd.doneCh:
				// After doneCh is closed, sending event to downstream
				// is immediately stopped.
				return
			}

		}
	}()

	// Detect from websocket errors
	go func() {
		defer close(errCh_)
		for err := range errCh {
			// Disconnected because nozzle couldn't keep up.
			// Please try scaling up the nozzle.
			switch t := err.(type) {
			case *websocket.CloseError:
				if t.Code == websocket.ClosePolicyViolation {
					detectCh <- struct{}{}
				}
			}
			select {
			case errCh_ <- err:
			case <-sd.doneCh:
				// After doneCh is closed, sending events to downstream
				// is immediately stopped.
				return
			}

		}
	}()

	return eventCh_, errCh_, detectCh
}

func (sd *defaultSlowDetector) Stop() error {
	sd.logger.Println("[INFO] Stop detecting slowConsumerAlert event")
	if sd.doneCh == nil {
		return fmt.Errorf("slow detector is not running")
	}

	close(sd.doneCh)
	return nil
}

// isTruncated detects message from the Doppler that the nozzle
// could not consume messages as quickly as the firehose was sending them.
func isTruncated(envelope *events.Envelope) bool {
	if envelope.GetEventType() == events.Envelope_CounterEvent &&
		envelope.CounterEvent.GetName() == "TruncatingBuffer.DroppedMessages" &&
		envelope.GetOrigin() == "doppler" {
		return true
	}

	return false
}
