package sim

import (
	"testing"

	"hollowward/internal/content"
)

// toDusk sets the city by hand and ends the day, so dusk reads known numbers.
// Needs are given as they should stand at dusk, after the day's decay.
func toDusk(t testing.TB, seed uint64, setup func(s *State)) State {
	t.Helper()
	tn := content.T
	s := start(t, seed)
	s = answer(t, s)
	calm(&s)
	setup(&s)
	for id := range s.Districts { // undo the decay the day tick is about to apply
		d := &s.Districts[id]
		for n := range d.Needs {
			d.Needs[n] += tn.NeedBaseDecay + tn.NeedDecayAdj[id][n]
		}
		for b := range d.Buildings {
			d.Buildings[b].Condition += tn.BuildingDecay
		}
	}
	s, _ = do(t, s, EndDay())
	if s.Phase != PhaseDusk || s.Dusk == nil {
		t.Fatalf("dusk did not fall: phase %q", s.Phase)
	}
	return s
}

func setNeeds(s *State, d content.DistrictID, v int) {
	s.Districts[d].Needs = [3]int{v, v, v}
}

func TestDuskTurnsTheMostNeglected(t *testing.T) {
	tn := content.T
	tests := []struct {
		name  string
		setup func(s *State)
		want  content.DistrictID
	}{
		{"a well-kept city has a quiet night", func(s *State) {}, content.NoDistrict},
		{"one point under the threshold is quiet", func(s *State) { setNeeds(s, content.Docks, 100-tn.TurnThreshold+1) }, content.NoDistrict},
		{"at the threshold it turns", func(s *State) { setNeeds(s, content.Docks, 100-tn.TurnThreshold) }, content.Docks},
		{"the highest Neglect turns, not the first", func(s *State) {
			setNeeds(s, content.Market, 50)
			setNeeds(s, content.Mills, 30)
			setNeeds(s, content.Watch, 40)
		}, content.Mills},
		{"ruined buildings count", func(s *State) {
			setNeeds(s, content.Market, 45)
			setNeeds(s, content.Temple, 50)
			s.Districts[content.Temple].Buildings[0].Condition = 0
		}, content.Temple},
		{"a district already turned cannot turn again", func(s *State) {
			setNeeds(s, content.Mills, 0)
			s.Districts[content.Mills].Turned = true
			setNeeds(s, content.Watch, 40)
		}, content.Watch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := toDusk(t, 1, tt.setup)
			if s.Dusk.Turned != tt.want {
				t.Fatalf("turned %s, want %s", s.Dusk.Turned, tt.want)
			}
			for id := range s.Districts {
				d := &s.Districts[id]
				preTurned := tt.name == "a district already turned cannot turn again" && d.ID == content.Mills
				if d.Turned != (d.ID == tt.want || preTurned) {
					t.Errorf("%s turned = %v", d.ID, d.Turned)
				}
				if (s.Hollows[id] != nil) != (d.ID == tt.want) {
					t.Errorf("%s hollow present = %v", d.ID, s.Hollows[id] != nil)
				}
			}
			if tt.want.Valid() {
				d := &s.Districts[tt.want]
				h := s.Hollows[tt.want]
				if msg := checkHollow(d, h); msg != "" {
					t.Error(msg)
				}
				if s.Dusk.Neglect != Neglect(d).Neglect || h.Night != s.Day {
					t.Errorf("dusk report %+v does not match the district", s.Dusk)
				}
			}
		})
	}
}

func TestDuskBreaksTiesBySeed(t *testing.T) {
	tie := func(s *State) {
		setNeeds(s, content.Market, 30)
		setNeeds(s, content.Temple, 30)
	}
	picks := map[content.DistrictID]int{}
	for seed := uint64(1); seed <= 40; seed++ {
		a, b := toDusk(t, seed, tie), toDusk(t, seed, tie)
		if a.Dusk.Turned != b.Dusk.Turned {
			t.Fatalf("seed %d broke the tie two ways", seed)
		}
		picks[a.Dusk.Turned]++
	}
	if picks[content.Market] == 0 || picks[content.Temple] == 0 || len(picks) != 2 {
		t.Errorf("40 seeds picked %v; want both tied districts and no other", picks)
	}
}

