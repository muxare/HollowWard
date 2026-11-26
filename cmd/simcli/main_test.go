package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
		miss []string
	}{
		{"by seed", []string{"generate", "-seed", "7"}, []string{"THE HOLLOW OF THE", "Neglect", "heart", "why:"}, nil},
		{"the spec's fixture", []string{"generate", "-seed", "1", "-district", "market", "-needs", "10,80,80", "-conds", "50,50,50"},
			[]string{"THE HOLLOW OF THE MARKET", "Market Supply was 10", "The Starved Grain Factor"},
			[]string{"Ash-Mask", "Brand-Bearer", "Many-Hands", "Blackcap", "Mould-Hound", "Sump", "hazard:", "safe room"}},
		{"buildings", []string{"generate", "-district", "docks", "-needs", "60,60,60", "-conds", "5,90,50"},
			[]string{"The Flooded Warehouse", "the Warehouse was at 5", "The Crane Yard", "the Crane was at 90", "empty:"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := run(tt.args, &out); err != nil {
				t.Fatal(err)
			}
			for _, w := range tt.want {
				if !strings.Contains(out.String(), w) {
					t.Errorf("output lacks %q:\n%s", w, out.String())
				}
			}
			for _, m := range tt.miss {
				if strings.Contains(out.String(), m) {
					t.Errorf("output must not contain %q:\n%s", m, out.String())
				}
			}
			var again bytes.Buffer
			if err := run(tt.args, &again); err != nil || again.String() != out.String() {
				t.Error("the same arguments printed a different hollow")
			}
		})
	}
	for _, bad := range [][]string{{}, {"dance"}, {"generate", "-district", "moon"}, {"generate", "-needs", "1,2"}, {"generate", "-needs", "a,b,c"}} {
		if err := run(bad, &bytes.Buffer{}); err == nil {
			t.Errorf("simcli %v: want an error", bad)
		}
	}
}
