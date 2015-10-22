package nozzle

import "testing"

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
