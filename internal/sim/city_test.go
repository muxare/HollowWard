package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"hollowward/internal/content"
)

func TestNeglect(t *testing.T) {
	tests := []struct {
		name      string
		needs     [3]int
		conds     []int
		want      int
		wantFirst string // ID of the largest cause
		wantN     int    // number of causes
	}{
		{"all full", [3]int{100, 100, 100}, []int{80, 80}, 0, "", 0},
		{"all half", [3]int{50, 50, 50}, []int{80, 80}, 50, "market/supply", 3},
		{"starved market", [3]int{18, 80, 80}, []int{80, 80}, 41, "market/supply", 3},
		{"order worst", [3]int{70, 10, 60}, []int{80, 80}, 53, "market/order", 3},
		{"one ruin adds ten", [3]int{50, 50, 50}, []int{29, 80}, 60, "market/supply", 4},
		{"thirty is not ruined", [3]int{50, 50, 50}, []int{30, 80}, 50, "market/supply", 3},
		{"ruin outranks a mild need", [3]int{90, 95, 100}, []int{10, 80}, 15, "market/granary", 3},
		{"not capped at 100", [3]int{0, 0, 0}, []int{0, 0, 0}, 130, "market/supply", 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := District{ID: content.Market, Needs: tt.needs}
			for _, c := range tt.conds {
				d.Buildings = append(d.Buildings, Building{Condition: c})
			}
			rep := Neglect(&d)
			if rep.Neglect != tt.want {
				t.Errorf("Neglect = %d, want %d", rep.Neglect, tt.want)
			}
			if len(rep.Causes) != tt.wantN {
				t.Errorf("%d causes, want %d: %+v", len(rep.Causes), tt.wantN, rep.Causes)
			}
			if tt.wantFirst != "" && rep.Causes[0].ID != tt.wantFirst {
				t.Errorf("largest cause %s, want %s", rep.Causes[0].ID, tt.wantFirst)
			}
			sum := 0
			for i, c := range rep.Causes {
				sum += c.Thirds
				if i > 0 && c.Thirds > rep.Causes[i-1].Thirds {
					t.Errorf("causes not largest first: %+v", rep.Causes)
				}
				if v, ok := c.Resolve(&d); !ok || v != c.ValueAtGeneration {
					t.Errorf("cause %+v does not point at real state (got %d, %v)", c.Cause, v, ok)
				}
			}
			if (sum+1)/3 != rep.Neglect {
				t.Errorf("causes add up to %d thirds, Neglect is %d", sum, rep.Neglect)
			}
		})
	}
}

func TestNewTenureShape(t *testing.T) {
	tn := content.T
	for _, seed := range []uint64{1, 2, 3, 99, 12345} {
		s := start(t, seed)
		if s.Day != 1 || s.Phase != PhaseDay || s.Resources != tn.StartResources {
			t.Fatalf("seed %d: bad opening: day %d phase %q", seed, s.Day, s.Phase)
		}
		if len(s.Citizens) < 18 || len(s.Citizens) > 22 {
			t.Errorf("seed %d: %d citizens", seed, len(s.Citizens))
		}
		names := map[string]bool{}
		home := map[content.DistrictID]int{}
		for i, c := range s.Citizens {
			if c.ID != i || !c.Alive || c.HP != c.Stats().HP || c.HP <= 0 {
				t.Errorf("seed %d: citizen %d badly made: %+v", seed, i, c)
			}
			if c.Traits[0] == c.Traits[1] {
				t.Errorf("seed %d: %s has the same trait twice", seed, c.Name)
			}
			if opp, ok := content.Opposite(c.Traits[0]); ok && opp == c.Traits[1] {
				t.Errorf("seed %d: %s has opposite traits", seed, c.Name)
			}
			if names[c.Name] {
				t.Errorf("seed %d: two citizens named %s", seed, c.Name)
			}
			names[c.Name] = true
			home[c.Home]++
		}
		figures := map[string]bool{}
		for id, d := range s.Districts {
			if d.ID != content.DistrictID(id) || d.Turned || d.Figure == "" || figures[d.Figure] {
				t.Errorf("seed %d: district %d badly made", seed, id)
			}
			figures[d.Figure] = true
			if n := home[d.ID]; n < 2 || n > 4 {
				t.Errorf("seed %d: %s has %d residents", seed, d.ID, n)
			}
			for _, v := range d.Needs {
				if v < tn.StartNeedMin || v > tn.StartNeedMax {
					t.Errorf("seed %d: starting need %d out of range", seed, v)
				}
			}
			if len(d.Buildings) != len(content.Districts[id].Buildings) {
				t.Errorf("seed %d: %s building count", seed, d.ID)
			}
		}
		if s.Card == nil || s.Card.Resolved || !s.Card.District.Valid() {
			t.Errorf("seed %d: day 1 has no event card", seed)
		}
	}
}

