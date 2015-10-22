package nozzle

import "github.com/cloudfoundry/sonde-go/events"

type FakeProducer struct{}

func (p *FakeProducer) Produce(_ chan *events.Envelope, _ chan error, _ SlowDetectCh) {}
