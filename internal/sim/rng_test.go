package sim

import (
	"bytes"
	"testing"

	"hollowward/internal/content"
)

func TestStreamIsDeterministic(t *testing.T) {
	a, b, other := NewStream(42, 1), NewStream(42, 1), NewStream(42, 2)
	same, differ := true, false
	for range 50 {
		x, y, z := a.IntN(1000), b.IntN(1000), other.IntN(1000)
		same = same && x == y
		differ = differ || x != z
	}
	if !same {
		t.Error("two streams with one seed disagreed")
	}
	if !differ {
		t.Error("two streams of one tenure produced the same numbers")
	}
}

func TestStreamHelpers(t *testing.T) {
	s := NewStream(1, 1)
	for range 200 {
		if v := s.Between(3, 7); v < 3 || v > 7 {
			t.Fatalf("Between(3,7) = %d", v)
		}
	}
	if v := s.Between(5, 5); v != 5 {
		t.Errorf("Between(5,5) = %d", v)
	}
	p := s.Perm(12)
	seen := map[int]bool{}
	for _, v := range p {
		seen[v] = true
	}
	if len(p) != 12 || len(seen) != 12 {
		t.Errorf("Perm(12) = %v", p)
	}
	if s.Chance(0) || !s.Chance(100) {
		t.Error("Chance(0) must fail and Chance(100) must pass")
	}
}

// Drawing from one subsystem's stream must leave every other stream alone, so
// a change to one subsystem cannot reshuffle the others.
func TestStreamsAreIndependent(t *testing.T) {
	s := start(t, 77)
	fresh := newStreams(77)
	// Setup has been used; nothing else but Events has.
	for name, pair := range map[string][2]Stream{
		"dusk":   {s.RNG.Dusk, fresh.Dusk},
		"gen":    {s.RNG.Gen, fresh.Gen},
		"combat": {s.RNG.Combat, fresh.Combat},
		"fate":   {s.RNG.Fate, fresh.Fate},
	} {
		if !bytes.Equal(pair[0].PCG, pair[1].PCG) {
			t.Errorf("building the city advanced the %s stream", name)
		}
	}
	if bytes.Equal(s.RNG.Setup.PCG, fresh.Setup.PCG) {
		t.Error("building the city did not use the setup stream")
	}

	// Ten days of events must not move the setup stream either.
	setup := append([]byte(nil), s.RNG.Setup.PCG...)
	events := append([]byte(nil), s.RNG.Events.PCG...)
	for range 10 {
		s = answer(t, s)
		s, _ = do(t, s, EndDay())
	}
	if !bytes.Equal(s.RNG.Setup.PCG, setup) {
		t.Error("playing days advanced the setup stream")
	}
	if bytes.Equal(s.RNG.Events.PCG, events) {
		t.Error("ten days of cards did not advance the events stream")
	}

	// The same cards come up whatever the warden does with stewards and decrees.
	a, b := start(t, 77), start(t, 77)
	b, _ = do(t, b, AssignSteward(0, content.Market, content.NoNeed, -1))
	for range 15 {
		if a.Card.Card != b.Card.Card || a.Card.District != b.Card.District {
			t.Fatal("the event deck depended on steward orders")
		}
		a, b = answer(t, a), answer(t, b)
		a, _ = do(t, a, EndDay())
		b, _ = do(t, b, EndDay())
	}
}

func TestCloneSharesNoStream(t *testing.T) {
	s := start(t, 5)
	c := s.Clone()
	c.RNG.Events.IntN(10)
	c.Citizens[0].Scars = append(c.Citizens[0].Scars, content.Limp)
	c.Districts[0].Buildings[0].Condition = -99
	c.Card.Resolved = true
	c.Deck[0] = -1
	fresh := start(t, 5)
	if encode(t, s) != encode(t, fresh) {
		t.Error("changing a clone changed the original")
	}
}
