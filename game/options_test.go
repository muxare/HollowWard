package game

import "testing"

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want options
	}{
		{"none", nil, options{}},
		{"autoplay", []string{"--autoplay"}, options{autoplay: true}},
		{"autoplay fail", []string{"--autoplay-fail"}, options{autoplay: true, fail: true}},
		{"screenshot", []string{"--screenshot=/tmp/x.png"}, options{screenshot: "/tmp/x.png"}},
		{"unknown ignored", []string{"--nope", "--autoplay"}, options{autoplay: true}},
		{"shots and seed", []string{"--autoplay", "--shots=/tmp/s", "--seed=42"}, options{autoplay: true, shots: "/tmp/s", seed: 42}},
		{"a bad seed is ignored", []string{"--seed=many"}, options{}},
		{"a whole tenure", []string{"--autoplay-full"}, options{autoplay: true, full: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseOptions(tt.args); got != tt.want {
				t.Errorf("parseOptions(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}
