package policy

import (
	"math/rand/v2"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

// RepairWorst always staffs the most neglected districts and clears every
// hollow it can. It is the "careful warden" baseline.
type RepairWorst struct{}

// Name implements Policy.
func (RepairWorst) Name() string { return "repair-worst" }

// Next implements Policy.
func (RepairWorst) Next(s *sim.State) sim.Command {
	if s.Phase != sim.PhaseDay {
		return night(s, oldestHollow, 35)
	}
	return day(s, byNeglect(s), nil)
}

// day plays a day for the scripted wardens: stewards to the given districts
// in order, the best answer to the card, the best decree, then dusk.
func day(s *sim.State, staff []content.DistrictID, allow func(content.DistrictID) bool) sim.Command {
	for i := range s.Stewards {
		want := content.NoDistrict
		if i < len(staff) {
			want = staff[i]
		}
		post := s.Stewards[i]
		switch {
		case want == content.NoDistrict && post.District != content.NoDistrict:
			return sim.RecallSteward(i)
		case want != content.NoDistrict && (post.District != want || post.Need != content.NoNeed || post.Building != -1):
			return sim.AssignSteward(i, want, content.NoNeed, -1)
		}
	}
	if s.Card != nil && !s.Card.Resolved {
		return sim.ResolveEvent(bestChoice(s))
	}
	if !s.DecreeIssued {
		if cmd, ok := bestDecree(s, allow); ok {
			return cmd
		}
	}
	if up, ok := bestUpgrade(s); ok {
		return up
	}
	return sim.EndDay()
}

// Farm starves one district on purpose, to harvest what grows there, while
// holding the rest. It proves that neglect is a lever (pillar 3).
type Farm struct {
	Field content.DistrictID // the district left to starve
}

// Name implements Policy.
func (Farm) Name() string { return "farm" }

// Next implements Policy.
func (f Farm) Next(s *sim.State) sim.Command {
	if s.Phase != sim.PhaseDay {
		return night(s, oldestHollow, 35)
	}
	// Staff the most neglected districts, never the field.
	var staff []content.DistrictID
	for _, d := range byNeglect(s) {
		if d != f.Field {
			staff = append(staff, d)
		}
	}
	return day(s, staff, func(d content.DistrictID) bool { return d != f.Field })
}

// Random makes random legal moves. It is the floor: a warden who does not
// read the city should lose.
type Random struct {
	rng *rand.Rand
}

// NewRandom builds a random policy with its own seeded dice.
func NewRandom(seed uint64) *Random {
	return &Random{rng: rand.New(rand.NewPCG(seed, 0x52414e44))}
}

// Name implements Policy.
func (*Random) Name() string { return "random" }

// Next implements Policy.
func (p *Random) Next(s *sim.State) sim.Command {
	switch s.Phase {
	case sim.PhaseDay:
		return p.day(s)
	case sim.PhaseDusk:
		var open []content.DistrictID
		for i, h := range s.Hollows {
			if h != nil {
				open = append(open, content.DistrictID(i))
			}
		}
		var living []int
		for i := range s.Citizens {
			if s.Citizens[i].Alive {
				living = append(living, i)
			}
		}
		if len(open) == 0 || len(living) == 0 || p.rng.IntN(3) == 0 {
			return sim.StayHome()
		}
		p.rng.Shuffle(len(living), func(a, b int) { living[a], living[b] = living[b], living[a] })
		size := min(1+p.rng.IntN(content.T.MaxParty), len(living))
		return sim.SendParty(open[p.rng.IntN(len(open))], living[:size]...)
	case sim.PhaseNight:
		return p.night(s)
	}
	return sim.EndDay()
}

func (p *Random) day(s *sim.State) sim.Command {
	// Try a handful of random orders; whatever the rules accept, send.
	for range 8 {
		var cmd sim.Command
		switch p.rng.IntN(5) {
		case 0:
			cmd = sim.AssignSteward(p.rng.IntN(content.NumStewards), content.DistrictID(p.rng.IntN(int(content.NumDistricts))),
				content.Need(p.rng.IntN(int(content.NumNeeds)+1)-1), -1)
		case 1:
			cmd = sim.IssueDecree(content.DecreeID(p.rng.IntN(int(content.NumDecrees))), content.DistrictID(p.rng.IntN(int(content.NumDistricts))))
		case 2:
			if s.Card != nil && !s.Card.Resolved {
				cmd = sim.ResolveEvent(p.rng.IntN(len(content.EventCards[s.Card.Card].Choices)))
			} else {
				cmd = sim.EndDay()
			}
		case 3:
			cmd = sim.RecallSteward(p.rng.IntN(content.NumStewards))
		default:
			cmd = sim.EndDay()
		}
		if _, _, err := sim.Apply(*s, cmd); err == nil {
			return cmd
		}
	}
	if s.Card != nil && !s.Card.Resolved {
		return sim.ResolveEvent(freeChoice(s))
	}
	return sim.EndDay()
}

func (p *Random) night(s *sim.State) sim.Command {
	exp := s.Expedition
	room := &s.Hollows[exp.District].Rooms[exp.Room]
	if exp.Fight == nil {
		if p.rng.IntN(4) == 0 || len(room.Exits) == 0 {
			return sim.GoHome()
		}
		return sim.Advance(room.Exits[p.rng.IntN(len(room.Exits))])
	}
	var alive []int
	for i := range room.Enemies {
		if room.Enemies[i].HP > 0 {
			alive = append(alive, i)
		}
	}
	target := alive[p.rng.IntN(len(alive))]
	switch p.rng.IntN(4) {
	case 0:
		return sim.Act(sim.ActDefend, -1)
	case 1:
		// An ability, aimed at random; fall back to a blow if the rules refuse.
		aim := target
		if ab := content.Jobs[s.Citizens[exp.Party[s.Acting()].Citizen].Job].Ability; ab == content.Hold || ab == content.Mend {
			aim = exp.Party[p.rng.IntN(len(exp.Party))].Citizen
		}
		cmd := sim.Act(sim.ActAbility, aim)
		if _, _, err := sim.Apply(*s, cmd); err == nil {
			return cmd
		}
	}
	return sim.Act(sim.ActAttack, target)
}

// bestUpgrade is where the wardens spend materials. Upgrades arrive with the
// full tenure (M4).
func bestUpgrade(_ *sim.State) (sim.Command, bool) { return sim.Command{}, false }
