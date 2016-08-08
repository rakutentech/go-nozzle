package nozzle

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
)

// SlowDetectCh is channel used to send `slowConsumerAlert` event.
type slowDetectCh chan error

type noaaEventsCh <-chan *events.Envelope

// SlowDetector defines the interface for detecting `slowConsumerAlert`
// event. By default, defaultSlowDetetor is used. It implements same detection
// logic as https://github.com/cloudfoundry-incubator/datadog-firehose-nozzle.
type slowDetector interface {
	// Detect detects `slowConsumerAlert`. It works as pipe.
	// It receives events from upstream (RawConsumer) and inspects that events
	// and pass it to to downstream without modification.
	//
	// It returns SlowDetectCh and notify `slowConsumerAlert` there.
	Detect(context.Context, noaaEventsCh, <-chan error) (noaaEventsCh, <-chan error, slowDetectCh)
}

// defaultSlowDetector implements SlowDetector interface
type defaultSlowDetector struct {
	cancelFunc context.CancelFunc
	logger     *log.Logger
}

// Detect start to detect `slowConsumerAlert` event.
func (sd *defaultSlowDetector) Detect(ctx context.Context, eventCh noaaEventsCh, errCh <-chan error) (noaaEventsCh, <-chan error, slowDetectCh) {

	if ctx == nil {
		panic("nil context")
	}

	sd.logger.Println("[INFO] Start detecting slowConsumerAlert event")

	// Create new channel to pass producer
	eventCh_ := make(chan *events.Envelope)
	errCh_ := make(chan error)

	// deteCh is used to send `slowConsumerAlert` event
	detectCh := make(slowDetectCh)

	// Detect from from trafficcontroller event messages
	go func() {
		defer close(eventCh_)
		for {
			select {
			case event := <-eventCh:
				// Check nozzle can catch up firehose outputs speed.
				if isTruncated(event) {
					detectCh <- fmt.Errorf(
						"doppler dropped messages from its queue because nozzle is slow")
				}

				eventCh_ <- event

			case <-ctx.Done():
				// Send errCh_ that context is closed
				sd.logger.Println("[INFO] Canceled parent context: closing event channel")

				// close downstream eventCh
				return
			}
		}
	}()

	// Detect from websocket errors
	go func() {
		defer close(errCh_)
		for {
			select {
			case err := <-errCh:

				switch t := err.(type) {
				case *websocket.CloseError:
					if t.Code == websocket.ClosePolicyViolation {
						// ClosePolicyViolation (1008)
						// indicates that an endpoint is terminating the connection
						// because it has received a message that violates its policy.
						//
						// This is a generic status code that can be returned when there is no
						// other more suitable status code (e.g., 1003 or 1009) or if there
						// is a need to hide specific details about the policy.
						//
						// http://tools.ietf.org/html/rfc6455#section-11.7
						msg := "websocket terminates the connection because connection is too slow"
						detectCh <- fmt.Errorf(msg)
					}
				}
				errCh_ <- err

			case <-ctx.Done():
				// Send errCh_ that context is closed
				sd.logger.Println("[INFO] Canceled parent context: closing error channel")

				// close downstream errCh and eventCh
				return
			}
		}
	}()

	go func() {
		// Wait cancel signal and close detectCh
		<-ctx.Done()
		close(detectCh)
	}()

	return eventCh_, errCh_, detectCh
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
