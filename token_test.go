package nozzle

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"
)

func TestDefaultTokenFetcher_implement(t *testing.T) {
	var _ TokenFetcher = &defaultTokenFetcher{}
}

func TestDefaultTokenFetcher_timeout(t *testing.T) {
	t.Parallel()
	fetcher := &defaultTokenFetcher{
		uaaAddr:  "https://localhost",
		username: "admin",
		password: "nipr8qhbp89pq",
		logger:   log.New(ioutil.Discard, "", log.LstdFlags),

		// Set very very very short timeout time
		timeout: 1 * time.Microsecond,
	}

	// Execute fetcher
	_, err := fetcher.Fetch()

	expect := "timeout"
	if !strings.Contains(err.Error(), expect) {
		t.Fatalf("expects error message %q to contain %q", err.Error(), expect)
	}

}

func TestDefaultTokenFetcher_validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      *defaultTokenFetcher
		success bool
	}{
		{
			in: &defaultTokenFetcher{
				uaaAddr:  "https://uaa.cloudfoundry.net",
				username: "admin",
				password: "npi4Cgupn",
			},
			success: true,
		},

		{
			in: &defaultTokenFetcher{
				uaaAddr: "https://uaa.cloudfoundry.net",
			},
			success: false,
		},

		{
			in:      &defaultTokenFetcher{},
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
