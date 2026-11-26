package sim

import (
	"fmt"

	"hollowward/internal/content"
)

func (s *State) assignSteward(c Command) error {
	if c.Steward < 0 || c.Steward >= len(s.Stewards) {
		return fmt.Errorf("%w: no steward %d", ErrBadArgument, c.Steward)
	}
	if !c.District.Valid() {
		return fmt.Errorf("%w: no such district", ErrBadArgument)
	}
	d := &s.Districts[c.District]
	if d.Turned {
		return fmt.Errorf("%w: stewards will not enter the %s", ErrTurned, c.District)
	}
	if c.Need != content.NoNeed && !c.Need.Valid() {
		return fmt.Errorf("%w: no such need", ErrBadArgument)
	}
	if c.Building < -1 || c.Building >= len(d.Buildings) {
		return fmt.Errorf("%w: the %s has no building %d", ErrBadArgument, c.District, c.Building)
	}
	s.Stewards[c.Steward] = Post{District: c.District, Need: c.Need, Building: c.Building}
	return nil
}

func (s *State) recallSteward(c Command) error {
	if c.Steward < 0 || c.Steward >= len(s.Stewards) {
		return fmt.Errorf("%w: no steward %d", ErrBadArgument, c.Steward)
	}
	s.Stewards[c.Steward] = Post{District: content.NoDistrict, Need: content.NoNeed, Building: -1}
	return nil
}

// CanAfford reports whether the city holds the cost.
func (s *State) CanAfford(c content.Cost) bool {
	for r := range c {
		if s.Resources[r] < c[r] {
			return false
		}
	}
	return true
}

func (s *State) pay(c content.Cost) {
	for r := range c {
		s.Resources[r] -= c[r]
	}
}

func (s *State) issueDecree(c Command, l *log) error {
	if s.DecreeIssued {
		return fmt.Errorf("%w: one decree a day", ErrAlreadyDone)
	}
	if c.Decree < 0 || c.Decree >= content.NumDecrees {
		return fmt.Errorf("%w: no such decree", ErrBadArgument)
	}
	dec := content.Decrees[c.Decree]
	target := content.NoDistrict
	if dec.Targeted {
		if !c.District.Valid() {
			return fmt.Errorf("%w: %s needs a district", ErrBadArgument, dec.Name)
		}
		if s.Districts[c.District].Turned {
			return fmt.Errorf("%w: %s cannot reach the %s", ErrTurned, dec.Name, c.District)
		}
		target = c.District
	}
	if !s.CanAfford(dec.Cost) {
		return fmt.Errorf("%w: %s", ErrCannotAfford, dec.Name)
	}
	s.pay(dec.Cost)
	s.applyEffects(dec.Effects, target)
	s.DecreeIssued = true
	l.add(EvDecreeIssued).id(dec.Slug).in(target)
	return nil
}

func (s *State) resolveEvent(c Command, l *log) error {
	if s.Card == nil || s.Card.Resolved {
		return fmt.Errorf("%w: no event waits for an answer", ErrAlreadyDone)
	}
	card := content.EventCards[s.Card.Card]
	if c.Choice < 0 || c.Choice >= len(card.Choices) {
		return fmt.Errorf("%w: %q has no choice %d", ErrBadArgument, card.Title, c.Choice)
	}
	ch := card.Choices[c.Choice]
	if !s.CanAfford(ch.Cost) {
		return fmt.Errorf("%w: %s", ErrCannotAfford, ch.Label)
	}
	s.pay(ch.Cost)
	s.applyEffects(ch.Effects, s.Card.District)
	s.Card.Resolved, s.Card.Choice = true, c.Choice
	l.add(EvCardResolved).id(card.Slug).in(s.Card.District).n(c.Choice)
	return nil
}

