package nozzle

import "github.com/cloudfoundry/sonde-go/events"

type Producer interface {
	Produce(chan *events.Envelope, chan error, SlowDetectCh)
}
