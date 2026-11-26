package sim

import "hollowward/internal/content"

// newTenure builds the starting state from a seed: a reshuffled city with the
// same six districts.
func newTenure(seed uint64, l *log) State {
	t := content.T
	s := State{
		Version:   SaveVersion,
		Seed:      seed,
		Day:       1,
		Phase:     PhaseDay,
		Resources: t.StartResources,
		RNG:       newStreams(seed),
	}
	rng := &s.RNG.Setup

	figures := rng.Perm(len(content.FigureNames))
	names := rng.Perm(len(content.CitizenNames))
	for d := content.DistrictID(0); d < content.NumDistricts; d++ {
		def := content.Districts[d]
		dist := District{ID: d, Figure: content.FigureNames[figures[d]]}
		for n := range dist.Needs {
			dist.Needs[n] = rng.Between(t.StartNeedMin, t.StartNeedMax)
		}
		for range def.Buildings {
			dist.Buildings = append(dist.Buildings, Building{Condition: rng.Between(t.StartCondMin, t.StartCondMax)})
		}
		s.Districts[d] = dist

		for _, job := range def.Jobs {
			c := Citizen{
				ID:          len(s.Citizens),
				Name:        content.CitizenNames[names[len(s.Citizens)%len(names)]],
				Job:         job,
				Home:        d,
				JitterHP:    rng.Between(-t.StatJitterHP, t.StatJitterHP),
				JitterSpeed: rng.Between(-t.StatJitterSpeed, t.StatJitterSpeed),
				Alive:       true,
			}
			c.Traits[0] = content.TraitID(rng.IntN(int(content.NumTraits)))
			c.Traits[1] = content.TraitID(rng.IntN(int(content.NumTraits) - 1))
			if c.Traits[1] >= c.Traits[0] {
				c.Traits[1]++ // two different traits
			}
			if opp, ok := content.Opposite(c.Traits[0]); ok && opp == c.Traits[1] {
				c.Traits[1] = (c.Traits[1] + 2) % content.NumTraits // and not opposites
			}
			c.HP = c.Stats().HP
			s.Citizens = append(s.Citizens, c)
		}
	}
	for i := range s.Stewards {
		s.Stewards[i] = Post{District: content.NoDistrict, Need: content.NoNeed, Building: -1}
	}

	l.day = s.Day
	l.add(EvTenureStarted)
	s.startDay(l)
	return s
}

// startDay opens a day: a fresh decree, a new event card.
func (s *State) startDay(l *log) {
	l.day = s.Day
	s.Phase = PhaseDay
	s.DecreeIssued = false
	l.add(EvDayStarted)
	s.drawCard(l)
}

// drawCard takes the top card of the deck, reshuffling when it is empty, and
// picks a target for cards that have none of their own.
func (s *State) drawCard(l *log) {
	rng := &s.RNG.Events
	if len(s.Deck) == 0 {
		s.Deck = rng.Perm(len(content.EventCards))
	}
	card := s.Deck[0]
	s.Deck = s.Deck[1:]
	target := content.EventCards[card].District
	// Always draw, so the stream advances the same way whatever the card.
	pick := content.DistrictID(rng.IntN(int(content.NumDistricts)))
	if target == content.NoDistrict {
		target = pick
	}
	s.Card = &DrawnCard{Card: card, District: target}
	l.add(EvCardDrawn).id(content.EventCards[card].Slug).in(target)
}
