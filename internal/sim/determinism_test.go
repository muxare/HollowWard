package sim

import (
	"testing"

	"hollowward/internal/content"
)

// script plays a fixed, state-blind list of orders for a number of days and
// returns every command it sent. Commands the rules reject are left out.
func script(t testing.TB, seed uint64, days int) (State, []Command, []Event) {
	t.Helper()
	var cmds []Command
	var events []Event
	s := State{}
	send := func(c Command) {
		n, ev, err := Apply(s, c)
		if err != nil {
			return
		}
		s, cmds, events = n, append(cmds, c), append(events, ev...)
	}
	send(NewTenure(seed))
	for day := 0; day < days; day++ {
		for st := 0; st < content.NumStewards; st++ {
			send(AssignSteward(st, content.DistrictID((day+2*st)%int(content.NumDistricts)), content.NoNeed, -1))
		}
		send(ResolveEvent(day % 2))
		send(ResolveEvent(1 - day%2)) // if the first was too dear
		send(IssueDecree(content.DecreeID(day%int(content.NumDecrees)), content.DistrictID(day%int(content.NumDistricts))))
		send(EndDay())
		send(StayHome())
	}
	return s, cmds, events
}

// The same seed and command list must give an identical state and event log.
func TestDeterminism(t *testing.T) {
	for _, seed := range []uint64{1, 20261126, 0xdeadbeef} {
		final, cmds, events := script(t, seed, content.T.Nights)
		// Replay the recorded commands from nothing.
		var replayed []Event
		s := State{}
		for _, c := range cmds {
			var ev []Event
			var err error
			s, ev, err = Apply(s, c)
			if err != nil {
				t.Fatalf("seed %d: replay rejected %s: %v", seed, c.Kind, err)
			}
			replayed = append(replayed, ev...)
		}
		if encode(t, s) != encode(t, final) {
			t.Errorf("seed %d: replay ended in a different state", seed)
		}
		if encode(t, replayed) != encode(t, events) {
			t.Errorf("seed %d: replay produced a different event log", seed)
		}
		if len(events) < 100 {
			t.Errorf("seed %d: only %d events in a tenure; the script is not exercising the rules", seed, len(events))
		}
	}
}