// applyEffects applies decree or event effects. A turned district is beyond
// the reach of both: effects that land on one do nothing.
func (s *State) applyEffects(effects []content.Effect, target content.DistrictID) {
	t := content.T
	for _, e := range effects {
		if e.Kind == content.EffResource {
			s.Resources[e.Resource] = max(s.Resources[e.Resource]+e.Amount, 0)
			continue
		}
		var reach []content.DistrictID
		switch e.Scope {
		case content.ScopeTarget:
			reach = []content.DistrictID{target}
		case content.ScopeFixed:
			reach = []content.DistrictID{e.District}
		case content.ScopeAll:
			for d := content.DistrictID(0); d < content.NumDistricts; d++ {
				reach = append(reach, d)
			}
		}
		for _, id := range reach {
			if !id.Valid() || s.Districts[id].Turned {
				continue
			}
			d := &s.Districts[id]
			switch e.Kind {
			case content.EffNeed:
				for n := content.Need(0); n < content.NumNeeds; n++ {
					if e.Need == content.NoNeed || e.Need == n {
						d.Needs[n] = clamp(d.Needs[n]+e.Amount, 0, t.NeedMax)
					}
				}
			case content.EffBuilding:
				for i := range d.Buildings {
					d.Buildings[i].Condition = clamp(d.Buildings[i].Condition+e.Amount, 0, t.NeedMax)
				}
			}
		}
	}
}

// endDay runs the day tick, in this order: stewards work, districts produce,
// needs and buildings decay. Then dusk falls.
func (s *State) endDay(l *log) error {
	if s.Card != nil && !s.Card.Resolved {
		return ErrEventPending
	}
	s.stewardsWork(l)
	s.produce(l)
	s.decay(l)
	for i := range s.Citizens {
		s.Citizens[i].Exhausted = false // yesterday's night is paid for
	}
	l.add(EvDayEnded)
	s.dusk(l)
	return nil
}

func (s *State) stewardsWork(l *log) {
	t := content.T
	for i, post := range s.Stewards {
		if !post.District.Valid() || s.Districts[post.District].Turned {
			continue
		}
		d := &s.Districts[post.District]
		n := post.Need
		if n == content.NoNeed {
			n = d.Grievance()
		}
		before := d.Needs[n]
		d.Needs[n] = clamp(before+t.StewardNeed, 0, t.NeedMax)
		l.add(EvStewardWorked).in(d.ID).id(n.Slug()).n(d.Needs[n] - before).v(i)

		b := post.Building
		if b < 0 {
			b = d.WorstBuilding()
		}
		if b < len(d.Buildings) {
			was := d.Buildings[b].Condition
			d.Buildings[b].Condition = clamp(was+t.StewardBuilding, 0, t.NeedMax)
			l.add(EvRepaired).in(d.ID).id(content.Districts[d.ID].Buildings[b].Slug).n(d.Buildings[b].Condition - was).v(i)
		}
	}
}

// Output is what a district would produce today.
func (s *State) Output(id content.DistrictID) content.Cost {
	t := content.T
	var out content.Cost
	d := &s.Districts[id]
	if d.Turned {
		return out
	}
	work := 0 // percent-days of labour
	for i := range s.Citizens {
		if s.Citizens[i].Home == id {
			work += s.Citizens[i].WorkPct()
		}
	}
	mean := (d.Needs[0] + d.Needs[1] + d.Needs[2]) / int(content.NumNeeds)
	factor := t.OutputFloorPct + (100-t.OutputFloorPct)*mean/t.NeedMax
	for r := range out {
		out[r] = t.Yield[id][r] * work * factor / 10000
	}
	return out
}

func (s *State) produce(l *log) {
	for id := content.DistrictID(0); id < content.NumDistricts; id++ {
		out := s.Output(id)
		for r := range out {
			if out[r] > 0 {
				s.Resources[r] += out[r]
				l.add(EvProduced).in(id).id(content.Resource(r).String()).n(out[r])
			}
		}
	}
}

// decay wears down every district that has not turned. A turned district is
// frozen as the dream took it.
func (s *State) decay(l *log) {
	t := content.T
	for id := range s.Districts {
		d := &s.Districts[id]
		if d.Turned {
			continue
		}
		for n := range d.Needs {
			d.Needs[n] = clamp(d.Needs[n]-t.NeedBaseDecay-t.NeedDecayAdj[id][n], 0, t.NeedMax)
		}
		for i := range d.Buildings {
			was := d.Buildings[i].Condition
			now := clamp(was-s.buildingDecay(d, i), 0, t.NeedMax)
			d.Buildings[i].Condition = now
			if was >= t.RuinThreshold && now < t.RuinThreshold {
				l.add(EvRuined).in(d.ID).id(content.Districts[d.ID].Buildings[i].Slug).n(now)
			}
		}
	}
}

func (s *State) buildingDecay(_ *District, _ int) int { return content.T.BuildingDecay }
