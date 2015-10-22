package nozzle

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestConsumer_implement(t *testing.T) {
	var _ Consumer = &consumer{}
}

func TestRawConsumer_implement(t *testing.T) {
	// Test rawConsumer implements consumer
	var _ RawConsumer = &rawConsumer{}
}

func TestRawConsumerClose_no_connection(t *testing.T) {
	consumer := &rawConsumer{
		logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}
	err := consumer.Close()
	if err == nil {
		t.Fatalf("expects to be failed")
	}
}

func TestRawConsumer_validate(t *testing.T) {
	tests := []struct {
		in      *rawConsumer
		success bool
	}{
		{
			in: &rawConsumer{
				dopplerAddr:    "wss://doppler.cloudfoundry.com",
				token:          "POrr7uofS1TOqaGCpH0skk=",
				subscriptionID: "go-nozzle-A",
			},
			success: true,
		},

		{
			in: &rawConsumer{
				dopplerAddr:    "wss://doppler.cloudfoundry.com",
				subscriptionID: "go-nozzle-A",
			},
			success: false,
		},

		{
			in:      &rawConsumer{},
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
