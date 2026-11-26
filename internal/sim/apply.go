package sim

import "fmt"

// Apply is the only way state changes. It returns the new state and the
// events the command caused. The state passed in is never modified; when the
// command is rejected the same state comes back with an error.
func Apply(s State, c Command) (State, []Event, error) {
	l := &log{day: s.Day}
	if c.Kind == CmdNewTenure {
		if s.Phase != PhaseNone && s.Phase != PhaseOver {
			return s, nil, fmt.Errorf("%w: a tenure is under way", ErrWrongPhase)
		}
		return newTenure(c.Seed, l), l.events, nil
	}
	n := s.Clone()
	if err := n.apply(c, l); err != nil {
		return s, nil, fmt.Errorf("%s: %w", c.Kind, err)
	}
	return n, l.events, nil
}

func (s *State) apply(c Command, l *log) error {
	switch c.Kind {
	case CmdAssignSteward, CmdRecallSteward, CmdIssueDecree, CmdResolveEvent, CmdEndDay:
		if s.Phase != PhaseDay {
			return ErrWrongPhase
		}
	}
	switch c.Kind {
	case CmdSendParty, CmdStayHome:
		if s.Phase != PhaseDusk {
			return ErrWrongPhase
		}
	case CmdAdvance, CmdGoHome, CmdAct:
		if s.Phase != PhaseNight || s.Expedition == nil {
			return ErrWrongPhase
		}
	}
	switch c.Kind {
	case CmdSendParty:
		return s.sendParty(c, l)
	case CmdStayHome:
		return s.stayHome(l)
	case CmdAdvance:
		return s.advance(c, l)
	case CmdGoHome:
		return s.goHome(l)
	case CmdAct:
		return s.act(c, l)
	case CmdAssignSteward:
		return s.assignSteward(c)
	case CmdRecallSteward:
		return s.recallSteward(c)
	case CmdIssueDecree:
		return s.issueDecree(c, l)
	case CmdResolveEvent:
		return s.resolveEvent(c, l)
	case CmdEndDay:
		return s.endDay(l)
	}
	return ErrUnknown
}

// The pieces of the night that the full tenure (M4) fills in.

// duskClock is where the Dreamer rises at dusk.
func (s *State) duskClock(_ *log) {}

// dawn is where the tenure can end. It reports whether it did.
func (s *State) dawn(_ *log) bool { return false }

// searchHaven is where a safe room gives up what it hides.
func (s *State) searchHaven(_ *Room, _ *log) {}