func TestNewTenureReshuffles(t *testing.T) {
	a, b, again := start(t, 7), start(t, 8), start(t, 7)
	if encode(t, a) != encode(t, again) {
		t.Error("the same seed built two different cities")
	}
	if encode(t, a.Districts) == encode(t, b.Districts) {
		t.Error("two seeds built the same districts")
	}
	if encode(t, a.Citizens) == encode(t, b.Citizens) {
		t.Error("two seeds built the same citizens")
	}
}

func TestStewards(t *testing.T) {
	tn := content.T
	tests := []struct {
		name      string
		needs     [3]int
		conds     []int // market has three buildings
		need      content.Need
		building  int
		wantNeeds [3]int
		wantConds []int
	}{
		{"defaults pick the lowest need and worst building",
			[3]int{60, 40, 50}, []int{70, 50, 60}, content.NoNeed, -1,
			[3]int{60, 70, 50}, []int{70, 65, 60}},
		{"the player's picks are honoured",
			[3]int{60, 40, 50}, []int{70, 50, 60}, content.Supply, 0,
			[3]int{90, 40, 50}, []int{85, 50, 60}},
		{"gains stop at the maximum",
			[3]int{90, 95, 99}, []int{95, 96, 97}, content.Upkeep, 2,
			[3]int{90, 95, 100}, []int{95, 96, 100}},
		{"ties go to the first",
			[3]int{40, 40, 40}, []int{50, 50, 50}, content.NoNeed, -1,
			[3]int{70, 40, 40}, []int{65, 50, 50}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := start(t, 1)
			m := &s.Districts[content.Market]
			m.Needs = tt.needs
			for i, c := range tt.conds {
				m.Buildings[i].Condition = c
			}
			s, _ = do(t, s, AssignSteward(0, content.Market, tt.need, tt.building))
			s.stewardsWork(&log{})
			m = &s.Districts[content.Market]
			if m.Needs != tt.wantNeeds {
				t.Errorf("needs %v, want %v", m.Needs, tt.wantNeeds)
			}
			for i, want := range tt.wantConds {
				if m.Buildings[i].Condition != want {
					t.Errorf("building %d at %d, want %d", i, m.Buildings[i].Condition, want)
				}
			}
		})
	}

	t.Run("recall and reassign", func(t *testing.T) {
		s := start(t, 1)
		s, _ = do(t, s, AssignSteward(1, content.Docks, content.NoNeed, -1))
		s, _ = do(t, s, AssignSteward(1, content.Mills, content.Order, 1))
		if s.Stewards[1].District != content.Mills || s.Stewards[1].Need != content.Order {
			t.Errorf("reassign did not stick: %+v", s.Stewards[1])
		}
		s, _ = do(t, s, RecallSteward(1))
		if s.Stewards[1].District != content.NoDistrict {
			t.Errorf("recall did not stick: %+v", s.Stewards[1])
		}
	})
	t.Run("two stewards may share a district", func(t *testing.T) {
		s := start(t, 1)
		s.Districts[content.Watch].Needs = [3]int{10, 20, 30}
		s, _ = do(t, s, AssignSteward(0, content.Watch, content.NoNeed, -1))
		s, _ = do(t, s, AssignSteward(1, content.Watch, content.NoNeed, -1))
		s.stewardsWork(&log{})
		// first lifts Supply 10 to 40; the second then finds Order lowest, at 20.
		if got := s.Districts[content.Watch].Needs; got != [3]int{10 + tn.StewardNeed, 20 + tn.StewardNeed, 30} {
			t.Errorf("needs %v", got)
		}
	})
	t.Run("a turned district takes no steward", func(t *testing.T) {
		s := start(t, 1)
		s.Districts[content.Temple].Turned = true
		mustFail(t, s, AssignSteward(0, content.Temple, content.NoNeed, -1), ErrTurned)
	})
}

