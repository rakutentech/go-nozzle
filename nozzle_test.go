package nozzle

import (
	"strings"
	"testing"
)

func TestDefaultConsumer(t *testing.T) {
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
