package nozzle

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestDefaultSlowDetector_implement(t *testing.T) {
	var _ SlowDetector = &defaultSlowDetector{}
}

func TestDefaultSlowDetectorClose(t *testing.T) {
	detector := &defaultSlowDetector{
		logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}
	if err := detector.Close(); err == nil {
		t.Fatalf("expects to be failed")
	}
}