func TestDecreesApply(t *testing.T) {
	tn := content.T
	rich := content.Cost{100, 100, 100}
	for id := content.DecreeID(0); id < content.NumDecrees; id++ {
		dec := content.Decrees[id]
		t.Run(dec.Slug, func(t *testing.T) {
			s := start(t, 3)
			flat(&s, 40, 40)
			s.Resources = rich
			s.Districts[content.Docks].Turned = true // beyond a decree's reach
			s, ev := do(t, s, IssueDecree(id, content.Mills))
			for r := range rich {
				if s.Resources[r] != rich[r]-dec.Cost[r] {
					t.Errorf("%s: paid %d, want %d", content.Resource(r), rich[r]-s.Resources[r], dec.Cost[r])
				}
			}
			if len(ev) != 1 || ev[0].Kind != EvDecreeIssued || ev[0].ID != dec.Slug {
				t.Errorf("events %+v", ev)
			}
			for d := range s.Districts {
				dist := &s.Districts[d]
				wantNeeds, wantCond := [3]int{40, 40, 40}, 40
				reached := !dist.Turned && (!dec.Targeted || dist.ID == content.Mills)
				if reached {
					switch id {
					case content.RationGrain:
						wantNeeds[content.Supply] += tn.DecreeCityNeed
					case content.Curfew:
						wantNeeds[content.Order] += tn.DecreeCityNeed
					case content.Corvee:
						wantNeeds[content.Upkeep] += tn.DecreeCityNeed
					case content.ShoreUp:
						wantCond += tn.DecreeShoreUp
					case content.Relief:
						wantNeeds = [3]int{40 + tn.DecreeRelief, 40 + tn.DecreeRelief, 40 + tn.DecreeRelief}
					}
				}
				if dist.Needs != wantNeeds {
					t.Errorf("%s needs %v, want %v", dist.ID, dist.Needs, wantNeeds)
				}
				for _, b := range dist.Buildings {
					if b.Condition != wantCond {
						t.Errorf("%s building at %d, want %d", dist.ID, b.Condition, wantCond)
					}
				}
			}
			mustFail(t, s, IssueDecree(id, content.Mills), ErrAlreadyDone)
		})
	}

	t.Run("cannot afford", func(t *testing.T) {
		s := start(t, 3)
		s.Resources = content.Cost{}
		for id := content.DecreeID(0); id < content.NumDecrees; id++ {
			mustFail(t, s, IssueDecree(id, content.Mills), ErrCannotAfford)
		}
	})
	t.Run("a targeted decree needs a reachable district", func(t *testing.T) {
		s := start(t, 3)
		s.Resources = rich
		mustFail(t, s, IssueDecree(content.Relief, content.NoDistrict), ErrBadArgument)
		s.Districts[content.Mills].Turned = true
		mustFail(t, s, IssueDecree(content.Relief, content.Mills), ErrTurned)
	})
	t.Run("a new day allows a new decree", func(t *testing.T) {
		s := start(t, 3)
		s.Resources = rich
		s, _ = do(t, s, IssueDecree(content.Curfew, content.NoDistrict))
		s = pass(t, s)
		if s.DecreeIssued {
			t.Fatal("the decree flag survived the night")
		}
		do(t, s, IssueDecree(content.Curfew, content.NoDistrict))
	})
}