func TestTurningRecallsStewards(t *testing.T) {
	s := toDusk(t, 1, func(s *State) {
		setNeeds(s, content.Docks, 20)
		s.Stewards[1] = Post{District: content.Docks, Need: content.NoNeed, Building: -1}
		s.Districts[content.Docks].Needs[content.Supply] -= content.T.StewardNeed // the steward's visit is not the point here
	})
	if s.Dusk.Turned != content.Docks || s.Stewards[1].District != content.NoDistrict {
		t.Errorf("turned %s, steward at %s", s.Dusk.Turned, s.Stewards[1].District)
	}
}

// docksHollow is a dusk with the Docks turned in a known state.
func docksHollow(t testing.TB, needs [3]int, conds ...int) State {
	return toDusk(t, 1, func(s *State) {
		d := &s.Districts[content.Docks]
		d.Needs = needs
		for i, c := range conds {
			d.Buildings[i].Condition = c
		}
	})
}

func TestSendPartyRules(t *testing.T) {
	s := docksHollow(t, [3]int{20, 60, 60})
	s.Citizens[5].Alive = false
	tests := []struct {
		name string
		c    Command
		want error
	}{
		{"no such district", SendParty(9, 0), ErrBadArgument},
		{"no hollow there", SendParty(content.Market, 0), ErrBadArgument},
		{"nobody", SendParty(content.Docks), ErrBadArgument},
		{"five is too many", SendParty(content.Docks, 0, 1, 2, 3, 4), ErrBadArgument},
		{"the dead do not go", SendParty(content.Docks, 5), ErrBadArgument},
		{"nobody goes twice", SendParty(content.Docks, 1, 1), ErrBadArgument},
		{"no such citizen", SendParty(content.Docks, 99), ErrBadArgument},
		{"cannot advance at dusk", Advance(1), ErrWrongPhase},
		{"cannot act at dusk", Act(ActAttack, 0), ErrWrongPhase},
		{"cannot end the day at dusk", EndDay(), ErrWrongPhase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { mustFail(t, s, tt.c, tt.want) })
	}
	mustFail(t, start(t, 1), SendParty(content.Docks, 0), ErrWrongPhase)
	mustFail(t, start(t, 1), StayHome(), ErrWrongPhase)
	for size := 1; size <= content.T.MaxParty; size++ {
		n, _ := do(t, s, SendParty(content.Docks, []int{0, 1, 2, 3}[:size]...))
		if n.Phase != PhaseNight || len(n.Expedition.Party) != size {
			t.Errorf("a party of %d did not set out", size)
		}
	}
}

func TestStayingHomeLeavesTheDistrictTurned(t *testing.T) {
	s := docksHollow(t, [3]int{20, 60, 60})
	frozen := s.Districts[content.Docks]
	s, ev := do(t, s, StayHome())
	if s.Phase != PhaseDay || s.Day != 2 {
		t.Fatalf("morning did not come: day %d phase %q", s.Day, s.Phase)
	}
	if !s.Districts[content.Docks].Turned || s.Hollows[content.Docks] == nil {
		t.Fatal("the Docks came back by themselves")
	}
	if ev[0].Kind != EvStayedHome {
		t.Errorf("events %+v", ev)
	}
	// Turned: output 0, frozen, out of a steward's reach.
	if out := s.Output(content.Docks); out != (content.Cost{}) {
		t.Errorf("a turned district produced %v", out)
	}
	s = pass(t, s)
	if got := s.Districts[content.Docks]; encode(t, got) != encode(t, frozen) {
		t.Errorf("a turned district changed overnight:\n%+v\n%+v", frozen, got)
	}
	mustFail(t, s, AssignSteward(0, content.Docks, content.NoNeed, -1), ErrTurned)
	// The hollow still points at real state, because that state is frozen.
	if msg := checkHollow(&s.Districts[content.Docks], s.Hollows[content.Docks]); msg != "" {
		t.Error(msg)
	}
}

// walk sends the given citizens in and plays every fight with plain attacks,
// always taking the first exit, until the night ends or stop says so.
func walk(t testing.TB, s State, d content.DistrictID, stop func(s *State) bool, citizens ...int) (State, []Event) {
	t.Helper()
	s, events := do(t, s, SendParty(d, citizens...))
	for steps := 0; s.Phase == PhaseNight; steps++ {
		if steps > 5000 {
			t.Fatal("the night does not end")
		}
		if stop != nil && stop(&s) {
			break
		}
		var ev []Event
		if s.Expedition.Fight != nil {
			r := s.room()
			target := 0
			for i := range r.Enemies {
				if r.Enemies[i].HP > 0 {
					target = i
					break
				}
			}
			s, ev = do(t, s, Act(ActAttack, target))
		} else {
			s, ev = do(t, s, Advance(s.room().Exits[0]))
		}
		events = append(events, ev...)
	}
	return s, events
}

