package main

import (
	"bytes"
	"strings"
	"testing"
)

// The harness runs tenures in parallel goroutines. Under -race this test
// proves they share no state.
func TestRunInParallel(t *testing.T) {
	var out bytes.Buffer
	if _, err := run([]string{"-n", "24", "-workers", "8"}, &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"random", "repair-worst", "farm", "wins", "nights", "Market", "Tenements", "Targets:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("report lacks %q:\n%s", want, out.String())
		}
	}
	// One worker or eight, the numbers are the same.
	var serial bytes.Buffer
	if _, err := run([]string{"-n", "24", "-workers", "1"}, &serial); err != nil || serial.String() != out.String() {
		t.Errorf("the report depends on the number of workers (%v)", err)
	}
}

func TestOnePolicyAndBadFlags(t *testing.T) {
	var out bytes.Buffer
	if ok, err := run([]string{"-n", "4", "-policy", "farm"}, &out); err != nil || !ok || strings.Contains(out.String(), "repair-worst ") {
		t.Errorf("single policy run: ok=%v err=%v", ok, err)
	}
	if _, err := run([]string{"-policy", "pacifist"}, &out); err == nil {
		t.Error("an unknown policy must be an error")
	}
	if _, err := run([]string{"-n", "2", "-policy", "farm", "-check"}, &out); err == nil {
		t.Error("-check with one policy must be an error")
	}
}

func TestTargets(t *testing.T) {
	tests := []struct {
		name                 string
		random, repair, farm float64
		want                 [3]bool
	}{
		{"all met", 4, 40, 36, [3]bool{true, true, true}},
		{"random too good", 10, 40, 36, [3]bool{false, true, true}},
		{"repair too easy", 4, 51, 45, [3]bool{true, false, true}},
		{"repair too hard", 4, 29.9, 29, [3]bool{true, false, true}},
		{"farm not viable", 4, 40, 31.9, [3]bool{true, true, false}},
		{"farm exactly at the line", 0, 50, 40, [3]bool{true, true, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := targets(tt.random, tt.repair, tt.farm)
			for i := range got {
				if got[i].met != tt.want[i] {
					t.Errorf("target %d: met=%v, want %v (%s)", i, got[i].met, tt.want[i], got[i].text)
				}
			}
		})
	}
}
