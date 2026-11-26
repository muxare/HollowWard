package sim

import "math/rand/v2"

// Stream is one seeded PCG random stream. Its whole state lives here, in the
// game state, so a saved game resumes the same future. Each subsystem draws
// from its own stream: a change in how combat rolls dice cannot reshuffle the
// event deck.
type Stream struct {
	PCG []byte `json:"pcg"` // rand.PCG's binary form
}

// NewStream seeds a stream. The second word tells the streams of one tenure
// apart.
func NewStream(seed, stream uint64) Stream {
	b, err := rand.NewPCG(seed, stream).MarshalBinary()
	if err != nil {
		panic("sim: PCG marshal: " + err.Error()) // cannot happen
	}
	return Stream{PCG: b}
}

func (s Stream) clone() Stream { return Stream{PCG: append([]byte(nil), s.PCG...)} }

// draw runs f with the stream's generator and stores the advanced state.
func (s *Stream) draw(f func(r *rand.Rand)) {
	var p rand.PCG
	if err := p.UnmarshalBinary(s.PCG); err != nil {
		panic("sim: corrupt random stream: " + err.Error())
	}
	f(rand.New(&p))
	b, err := p.MarshalBinary()
	if err != nil {
		panic("sim: PCG marshal: " + err.Error())
	}
	s.PCG = b
}

// IntN returns a number in [0, n).
func (s *Stream) IntN(n int) int {
	var v int
	s.draw(func(r *rand.Rand) { v = r.IntN(n) })
	return v
}

// Between returns a number in [lo, hi].
func (s *Stream) Between(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + s.IntN(hi-lo+1)
}

// Chance reports true pct times in a hundred.
func (s *Stream) Chance(pct int) bool { return s.IntN(100) < pct }

// Perm returns a shuffled list of 0..n-1.
func (s *Stream) Perm(n int) []int {
	var v []int
	s.draw(func(r *rand.Rand) { v = r.Perm(n) })
	return v
}

// Streams are a tenure's random streams, one per subsystem.
type Streams struct {
	Setup  Stream `json:"setup"`  // the city shuffle at the start of a tenure
	Events Stream `json:"events"` // the event deck and its targets
	Dusk   Stream `json:"dusk"`   // tie-breaks between equally neglected districts
	Gen    Stream `json:"gen"`    // hollow generation
	Combat Stream `json:"combat"` // combat rolls
	Fate   Stream `json:"fate"`   // deaths, scars and loot at night's end
}

func newStreams(seed uint64) Streams {
	return Streams{
		Setup:  NewStream(seed, 1),
		Events: NewStream(seed, 2),
		Dusk:   NewStream(seed, 3),
		Gen:    NewStream(seed, 4),
		Combat: NewStream(seed, 5),
		Fate:   NewStream(seed, 6),
	}
}

func (s Streams) clone() Streams {
	return Streams{
		Setup: s.Setup.clone(), Events: s.Events.clone(), Dusk: s.Dusk.clone(),
		Gen: s.Gen.clone(), Combat: s.Combat.clone(), Fate: s.Fate.clone(),
	}
}