func count(events []Event, kind EventKind) int {
	n := 0
	for _, e := range events {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

// mighty makes citizens that cannot lose, for tests about what winning does.
func mighty(s *State, ids ...int) {
	for _, id := range ids {
		s.Citizens[id].JitterHP = 500
		s.Citizens[id].HP = s.Citizens[id].Stats().HP
	}
}

func TestClearingTheHeartRestoresTheDistrict(t *testing.T) {
	tn := content.T
	s := docksHollow(t, [3]int{10, 30, 70}, 50, 50, 50)
	mighty(&s, 0, 1, 2, 3)
	coin, bone := s.Resources[content.Coin], s.Materials[content.Hunger]
	hunger := 0
	for _, r := range s.Hollows[content.Docks].Rooms {
		for _, e := range r.Enemies {
			if !e.Boss && content.Enemies[e.Kind].Family == content.Hunger {
				hunger++
			}
		}
	}
	scavengers := 0
	for _, id := range []int{0, 1, 2, 3} {
		if s.Citizens[id].HasTrait(func(tr content.TraitDef) bool { return tr.Scavenger }) {
			scavengers++
		}
	}
	s, ev := walk(t, s, content.Docks, nil, 0, 1, 2, 3)
	d := &s.Districts[content.Docks]
	if d.Turned || s.Hollows[content.Docks] != nil {
		t.Fatal("the Docks are still turned after the heart was cleared")
	}
	if d.Needs != [3]int{tn.ClearRestoreNeed, tn.ClearRestoreNeed, 70} {
		t.Errorf("needs %v: want the low ones lifted to %d and the high one left alone", d.Needs, tn.ClearRestoreNeed)
	}
	if s.Phase != PhaseDay || s.Day != 2 || s.Expedition != nil {
		t.Errorf("the night did not end: day %d phase %q", s.Day, s.Phase)
	}
	if got := s.Resources[content.Coin] - coin; got != tn.HeartCoin {
		t.Errorf("heart dropped %d coin, want %d", got, tn.HeartCoin)
	}
	// The first exit at every fork misses some rooms, so count what fell.
	fell := 0
	for _, e := range ev {
		if e.Kind == EvEnemyFell {
			for _, def := range content.Enemies {
				if def.Slug == e.ID && def.Family == content.Hunger {
					fell++
				}
			}
		}
	}
	if fell == 0 || fell > hunger {
		t.Fatalf("%d Hunger enemies fell of %d", fell, hunger)
	}
	got := s.Materials[content.Hunger] - bone
	if lo, hi := fell+tn.HeartMaterials, fell+tn.HeartMaterials+scavengers*10; got < lo || got > hi {
		t.Errorf("gained %d hungerbone; want %d for the fallen and the heart (plus scavenging)", got, lo)
	}
	if count(ev, EvHeartCleared) != 1 || count(ev, EvRestored) != 1 {
		t.Error("the log lacks the heart or the restoration")
	}
}

func TestGoingHomeKeepsProgress(t *testing.T) {
	s := docksHollow(t, [3]int{10, 60, 60}, 10, 50, 50) // Hunger, and a ruined warehouse
	mighty(&s, 0, 1)
	// Fight through two rooms, then go home.
	s, _ = walk(t, s, content.Docks, func(s *State) bool { return s.Expedition.Fight == nil && s.room().Depth >= 2 }, 0, 1)
	if s.Phase != PhaseNight {
		t.Fatal("the night ended early")
	}
	mustFail(t, s, Advance(0), ErrBadArgument) // not an exit of this room
	s, ev := do(t, s, GoHome())
	if count(ev, EvWentHome) != 1 || s.Phase != PhaseDay {
		t.Fatalf("going home did not end the night: %+v", ev)
	}
	h := s.Hollows[content.Docks]
	if h == nil || !s.Districts[content.Docks].Turned {
		t.Fatal("going home cleared the district")
	}
	cleared := 0
	for _, r := range h.Rooms {
		if r.Cleared {
			cleared++
			if r.Alive() != 0 {
				t.Errorf("cleared room %q has living enemies", r.Name)
			}
		}
	}
	if cleared < 2 {
		t.Fatalf("%d rooms cleared, want the ones the party fought through", cleared)
	}
	// The next night the cleared rooms are quiet: no fights, no second hazard.
	s = answer(t, s)
	s, _ = do(t, s, EndDay())
	before := s.Citizens[0].HP
	s, ev = do(t, s, SendParty(content.Docks, 0, 1))
	for s.Phase == PhaseNight && s.Expedition.Fight == nil && s.room().Cleared && len(s.room().Exits) > 0 {
		next := s.room().Exits[0]
		if !h.Rooms[next].Cleared {
			break
		}
		var more []Event
		s, more = do(t, s, Advance(next))
		ev = append(ev, more...)
	}
	if count(ev, EvFightStarted) != 0 || count(ev, EvHazard) != 0 || s.Citizens[0].HP != before {
		t.Errorf("cleared rooms were not quiet on the way back in: %d fights, %d hazards", count(ev, EvFightStarted), count(ev, EvHazard))
	}
}

func TestHazardAndHavenRooms(t *testing.T) {
	// The Docks' warehouse follows Salt Pier, room 1.
	t.Run("a hazard costs everyone HP, once", func(t *testing.T) {
		s := docksHollow(t, [3]int{60, 60, 60}, 5, 50, 50)
		h := s.Hollows[content.Docks]
		var hazard *Room
		for i := range h.Rooms {
			if h.Rooms[i].Kind == RoomHazard {
				hazard = &h.Rooms[i]
			}
		}
		want := content.T.HazardDamageBase + (content.T.RuinThreshold-5)/content.T.HazardDamagePer
		if hazard == nil || hazard.Hazard != want {
			t.Fatalf("hazard %+v, want damage %d", hazard, want)
		}
		full0, full1 := s.Citizens[0].HP, s.Citizens[1].HP
		s, _ = do(t, s, SendParty(content.Docks, 0, 1))
		s, _ = do(t, s, Advance(1)) // Salt Pier, empty
		s, ev := do(t, s, Advance(hazard.ID))
		if s.Citizens[0].HP != full0-want || s.Citizens[1].HP != full1-want || count(ev, EvHazard) != 1 {
			t.Errorf("HP %d and %d, want %d and %d", s.Citizens[0].HP, s.Citizens[1].HP, full0-want, full1-want)
		}
		if ev[len(ev)-1].Causes[0].ID != "docks/warehouse" || ev[len(ev)-1].Causes[0].ValueAtGeneration != 5 {
			t.Errorf("the hazard does not say why: %+v", ev[len(ev)-1])
		}
		if s.Citizens[2].HP != s.Citizens[2].Stats().HP {
			t.Error("the hazard hurt someone who stayed home")
		}
	})
	t.Run("a hazard can down a citizen but HP stops at zero", func(t *testing.T) {
		s := docksHollow(t, [3]int{60, 60, 60}, 5, 50, 50)
		s.Citizens[0].HP = 1
		s, _ = do(t, s, SendParty(content.Docks, 0, 1))
		s, _ = do(t, s, Advance(1))
		s, ev := do(t, s, Advance(s.room().Exits[0]))
		if c := s.Citizens[0]; c.HP != 0 || !c.Downed || count(ev, EvDowned) != 1 {
			t.Errorf("citizen 0: HP %d downed %v", c.HP, c.Downed)
		}
	})
	t.Run("a safe room heals and holds no enemies", func(t *testing.T) {
		s := docksHollow(t, [3]int{60, 60, 60}, 95, 50, 50)
		s, _ = do(t, s, SendParty(content.Docks, 0, 1))
		s, _ = do(t, s, Advance(1))
		s.Citizens[0].HP, s.Citizens[1].HP = 1, s.Citizens[1].Stats().HP-1
		top := s.Citizens[0].Stats().HP
		s, ev := do(t, s, Advance(s.room().Exits[0]))
		if s.room().Kind != RoomHaven || len(s.room().Enemies) != 0 || count(ev, EvFightStarted) != 0 {
			t.Fatalf("not a safe room: %+v", s.room())
		}
		heal := (top*content.T.HavenHealPct + 99) / 100
		if s.Citizens[0].HP != 1+heal || s.Citizens[1].HP != s.Citizens[1].Stats().HP {
			t.Errorf("HP %d and %d after resting; want %d and full", s.Citizens[0].HP, s.Citizens[1].HP, 1+heal)
		}
	})
}

func TestDownedCitizensDieOrScar(t *testing.T) {
	tn := content.T
	run := func(lucky bool) (died, scarred int) {
		for seed := uint64(1); seed <= 2000; seed++ {
			s := State{Phase: PhaseNight, Day: 1, RNG: newStreams(seed)}
			s.Citizens = []Citizen{{ID: 0, Name: "Hale", Job: content.Smith, Alive: true, Downed: true, Traits: [2]content.TraitID{content.Wary, content.Strong}}}
			if lucky {
				s.Citizens[0].Traits[0] = content.Lucky
			}
			s.Expedition = &Expedition{District: content.Docks, Party: []Member{{Citizen: 0, Fought: true}}}
			l := &log{}
			s.endNight(l)
			c := s.Citizens[0]
			switch {
			case !c.Alive:
				died++
				if len(c.Scars) != 0 || count(l.events, EvDied) != 1 {
					t.Fatalf("seed %d: a death with scars or without a log line", seed)
				}
			case len(c.Scars) == 1 && !c.Downed && c.HP == c.Stats().HP && count(l.events, EvScarred) == 1:
				scarred++
				if m := content.Scars[c.Scars[0]].Mods(); m == (content.Mods{}) {
					t.Fatalf("seed %d: a scar that costs nothing", seed)
				}
			default:
				t.Fatalf("seed %d: a Downed citizen neither died nor was scarred: %+v", seed, c)
			}
		}
		return died, scarred
	}
	died, scarred := run(false)
	if pct := died * 100 / 2000; pct < tn.DownedDeathPct-4 || pct > tn.DownedDeathPct+4 {
		t.Errorf("%d of 2000 Downed citizens died (%d%%), want about %d%%", died, pct, tn.DownedDeathPct)
	}
	if died+scarred != 2000 {
		t.Error("some survivors carry no scar")
	}
	luckyDied, _ := run(true)
	if pct := luckyDied * 100 / 2000; pct < tn.LuckyDeathPct-4 || pct > tn.LuckyDeathPct+4 {
		t.Errorf("%d of 2000 Lucky citizens died (%d%%), want about %d%%", luckyDied, pct, tn.LuckyDeathPct)
	}
}

func TestTheDeadStayDead(t *testing.T) {
	s := docksHollow(t, [3]int{20, 60, 60})
	s.Citizens[0].Alive = false
	s = sleep(t, s)
	for range 3 {
		s = pass(t, s)
	}
	if s.Citizens[0].Alive || s.Citizens[0].WorkPct() != 0 {
		t.Error("a dead citizen came back")
	}
}

func TestExhaustion(t *testing.T) {
	s := docksHollow(t, [3]int{10, 60, 60})
	mighty(&s, 0, 1)
	for i := range s.Citizens {
		s.Citizens[i].Traits = [2]content.TraitID{content.Wary, content.Strong}
	}
	// 0 and 1 fight one room and go home; everyone else stays in.
	s, _ = walk(t, s, content.Docks, func(s *State) bool { return s.Expedition.Fight == nil && s.room().Cleared && s.room().Depth >= 1 }, 0, 1)
	if s.Phase != PhaseNight || !s.Expedition.Party[0].Fought {
		t.Fatal("the party never fought")
	}
	s, ev := do(t, s, GoHome())
	if count(ev, EvExhausted) != 2 {
		t.Errorf("%d citizens exhausted, want the two who fought", count(ev, EvExhausted))
	}
	for i := range s.Citizens {
		c := &s.Citizens[i]
		want, pct := i <= 1, 100
		if want {
			pct = content.T.ExhaustedWorkPct
		}
		if c.Exhausted != want || c.WorkPct() != pct {
			t.Errorf("%s: exhausted %v, works at %d%%; want %v and %d%%", c.Name, c.Exhausted, c.WorkPct(), want, pct)
		}
	}
	// It lasts one day.
	s = answer(t, s)
	s, _ = do(t, s, EndDay())
	for i := range s.Citizens {
		if s.Citizens[i].Exhausted {
			t.Errorf("%s is still exhausted a day later", s.Citizens[i].Name)
		}
	}

	t.Run("walking through empty rooms tires nobody", func(t *testing.T) {
		s := docksHollow(t, [3]int{60, 60, 60})
		s, _ = do(t, s, SendParty(content.Docks, 0, 1))
		s, _ = do(t, s, Advance(1))
		s, ev := do(t, s, GoHome())
		if count(ev, EvExhausted) != 0 || s.Citizens[0].Exhausted {
			t.Error("citizens who never fought are exhausted")
		}
	})
}
