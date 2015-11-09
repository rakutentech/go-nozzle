package nozzle

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaultTokenFetcher_implement(t *testing.T) {
	var _ TokenFetcher = &defaultTokenFetcher{}
}

func TestDefaultTokenFetcher_fetch(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validRequest(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authValue := "Basic " + base64.StdEncoding.EncodeToString([]byte("gonozzle:passw0rd"))
		if authValue != r.Header.Get("Authorization") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData := []byte(`
{
    "access_token":"np9q34bcanBIUI98b9q3vnaoirv",
    "token_type":"bearer",
    "expires_in":599,
    "scope":"cloud_controller.write doppler.firehose",
    "jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
}
`)
		w.Write(jsonData)
		return
	}))
	defer ts.Close()

	fetcher := defaultTokenFetcher{
		uaaAddr:  ts.URL,
		username: "gonozzle",
		password: "passw0rd",
		logger:   log.New(ioutil.Discard, "", log.LstdFlags),
	}

	token, err := fetcher.Fetch()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expect := "bearer np9q34bcanBIUI98b9q3vnaoirv"
	if token != expect {
		t.Fatalf("expect %q to be eq %q", token, expect)
	}

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

// Validate requests of uaa-go.
// This logic comes from https://github.com/cloudfoundry-incubator/uaago/blob/master/client_test.go
func validRequest(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}

	if r.URL.Path != "/oauth/token" {
		return false
	}

	if r.Header.Get("content-type") != "application/x-www-form-urlencoded" {
		return false
	}

	if err := r.ParseForm(); err != nil {
		return false
	}

	if len(r.PostForm.Get("client_id")) == 0 {
		return false
	}

	if len(r.PostForm.Get("grant_type")) == 0 {
		return false
	}

	return true
}
