package nozzle

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultConsumer(t *testing.T) {
	t.Parallel()

	// inputCh is used to send message from test web socket server
	inputCh := make(chan []byte, 1)

	// authToken is valid auth token used for authorizing web socket connection
	authToken := "ncp9q3vbap98r4denpiubg"

	// Setup web socket server
	ds := NewDopplerServer(t, inputCh, authToken)
	defer ds.Close()

	config := &Config{
		DopplerAddr:    strings.Replace(ds.URL, "http:", "ws:", 1),
		Insecure:       true,
		Token:          authToken,
		SubscriptionID: "A",
	}

	consumer, err := NewDefaultConsumer(config)
	if err != nil {
		t.Fatalf("Expect not to err: %s", err)
	}

	// Start consuming.
	if err := consumer.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Create test message send from web socket.
	// It will be encoded to protocol buffer.
	timestamp := time.Now().UnixNano()
	message := "Hello from fake loggregator"

	eventBytes, err := NewEvent(message, timestamp)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Send message from doppler.
	inputCh <- eventBytes

	select {
	case event := <-consumer.Events():
		got := string(event.GetLogMessage().Message)
		if got != message {
			t.Fatalf("expect %q to be eq %q", got, message)
		}
	case err := <-consumer.Errors():
		if err != nil {
			t.Fatalf("err :%s", err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatalf("expect not timeout")
	}
}

func TestDefaultConsumer_withoutStart(t *testing.T) {

	// inputCh is used to send message from test web socket server
	inputCh := make(chan []byte, 1)

	// authToken is valid auth token used for authorizing web socket connection
	authToken := "ncp9q3vbap98r4denpiubg"

	// Setup web socket server
	ds := NewDopplerServer(t, inputCh, authToken)
	defer ds.Close()

	config := &Config{
		DopplerAddr:    strings.Replace(ds.URL, "http:", "ws:", 1),
		Insecure:       true,
		Token:          authToken,
		SubscriptionID: "A",
	}

	// Create consumer but not start consumer
	consumer, err := NewDefaultConsumer(config)
	if err != nil {
		t.Fatalf("Expect not to err: %s", err)
	}

	// Create test message send from web socket.
	// It will be encoded to protocol buffer.
	timestamp := time.Now().UnixNano()
	message := "Hello from fake loggregator"

	eventBytes, err := NewEvent(message, timestamp)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Send message from doppler.
	inputCh <- eventBytes
	received := false
	select {
	case <-consumer.Events():
		received = true
	case <-consumer.Detects():
		received = true
	case <-consumer.Errors():
		received = true
	case <-time.After(100 * time.Millisecond):
	}

	if received {
		t.Fatalf("expect consumer doesn't start")
	}
}

func TestNewDefaultConsumer(t *testing.T) {
	cases := []struct {
		in      *Config
		success bool
		errStr  string
	}{
		{
			in:      &Config{},
			success: false,
			errStr:  "both Token and UaaAddr can not be empty",
		},

		{
			in: &Config{
				Token: "xyz",
			},
			success: false,
			errStr:  "DopplerAddr must not be empty",
		},

		{
			in: &Config{
				Token:       "xyz",
				rawConsumer: &testRawConsumer{},
			},
			success: true,
		},

		{
			in: &Config{
				UaaAddr: "https://uaa.cloudfoundry.net",
			},
			success: false,
			errStr:  "Username must not be empty",
		},

		{
			in: &Config{
				UaaAddr:      "https://uaa.cloudfoundry.net",
				tokenFetcher: &testTokenFetcher{},
			},
			success: false,
			errStr:  "no token found",
		},

		{
			in: &Config{
				UaaAddr: "https://uaa.cloudfoundry.net",
				tokenFetcher: &testTokenFetcher{
					Token: "abc",
				},
			},
			success: false,
			errStr:  "DopplerAddr must not be empty",
		},

		{
			in: &Config{
				UaaAddr: "https://uaa.cloudfoundry.net",
				tokenFetcher: &testTokenFetcher{
					Token: "abc",
				},
				rawConsumer: &testRawConsumer{},
			},
			success: true,
		},
	}

	for i, tc := range cases {
		_, err := NewDefaultConsumer(tc.in)
		if tc.success {
			if err == nil {
				// ok
				continue
			}

			t.Fatalf("#%d expects %q to be nil", i, err)
		}

		if err == nil {
			t.Fatalf("#%d expects to be failed", i)
		}

		if !strings.Contains(err.Error(), tc.errStr) {
			t.Fatalf("#%d expects err message %q to contain %q", i, err.Error(), tc.errStr)
		}
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		in, expect string
	}{
		{
			in:     "nCOB98",
			expect: "**** (masked)",
		},

		{
			in:     "nuVHBvbguP4713tnpuUIU9uI",
			expect: "nuVHBvbguP**** (masked)",
		},
	}

	for i, tt := range tests {
		out := maskString(tt.in)
		if out != tt.expect {
			t.Fatalf("#%d expects %q to be eq %q", i, out, tt.expect)
		}
	}
}
