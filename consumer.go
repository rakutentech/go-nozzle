package nozzle

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	noaaConsumer "github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

// Consumer defines the interface of consumer it receives
// upstream firehose events and slowConsumerAlerts events and errors.
type Consumer interface {
	// Events returns the read channel for the events that consumed by
	// rawConsumer(by default Noaa).
	Events() <-chan *events.Envelope

	// Detects returns the read channel that is notified slowConsumerAlerts
	// handled by SlowDetector.
	Detects() <-chan error

	// Error returns the read channel of erros that occured during consuming.
	Errors() <-chan error

	// Start starts consuming upstream events by RawConsumer and stop SlowDetector.
	// If any, returns error.
	Start() error

	// Close stop consuming upstream events by RawConsumer and stop SlowDetector.
	// If any, returns error.
	Close() error
}

type consumer struct {
	rawConsumer rawConsumer
	cancelFunc  context.CancelFunc
	logger      *log.Logger

	eventCh  <-chan *events.Envelope
	errCh    <-chan error
	detectCh <-chan error
}

// Events returns the read channel for the events that consumed by rawConsumer
func (c *consumer) Events() <-chan *events.Envelope {
	return c.eventCh
}

// Detects returns the read channel that is notified slowConsumerAlerts
func (c *consumer) Detects() <-chan error {
	return c.detectCh
}

// Error returns the read channel of erros that occured during consuming.
func (c *consumer) Errors() <-chan error {
	return c.errCh
}

// Start starts consuming & slowDetector
func (c *consumer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel
	return c.StartContext(ctx)
}

func (c *consumer) StartContext(ctx context.Context) error {

	if ctx == nil {
		panic("nil context")
	}

	// Start consuming events from firehose.
	eventsCh, errCh := c.rawConsumer.ConsumeContext(ctx)

	// Construct default slowDetector
	sd := &defaultSlowDetector{
		logger: c.logger,
	}

	// Start reading events from firehose and detect `slowConsumerAlert`.
	// The detection is notified by detectCh.
	c.eventCh, c.errCh, c.detectCh = sd.DetectContext(ctx, eventsCh, errCh)

	// In current implementation no errors are happened.
	//
	// This is for preventing interfance change in future when
	// we need to put some other function here and it returns error.
	return nil
}

// Close closes connection with firehose and stop slowDetector.
func (c *consumer) Close() error {
	if c.cancelFunc == nil {
		return fmt.Errorf("")
	}

	c.cancelFunc()
	return nil
}

// rawConsumer defines the interface for consuming events from doppler firehose.
// The events pulled by RawConsumer pass to slowDetector and check slowDetector.
//
// By default, it uses https://github.com/cloudfoundry/noaa
type rawConsumer interface {
	// Consume starts cosuming firehose events. It must return 2 channel.
	// The one is for sending the events from firehose
	// and the other is for error occured while consuming.
	// These channels are used donwstream process (SlowConsumer).
	Consume() (<-chan *events.Envelope, <-chan error)

	// ConsumeContext start consuming firehose events using the given context
	ConsumeContext(context.Context) (noaaEventsCh, <-chan error)

	// Close closes connection with firehose. If any, returns error.
	Close() error
}

type rawDefaultConsumer struct {
	dopplerAddr    string
	token          string
	subscriptionID string
	insecure       bool
	debugPrinter   noaaConsumer.DebugPrinter

	logger *log.Logger

	cancelFunc context.CancelFunc
}

func (c *rawDefaultConsumer) ConsumeContext(ctx context.Context) (noaaEventsCh, <-chan error) {
	if ctx == nil {
		panic("nil context")
	}

	c.logger.Printf(
		"[INFO] Start consuming firehose events from Doppler (%s) with subscription ID %q",
		c.dopplerAddr, c.subscriptionID)

	// Setup Noaa Consumer
	tlsConfig := tls.Config{
		InsecureSkipVerify: c.insecure,
	}
	nc := noaaConsumer.New(c.dopplerAddr, &tlsConfig, nil)

	if c.debugPrinter != nil {
		nc.SetDebugPrinter(c.debugPrinter)
	}

	// Start connection
	eventChan, errChan := nc.Firehose(c.subscriptionID, c.token)

	// Start to watch context is canceled
	go func() {
		<-ctx.Done()
		err := nc.Close()

		// TODO(tcnksm)
		_ = err
	}()

	return eventChan, errChan
}

// Consume consumes firehose events from doppler.
// Retry function is handled in noaa library (It will retry 5 times).
func (c *rawDefaultConsumer) Consume() (<-chan *events.Envelope, <-chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel
	return c.ConsumeContext(ctx)
}

func (c *rawDefaultConsumer) Close() error {
	c.logger.Printf("[INFO] Stop consuming firehose events")
	if c.cancelFunc == nil {
		return fmt.Errorf("cancel function is not given")
	}

	c.cancelFunc()
	return nil
}

// validate validates struct has requirement fields or not
func (c *rawDefaultConsumer) validate() error {
	if c.dopplerAddr == "" {
		return fmt.Errorf("DopplerAddr must not be empty")
	}

	if c.token == "" {
		return fmt.Errorf("Token must not be empty")
	}

	if c.subscriptionID == "" {
		return fmt.Errorf("SubscriptionID must not be empty")
	}

	return nil
}

// newRawConsumer constructs new rawConsumer.
func newRawDefaultConsumer(config *Config) (*rawDefaultConsumer, error) {
	c := &rawDefaultConsumer{
		dopplerAddr:    config.DopplerAddr,
		token:          config.Token,
		subscriptionID: config.SubscriptionID,
		insecure:       config.Insecure,
		debugPrinter:   config.DebugPrinter,
		logger:         config.Logger,
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}
