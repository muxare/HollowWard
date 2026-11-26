// Package policy holds the scripted players (random, repair-worst, farm) that
// drive the sim for simcli, the balance harness and the smoke test.
//
// A policy reads a state and returns the next command; it never touches state
// directly and draws randomness only from a seeded source it is given.
package policy

import (
	"fmt"
	"sort"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

// Policy is a scripted player. Next returns the command to send in the
// current state. It must always return a command the rules accept.
type Policy interface {
	Name() string
	Next(s *sim.State) sim.Command
}

// Names lists the policies New can build.
func Names() []string { return []string{"random", "repair-worst", "farm"} }

// New builds a policy by name. The seed feeds the policy's own dice, which are
// separate from the game's.
func New(name string, seed uint64) (Policy, error) {
	switch name {
	case "random":
		return NewRandom(seed), nil
	case "repair-worst":
		return RepairWorst{}, nil
	case "farm":
		return Farm{Field: content.Market}, nil
	}
	return nil, fmt.Errorf("no policy called %q (have %v)", name, Names())
}

// Play runs a policy from the given state until the tenure is over or until
// stop returns true. It returns the final state and every event along the way.
// A policy that sends a command the rules reject is a bug, and is reported.
func Play(s sim.State, p Policy, stop func(*sim.State) bool) (sim.State, []sim.Event, error) {
	var log []sim.Event
	for steps := 0; s.Phase != sim.PhaseOver && s.Phase != sim.PhaseNone; steps++ {
		if stop != nil && stop(&s) {
			break
		}
		if steps > 200000 {
			return s, log, fmt.Errorf("policy %s: no end in sight after %d commands", p.Name(), steps)
		}
		cmd := p.Next(&s)
		next, ev, err := sim.Apply(s, cmd)
		if err != nil {
			return s, log, fmt.Errorf("policy %s on day %d (%s): %w", p.Name(), s.Day, s.Phase, err)
		}
		s, log = next, append(log, ev...)
	}
	return s, log, nil
}

// --- Pieces the policies share ---

// byNeglect lists the districts that have not turned, most neglected first.
func byNeglect(s *sim.State) []content.DistrictID {
	var out []content.DistrictID
	score := map[content.DistrictID]int{}
	for i := range s.Districts {
		if d := &s.Districts[i]; !d.Turned {
			out = append(out, d.ID)
			score[d.ID] = sim.Neglect(d).Neglect
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return score[out[a]] > score[out[b]] })
	return out
}

// freeChoice is the first choice of today's card that costs nothing.
func freeChoice(s *sim.State) int {
	for i, ch := range content.EventCards[s.Card.Card].Choices {
		if ch.Cost.IsZero() {
			return i
		}
	}
	return 0
}

// worth scores a city for comparing the outcomes of choices: needs count
// most where they are lowest, and stores count a little.
func worth(s *sim.State) int {
	total := 0
	for i := range s.Districts {
		d := &s.Districts[i]
		if d.Turned {
			continue
		}
		for _, v := range d.Needs {
			total += 3*v - v*v/100 // a point is worth more when the need is low
		}
		for _, b := range d.Buildings {
			total += b.Condition / 2
		}
	}
	for _, r := range s.Resources {
		total += 2 * r
	}
	return total
}

// bestChoice answers today's card with the affordable choice that leaves the
// city best off.
func bestChoice(s *sim.State) int {
	best, bestWorth := freeChoice(s), -1<<30
	for i := range content.EventCards[s.Card.Card].Choices {
		if next, _, err := sim.Apply(*s, sim.ResolveEvent(i)); err == nil {
			if w := worth(&next); w > bestWorth {
				best, bestWorth = i, w
			}
		}
	}
	return best
}

// bestDecree is the affordable decree that leaves the city best off, if any
// is worth more than keeping the resources. reserve is kept unspent.
func bestDecree(s *sim.State, allow func(content.DistrictID) bool) (sim.Command, bool) {
	var best sim.Command
	bestWorth := worth(s) + 20 // a decree must clearly pay for itself
	found := false
	for id := content.DecreeID(0); id < content.NumDecrees; id++ {
		targets := []content.DistrictID{content.NoDistrict}
		if content.Decrees[id].Targeted {
			targets = targets[:0]
			for d := content.DistrictID(0); d < content.NumDistricts; d++ {
				if allow == nil || allow(d) {
					targets = append(targets, d)
				}
			}
		}
		for _, target := range targets {
			cmd := sim.IssueDecree(id, target)
			if next, _, err := sim.Apply(*s, cmd); err == nil {
				if w := worth(&next); w > bestWorth {
					best, bestWorth, found = cmd, w, true
				}
			}
		}
	}
	return best, found
}

// strongest picks up to n living citizens, the hardiest first.
func strongest(s *sim.State, n int) []int {
	var ids []int
	for i := range s.Citizens {
		if s.Citizens[i].Alive {
			ids = append(ids, i)
		}
	}
	power := func(id int) int {
		st := s.Citizens[id].Stats()
		return st.HP + 3*st.Attack + 2*st.Guard + st.Speed
	}
	sort.SliceStable(ids, func(a, b int) bool { return power(ids[a]) > power(ids[b]) })
	if len(ids) > n {
		ids = ids[:n]
	}
	return ids
}

// fight is a sound, simple way to play a combat turn: mend the badly hurt,
// stun or drag the dangerous, otherwise finish off the weakest enemy.
func fight(s *sim.State) sim.Command {
	exp := s.Expedition
	slot := s.Acting()
	m := exp.Party[slot]
	me := &s.Citizens[m.Citizen]
	room := &s.Hollows[exp.District].Rooms[exp.Room]

	weakest, alive := -1, 0
	for i := range room.Enemies {
		if e := &room.Enemies[i]; e.HP > 0 {
			alive++
			if weakest < 0 || e.HP < room.Enemies[weakest].HP {
				weakest = i
			}
		}
	}
	strongestEnemy := weakest
	for i := range room.Enemies {
		if e := &room.Enemies[i]; e.HP > 0 && e.Attack+e.AttackMod > room.Enemies[strongestEnemy].Attack+room.Enemies[strongestEnemy].AttackMod {
			strongestEnemy = i
		}
	}

	if m.Cooldown == 0 {
		switch content.Jobs[me.Job].Ability {
		case content.Mend:
			worst, lack := -1, 0
			for _, pm := range exp.Party {
				c := &s.Citizens[pm.Citizen]
				if !c.Downed && c.Alive && c.Stats().HP-c.HP > lack {
					worst, lack = c.ID, c.Stats().HP-c.HP
				}
			}
			if lack >= content.T.Ability.MendHeal-1 {
				return sim.Act(sim.ActAbility, worst)
			}
		case content.Hold:
			// Guard whoever an enemy means to hit, if it is not us.
			for i := range room.Enemies {
				if e := &room.Enemies[i]; e.HP > 0 && e.Upcoming().Target >= 0 && e.Upcoming().Target != me.ID {
					if c := &s.Citizens[e.Upcoming().Target]; !c.Downed && c.HP*2 <= c.Stats().HP {
						return sim.Act(sim.ActAbility, c.ID)
					}
				}
			}
		case content.Hammer, content.Hook, content.Haggle:
			return sim.Act(sim.ActAbility, strongestEnemy)
		case content.Grind:
			return sim.Act(sim.ActAbility, weakest)
		case content.Heave:
			if alive >= 2 {
				return sim.Act(sim.ActAbility, -1)
			}
		}
	}
	return sim.Act(sim.ActAttack, weakest)
}

// partyHealth is the party's HP as a percentage of its full HP, counting only
// those still standing, and how many of them there are.
func partyHealth(s *sim.State) (pct, standing int) {
	have, full := 0, 0
	for _, m := range s.Expedition.Party {
		c := &s.Citizens[m.Citizen]
		full += c.Stats().HP
		if c.Alive && !c.Downed {
			have += c.HP
			standing++
		}
	}
	if full == 0 {
		return 0, 0
	}
	return have * 100 / full, standing
}

// explore decides between going deeper and going home, and picks the way:
// toward a safe room when hurt, otherwise the first way on.
func explore(s *sim.State, retreatBelow int) sim.Command {
	exp := s.Expedition
	h := s.Hollows[exp.District]
	room := &h.Rooms[exp.Room]
	pct, standing := partyHealth(s)
	if pct < retreatBelow || standing*2 < len(exp.Party) {
		return sim.GoHome()
	}
	next := room.Exits[0]
	for _, x := range room.Exits {
		if h.Rooms[x].Kind == sim.RoomHaven {
			next = x
		}
	}
	return sim.Advance(next)
}

// night plays dusk and the hollow for the scripted wardens: send the
// strongest four into the chosen hollow, fight sensibly, come home when hurt.
func night(s *sim.State, choose func(s *sim.State) content.DistrictID, retreatBelow int) sim.Command {
	switch s.Phase {
	case sim.PhaseDusk:
		target := choose(s)
		if !target.Valid() || s.Hollows[target] == nil {
			return sim.StayHome()
		}
		party := strongest(s, content.T.MaxParty)
		if len(party) == 0 {
			return sim.StayHome()
		}
		return sim.SendParty(target, party...)
	case sim.PhaseNight:
		if s.Expedition.Fight != nil {
			return fight(s)
		}
		return explore(s, retreatBelow)
	}
	return sim.EndDay()
}

// oldestHollow picks the hollow that has been open longest: each one left
// open feeds the Dreamer every night.
func oldestHollow(s *sim.State) content.DistrictID {
	best := content.NoDistrict
	for i, h := range s.Hollows {
		if h != nil && (best == content.NoDistrict || h.Night < s.Hollows[best].Night) {
			best = content.DistrictID(i)
		}
	}
	return best
}