func TestEventCardsResolve(t *testing.T) {
	for ci, card := range content.EventCards {
		for chi, ch := range card.Choices {
			t.Run(card.Slug+"/"+ch.Label, func(t *testing.T) {
				s := start(t, 5)
				flat(&s, 50, 50)
				s.Resources = content.Cost{50, 50, 50}
				target := card.District
				if target == content.NoDistrict {
					target = content.Temple
				}
				s.Card = &DrawnCard{Card: ci, District: target}
				before := s.Clone()
				s, ev := do(t, s, ResolveEvent(chi))
				if len(ev) != 1 || ev[0].Kind != EvCardResolved || ev[0].ID != card.Slug || ev[0].N != chi {
					t.Errorf("events %+v", ev)
				}
				if !s.Card.Resolved || s.Card.Choice != chi {
					t.Error("card not marked resolved")
				}
				// The cost is paid and something else changed.
				for r := range ch.Cost {
					gain := 0
					for _, e := range ch.Effects {
						if e.Kind == content.EffResource && int(e.Resource) == r {
							gain += e.Amount
						}
					}
					if s.Resources[r] != before.Resources[r]-ch.Cost[r]+gain {
						t.Errorf("%s: %d, want %d", content.Resource(r), s.Resources[r], before.Resources[r]-ch.Cost[r]+gain)
					}
				}
				if reflect.DeepEqual(s.Districts, before.Districts) && s.Resources == before.Resources {
					t.Error("the choice changed nothing")
				}
				mustFail(t, s, ResolveEvent(chi), ErrAlreadyDone)
			})
		}
	}

	t.Run("a paid choice needs the price", func(t *testing.T) {
		s := start(t, 5)
		s.Resources = content.Cost{}
		for ci, card := range content.EventCards {
			for chi, ch := range card.Choices {
				if ch.Cost.IsZero() {
					continue
				}
				s.Card = &DrawnCard{Card: ci, District: content.Temple}
				mustFail(t, s, ResolveEvent(chi), ErrCannotAfford)
			}
		}
	})
	t.Run("the day cannot end with the card unanswered", func(t *testing.T) {
		s := start(t, 5)
		mustFail(t, s, EndDay(), ErrEventPending)
	})
	t.Run("a turned district is beyond an event's reach", func(t *testing.T) {
		s := start(t, 5)
		flat(&s, 50, 50)
		s.Districts[content.Docks].Turned = true
		s.Card = &DrawnCard{Card: 0, District: content.Docks} // the barge: Docks Supply
		s, _ = do(t, s, ResolveEvent(1))
		if s.Districts[content.Docks].Needs != [3]int{50, 50, 50} {
			t.Errorf("the turned Docks changed: %v", s.Districts[content.Docks].Needs)
		}
	})
}

func TestDeckDealsEveryCardOnce(t *testing.T) {
	s := start(t, 11)
	seen := map[int]int{}
	for day := 1; day <= 2*len(content.EventCards); day++ {
		seen[s.Card.Card]++
		if day == len(content.EventCards) {
			for c := range content.EventCards {
				if seen[c] != 1 {
					t.Fatalf("after one pass card %d was seen %d times", c, seen[c])
				}
			}
		}
		s = pass(t, s)
	}
	for c := range content.EventCards {
		if seen[c] != 2 {
			t.Errorf("after two passes card %d was seen %d times", c, seen[c])
		}
	}
}

func TestDecay(t *testing.T) {
	tn := content.T
	s := start(t, 2)
	flat(&s, 60, 60)
	s.Districts[content.Market].Buildings[0].Condition = tn.RuinThreshold + 1 // about to be ruined
	s.Districts[content.Temple].Turned = true
	l := &log{}
	s.decay(l)
	for id := range s.Districts {
		d := &s.Districts[id]
		for n := range d.Needs {
			want := 60 - tn.NeedBaseDecay - tn.NeedDecayAdj[id][n]
			if d.Turned {
				want = 60
			}
			if d.Needs[n] != want {
				t.Errorf("%s %s = %d, want %d", d.ID, content.Need(n), d.Needs[n], want)
			}
		}
	}
	if got := s.Districts[content.Docks].Buildings[0].Condition; got != 60-tn.BuildingDecay {
		t.Errorf("building decayed to %d, want %d", got, 60-tn.BuildingDecay)
	}
	if got := s.Districts[content.Temple].Buildings[0].Condition; got != 60 {
		t.Errorf("a turned district's building decayed to %d", got)
	}
	if len(l.events) != 1 || l.events[0].Kind != EvRuined || l.events[0].ID != "granary" {
		t.Errorf("want one ruined event for the granary, got %+v", l.events)
	}

	t.Run("needs stop at zero", func(t *testing.T) {
		s := start(t, 2)
		flat(&s, 3, 2)
		s.decay(&log{})
		for _, d := range s.Districts {
			for _, v := range d.Needs {
				if v != 0 {
					t.Errorf("need at %d", v)
				}
			}
			for _, b := range d.Buildings {
				if b.Condition != 0 {
					t.Errorf("condition at %d", b.Condition)
				}
			}
		}
	})
}

