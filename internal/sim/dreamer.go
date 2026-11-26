package sim

import (
	"fmt"

	"hollowward/internal/content"
)

// What the full tenure adds to the log.
const (
	EvDreamerRose    EventKind = "dreamer-rose"    // N: rise in tenths; V: the meter now, in tenths
	EvDreamerSoothed EventKind = "dreamer-soothed" // N: fall in tenths; V: the meter now
	EvDreamLine      EventKind = "dream-line"      // ID: what the Dreamer says tonight
	EvUpgradeBought  EventKind = "upgrade-bought"  // ID: upgrade slug; District and V: the building, for a ward
	EvFragment       EventKind = "fragment"        // N: which of the old warden's notes
	EvTenureEnded    EventKind = "tenure-ended"    // ID: outcome; N: nights survived
)

// duskClock is the Dreamer: it rises every dusk, by the turned district's
// Neglect or a little on a quiet night, and more for every district still
// turned from an earlier night. At its limit the lid breaks.
func (s *State) duskClock(l *log) {
	t := content.T
	rep := s.Dusk
	rep.Before = s.Dreamer
	if rep.Turned.Valid() {
		rep.Rise = append(rep.Rise, Rise{RiseTurned, rep.Turned, TurnRise(rep.Neglect)})
	} else {
		rep.Rise = append(rep.Rise, Rise{RiseQuiet, content.NoDistrict, t.DreamerQuiet})
	}
	for id, h := range s.Hollows {
		if h != nil && h.Night < s.Day {
			rep.Rise = append(rep.Rise, Rise{RiseOpen, content.DistrictID(id), t.DreamerPerOpenHollow})
		}
	}
	total := 0
	for _, r := range rep.Rise {
		total += r.Tenths
	}
	s.Dreamer = min(s.Dreamer+total, t.DreamerMax)
	rep.After = s.Dreamer
	rep.Line = content.DreamLine(s.Dreamer/10, s.Day)
	l.add(EvDreamerRose).n(total).v(s.Dreamer)
	l.add(EvDreamLine).id(rep.Line)
	if s.Dreamer >= t.DreamerMax {
		s.endTenure(LostLid, l)
	}
}

// TurnRise is how far the Dreamer rises, in tenths of a point, when a district
// turns at the given Neglect: Neglect divided by ten, in points.
func TurnRise(neglect int) int { return neglect * 10 / content.T.DreamerNeglectDiv }

// dawn is where a tenure can end after a night. It reports whether it did.
func (s *State) dawn(l *log) bool {
	s.Nights = s.Day
	switch {
	case s.Living() == 0:
		s.endTenure(LostNoParty, l)
	case s.Day >= content.T.Nights:
		s.endTenure(Won, l)
	default:
		return false
	}
	return true
}

func (s *State) endTenure(o Outcome, l *log) {
	s.Phase, s.Outcome, s.Expedition = PhaseOver, o, nil
	l.add(EvTenureEnded).id(string(o)).n(s.Nights)
}

// searchHaven gives up the next of the old warden's notes, once per safe room.
func (s *State) searchHaven(r *Room, l *log) {
	if r.Searched || s.Fragments >= min(content.T.FragmentsPerTenure, len(content.Fragments)) {
		return
	}
	r.Searched = true
	l.add(EvFragment).in(s.Expedition.District).id(r.Name).n(s.Fragments)
	s.Fragments++
}

// CanBuy reports whether the city holds the materials for an upgrade.
func (s *State) CanBuy(id content.UpgradeID) bool {
	cost := content.Upgrades[id].Cost()
	for f := range cost {
		if s.Materials[f] < cost[f] {
			return false
		}
	}
	return true
}

func (s *State) buyUpgrade(c Command, l *log) error {
	if c.Upgrade < 0 || c.Upgrade >= content.NumUpgrades {
		return fmt.Errorf("%w: no such upgrade", ErrBadArgument)
	}
	up := content.Upgrades[c.Upgrade]
	ev := Event{Kind: EvUpgradeBought, ID: up.Slug, District: content.NoDistrict, Citizen: -1}
	if up.Building {
		if !c.District.Valid() {
			return fmt.Errorf("%w: a %s needs a building", ErrBadArgument, up.Name)
		}
		d := &s.Districts[c.District]
		if d.Turned {
			return fmt.Errorf("%w: nobody can reach the %s", ErrTurned, c.District)
		}
		if c.Building < 0 || c.Building >= len(d.Buildings) {
			return fmt.Errorf("%w: the %s has no building %d", ErrBadArgument, c.District, c.Building)
		}
		if d.Buildings[c.Building].Warded {
			return fmt.Errorf("%w: that building is already warded", ErrAlreadyDone)
		}
	} else if s.Owned[c.Upgrade] {
		return fmt.Errorf("%w: the city already has %s", ErrAlreadyDone, up.Name)
	}
	if !s.CanBuy(c.Upgrade) {
		return fmt.Errorf("%w: %s", ErrCannotAfford, up.Name)
	}
	cost := up.Cost()
	for f := range cost {
		s.Materials[f] -= cost[f]
	}
	if up.Building {
		s.Districts[c.District].Buildings[c.Building].Warded = true
		ev.District, ev.V = c.District, c.Building
	} else {
		s.Owned[c.Upgrade] = true
		for i := range s.Citizens {
			cz := &s.Citizens[i]
			cz.Gear.HP += up.Gear.HP
			cz.Gear.Attack += up.Gear.Attack
			cz.Gear.Guard += up.Gear.Guard
			cz.Gear.Speed += up.Gear.Speed
			if cz.Alive {
				cz.HP = cz.Stats().HP
			}
		}
	}
	ev.Day = l.day
	l.events = append(l.events, ev)
	return nil
}
