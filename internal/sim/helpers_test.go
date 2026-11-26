package sim

import (
	"encoding/json"
	"errors"
	"testing"

	"hollowward/internal/content"
)

// start begins a tenure or fails the test.
func start(t testing.TB, seed uint64) State {
	t.Helper()
	s, _, err := Apply(State{}, NewTenure(seed))
	if err != nil {
		t.Fatalf("new tenure: %v", err)
	}
	return s
}

// do applies a command or fails the test.
func do(t testing.TB, s State, c Command) (State, []Event) {
	t.Helper()
	n, ev, err := Apply(s, c)
	if err != nil {
		t.Fatalf("%s: %v", c.Kind, err)
	}
	return n, ev
}

// mustFail applies a command that has to be rejected with the wanted error,
// and checks that a rejection changes nothing.
func mustFail(t testing.TB, s State, c Command, want error) {
	t.Helper()
	before := encode(t, s)
	n, ev, err := Apply(s, c)
	if !errors.Is(err, want) {
		t.Fatalf("%s: error %v, want %v", c.Kind, err, want)
	}
	if ev != nil || encode(t, n) != before || encode(t, s) != before {
		t.Fatalf("%s: a rejected command changed something", c.Kind)
	}
}

func encode(t testing.TB, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return string(b)
}

// flat sets every need of every district to v and every building to cond.
func flat(s *State, v, cond int) {
	for d := range s.Districts {
		for n := range s.Districts[d].Needs {
			s.Districts[d].Needs[n] = v
		}
		for b := range s.Districts[d].Buildings {
			s.Districts[d].Buildings[b].Condition = cond
		}
	}
}

// answer resolves today's card with its first free choice.
func answer(t testing.TB, s State) State {
	t.Helper()
	if s.Card == nil || s.Card.Resolved {
		return s
	}
	for i, ch := range content.EventCards[s.Card.Card].Choices {
		if ch.Cost.IsZero() {
			s, _ = do(t, s, ResolveEvent(i))
			return s
		}
	}
	t.Fatalf("card %d has no free choice", s.Card.Card)
	return s
}