func TestOutput(t *testing.T) {
	tn := content.T
	plain := [2]content.TraitID{content.Wary, content.Strong} // no effect on work
	tests := []struct {
		name  string
		setup func(s *State)
		pct   int // labour, in percent-days, of the Mills' four residents
		need  int
	}{
		{"full needs, everyone working", func(s *State) {}, 400, 100},
		{"half needs", func(s *State) {}, 400, 50},
		{"empty needs still yield the floor", func(s *State) {}, 400, 0},
		{"one exhausted", func(s *State) { s.Citizens[millers(s)[0]].Exhausted = true }, 300 + tn.ExhaustedWorkPct, 100},
		{"one dead", func(s *State) { s.Citizens[millers(s)[0]].Alive = false }, 300, 100},
		{"tireless shrugs off exhaustion", func(s *State) {
			c := &s.Citizens[millers(s)[0]]
			c.Exhausted, c.Traits[0] = true, content.Tireless
		}, 400, 100},
		{"diligent and scarred", func(s *State) {
			c := &s.Citizens[millers(s)[0]]
			c.Traits[0], c.Scars = content.Diligent, []content.ScarID{content.HollowEye}
		}, 300 + 100 + tn.TraitMods[content.Diligent].WorkPct + tn.ScarMods[content.HollowEye].WorkPct, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := start(t, 4)
			for i := range s.Citizens {
				s.Citizens[i].Traits = plain
			}
			flat(&s, tt.need, 80)
			tt.setup(&s)
			factor := tn.OutputFloorPct + (100-tn.OutputFloorPct)*tt.need/100
			want := tn.Yield[content.Mills][content.Grain] * tt.pct * factor / 10000
			if got := s.Output(content.Mills); got[content.Grain] != want || got[content.Coin] != 0 {
				t.Errorf("Mills output %v, want %d grain", got, want)
			}
		})
	}
	t.Run("a turned district produces nothing", func(t *testing.T) {
		s := start(t, 4)
		s.Districts[content.Mills].Turned = true
		if got := s.Output(content.Mills); got != (content.Cost{}) {
			t.Errorf("output %v", got)
		}
	})
	t.Run("production reaches the stores", func(t *testing.T) {
		s := start(t, 4)
		var want content.Cost
		for d := content.DistrictID(0); d < content.NumDistricts; d++ {
			out := s.Output(d)
			for r := range out {
				want[r] += out[r]
			}
		}
		before := s.Resources
		l := &log{}
		s.produce(l)
		logged := content.Cost{}
		for _, e := range l.events {
			for r := content.Resource(0); r < content.NumResources; r++ {
				if e.ID == r.String() {
					logged[r] += e.N
				}
			}
		}
		for r := range want {
			if s.Resources[r] != before[r]+want[r] || logged[r] != want[r] {
				t.Errorf("%s: have %d, want %d; logged %d", content.Resource(r), s.Resources[r], before[r]+want[r], logged[r])
			}
		}
	})
}

func millers(s *State) []int {
	var out []int
	for i, c := range s.Citizens {
		if c.Home == content.Mills {
			out = append(out, i)
		}
	}
	return out
}

func TestEndDayOrder(t *testing.T) {
	// Stewards work, then districts produce, then needs decay.
	tn := content.T
	s := start(t, 6)
	flat(&s, 40, 50)
	s = answer(t, s)
	flat(&s, 40, 50) // undo the card's effects
	s, _ = do(t, s, AssignSteward(0, content.Mills, content.Supply, 0))
	produced := func() content.Cost { // what the Mills make after the steward's visit
		c := s.Clone()
		c.stewardsWork(&log{})
		return c.Output(content.Mills)
	}()
	next, ev := do(t, s, EndDay())
	wantSupply := 40 + tn.StewardNeed - tn.NeedBaseDecay - tn.NeedDecayAdj[content.Mills][content.Supply]
	if got := next.Districts[content.Mills].Needs[content.Supply]; got != wantSupply {
		t.Errorf("Mills Supply %d, want %d", got, wantSupply)
	}
	if got := next.Districts[content.Mills].Buildings[0].Condition; got != 50+tn.StewardBuilding-tn.BuildingDecay {
		t.Errorf("Mills building %d", got)
	}
	var grain int
	order := []EventKind{}
	for _, e := range ev {
		if len(order) == 0 || order[len(order)-1] != e.Kind {
			order = append(order, e.Kind)
		}
		if e.Kind == EvProduced && e.District == content.Mills {
			grain = e.N
		}
	}
	if grain != produced[content.Grain] {
		t.Errorf("Mills made %d grain, want %d (produced after the steward, before decay)", grain, produced[content.Grain])
	}
	want := []EventKind{EvStewardWorked, EvRepaired, EvProduced}
	for i, k := range want {
		if i >= len(order) || order[i] != k {
			t.Fatalf("event order %v, want it to start %v", order, want)
		}
	}
	if next.Phase != PhaseDusk || next.Dusk == nil {
		t.Fatalf("dusk did not fall: phase %q", next.Phase)
	}
	next = sleep(t, next)
	if next.Day != s.Day+1 || next.Phase != PhaseDay || next.Card == nil || next.Card.Resolved {
		t.Errorf("the next day did not open properly: day %d, phase %q", next.Day, next.Phase)
	}
}

func TestIllegalCommands(t *testing.T) {
	s := start(t, 9)
	tests := []struct {
		name string
		s    State
		c    Command
		want error
	}{
		{"nothing before a tenure", State{}, EndDay(), ErrWrongPhase},
		{"no second tenure on top", s, NewTenure(1), ErrWrongPhase},
		{"unknown command", s, Command{Kind: "dance"}, ErrUnknown},
		{"steward out of range", s, AssignSteward(3, content.Market, content.NoNeed, -1), ErrBadArgument},
		{"steward negative", s, AssignSteward(-1, content.Market, content.NoNeed, -1), ErrBadArgument},
		{"district out of range", s, AssignSteward(0, 6, content.NoNeed, -1), ErrBadArgument},
		{"no district", s, AssignSteward(0, content.NoDistrict, content.NoNeed, -1), ErrBadArgument},
		{"need out of range", s, AssignSteward(0, content.Market, 3, -1), ErrBadArgument},
		{"building out of range", s, AssignSteward(0, content.Watch, content.NoNeed, 2), ErrBadArgument},
		{"recall out of range", s, RecallSteward(9), ErrBadArgument},
		{"decree out of range", s, IssueDecree(content.NumDecrees, content.Market), ErrBadArgument},
		{"choice out of range", s, ResolveEvent(7), ErrBadArgument},
		{"choice negative", s, ResolveEvent(-1), ErrBadArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { mustFail(t, tt.s, tt.c, tt.want) })
	}
}

func TestApplyLeavesItsInputAlone(t *testing.T) {
	s := start(t, 10)
	before := encode(t, s)
	cmds := []Command{
		AssignSteward(0, content.Market, content.NoNeed, -1),
		ResolveEvent(1),
		IssueDecree(content.RationGrain, content.NoDistrict),
	}
	cur := s
	for _, c := range cmds {
		cur, _ = do(t, cur, c)
		if encode(t, s) != before {
			t.Fatalf("%s changed the state it was given", c.Kind)
		}
	}
	prev := encode(t, cur)
	next, _ := do(t, cur, EndDay())
	if encode(t, cur) != prev {
		t.Fatal("end-day changed the state it was given")
	}
	if encode(t, next) == prev {
		t.Fatal("end-day changed nothing")
	}
}

func TestStateSurvivesJSON(t *testing.T) {
	s := start(t, 21)
	s, _ = do(t, s, AssignSteward(2, content.Docks, content.Order, 1))
	s = pass(t, s)
	first := encode(t, s)
	var back State
	if err := json.Unmarshal([]byte(first), &back); err != nil {
		t.Fatal(err)
	}
	if again := encode(t, back); again != first {
		t.Fatalf("round trip changed the state:\n%s\n%s", first, again)
	}
	// And the loaded game plays out the same future.
	a, b := s, back
	for range 5 {
		a, b = pass(t, a), pass(t, b)
	}
	if encode(t, a) != encode(t, b) {
		t.Fatal("the loaded game diverged from the original")
	}
}
