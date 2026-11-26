package sim

import (
	"errors"
	"strings"
	"testing"

	"hollowward/internal/content"
)

func TestDreamerRises(t *testing.T) {
	tn := content.T
	t.Run("a quiet night", func(t *testing.T) {
		s := toDusk(t, 1, func(s *State) {})
		if s.Dreamer != tn.DreamerQuiet || len(s.Dusk.Rise) != 1 || s.Dusk.Rise[0].Reason != RiseQuiet {
			t.Errorf("Dreamer %d after a quiet night, want %d; rise %+v", s.Dreamer, tn.DreamerQuiet, s.Dusk.Rise)
		}
	})
	t.Run("a turning adds its Neglect over ten", func(t *testing.T) {
		s := toDusk(t, 1, func(s *State) { setNeeds(s, content.Docks, 30) }) // Neglect 70
		if s.Dreamer != 70*10/tn.DreamerNeglectDiv || s.Dusk.Rise[0].Reason != RiseTurned || s.Dusk.Rise[0].District != content.Docks {
			t.Errorf("Dreamer %d tenths after Neglect 70; rise %+v", s.Dreamer, s.Dusk.Rise)
		}
		if s.Dusk.Before != 0 || s.Dusk.After != s.Dreamer || s.Dusk.Line == "" {
			t.Errorf("dusk report %+v", s.Dusk)
		}
	})
	t.Run("every district still turned adds fifteen", func(t *testing.T) {
		s := toDusk(t, 1, func(s *State) { setNeeds(s, content.Docks, 30) })
		s = sleep(t, s)
		first := s.Dreamer
		// Night 2: the Mills turn as well; the Docks are still open.
		s = answer(t, s)
		setNeeds(&s, content.Mills, 30+tn.NeedBaseDecay+3)
		s.Districts[content.Mills].Needs = [3]int{20, 20, 20}
		s, _ = do(t, s, EndDay())
		neglect := s.Dusk.Neglect
		want := first + TurnRise(neglect) + tn.DreamerPerOpenHollow
		if s.Dusk.Turned != content.Mills || s.Dreamer != want {
			t.Fatalf("turned %s, Dreamer %d, want %d: %+v", s.Dusk.Turned, s.Dreamer, want, s.Dusk.Rise)
		}
		// Night 3: quiet, with two open hollows.
		s = sleep(t, s)
		s = answer(t, s)
		calm(&s)
		s, _ = do(t, s, EndDay())
		want += tn.DreamerQuiet + 2*tn.DreamerPerOpenHollow
		if s.Dusk.Turned != content.NoDistrict || s.Dreamer != want || len(s.Dusk.Rise) != 3 {
			t.Errorf("Dreamer %d, want %d: %+v", s.Dreamer, want, s.Dusk.Rise)
		}
	})
	t.Run("the meter never falls on its own, and clearing a heart soothes by the knob", func(t *testing.T) {
		s := docksHollow(t, [3]int{10, 60, 60})
		mighty(&s, 0, 1, 2, 3)
		before := s.Dreamer
		s, _ = walk(t, s, content.Docks, nil, 0, 1, 2, 3)
		if s.Dreamer != before-tn.DreamerSootheOnClear {
			t.Errorf("Dreamer %d after clearing a heart, want %d", s.Dreamer, before-tn.DreamerSootheOnClear)
		}
	})
}

func TestTheLidBreaks(t *testing.T) {
	tn := content.T
	s := start(t, 1)
	s = answer(t, s)
	calm(&s)
	s.Dreamer = tn.DreamerMax - tn.DreamerQuiet // one quiet night from the end
	s.Nights, s.Day = 6, 7
	s, ev := do(t, s, EndDay())
	if s.Phase != PhaseOver || s.Outcome != LostLid || s.Dreamer != tn.DreamerMax {
		t.Fatalf("phase %q outcome %q Dreamer %d", s.Phase, s.Outcome, s.Dreamer)
	}
	last := ev[len(ev)-1]
	if last.Kind != EvTenureEnded || last.ID != string(LostLid) || last.N != 6 {
		t.Errorf("last event %+v", last)
	}
	// Nothing more can be done but start again.
	for _, c := range []Command{EndDay(), StayHome(), SendParty(content.Docks, 0), GoHome(), AssignSteward(0, content.Market, content.NoNeed, -1)} {
		mustFail(t, s, c, ErrWrongPhase)
	}
	next, _ := do(t, s, NewTenure(99))
	if next.Phase != PhaseDay || next.Day != 1 || next.Dreamer != 0 || next.Seed != 99 || next.Outcome != Ongoing {
		t.Errorf("a new tenure did not start clean: %+v", next)
	}
}

func TestSurvivingTheLastNightWins(t *testing.T) {
	tn := content.T
	s := start(t, 1)
	s = answer(t, s)
	calm(&s)
	s.Day, s.Nights = tn.Nights, tn.Nights-1
	s, _ = do(t, s, EndDay())
	if s.Phase != PhaseDusk {
		t.Fatalf("the last dusk must still fall: phase %q", s.Phase)
	}
	s, ev := do(t, s, StayHome())
	if s.Phase != PhaseOver || s.Outcome != Won || s.Nights != tn.Nights {
		t.Fatalf("phase %q outcome %q nights %d", s.Phase, s.Outcome, s.Nights)
	}
	if last := ev[len(ev)-1]; last.Kind != EvTenureEnded || last.N != tn.Nights {
		t.Errorf("last event %+v", last)
	}

	t.Run("the lid breaking on the last dusk is still a loss", func(t *testing.T) {
		s := start(t, 1)
		s = answer(t, s)
		calm(&s)
		s.Day, s.Nights, s.Dreamer = tn.Nights, tn.Nights-1, tn.DreamerMax-1
		s, _ = do(t, s, EndDay())
		if s.Outcome != LostLid {
			t.Errorf("outcome %q", s.Outcome)
		}
	})
}

func TestNoOneLeftToSend(t *testing.T) {
	s := docksHollow(t, [3]int{20, 60, 60})
	for i := range s.Citizens {
		s.Citizens[i].Alive = false
	}
	mustFail(t, s, SendParty(content.Docks, 0), ErrBadArgument)
	s, ev := do(t, s, StayHome())
	if s.Phase != PhaseOver || s.Outcome != LostNoParty {
		t.Fatalf("phase %q outcome %q", s.Phase, s.Outcome)
	}
	if last := ev[len(ev)-1]; last.Kind != EvTenureEnded || last.N != 1 {
		t.Errorf("last event %+v", last)
	}
}

func TestUpgrades(t *testing.T) {
	tn := content.T
	rich := content.Materials{50, 50, 50}
	t.Run("costs are paid in materials", func(t *testing.T) {
		for id := content.UpgradeID(0); id < content.NumUpgrades; id++ {
			s := start(t, 2)
			s.Materials = rich
			s, ev := do(t, s, BuyUpgrade(id, content.Market, 0))
			cost := content.Upgrades[id].Cost()
			for f := range cost {
				if s.Materials[f] != rich[f]-cost[f] {
					t.Errorf("%s: paid %d %s, want %d", content.Upgrades[id].Name, rich[f]-s.Materials[f], content.Family(f).Material(), cost[f])
				}
			}
			if len(ev) != 1 || ev[0].Kind != EvUpgradeBought || ev[0].ID != content.Upgrades[id].Slug {
				t.Errorf("events %+v", ev)
			}
			if cost == (content.Materials{}) {
				t.Errorf("%s is free", content.Upgrades[id].Name)
			}
		}
	})
	t.Run("some upgrades take a single material", func(t *testing.T) {
		single := 0
		for _, u := range content.Upgrades {
			kinds := 0
			for _, v := range u.Cost() {
				if v > 0 {
					kinds++
				}
			}
			if kinds == 1 {
				single++
			}
		}
		if single < 3 {
			t.Errorf("%d single-material upgrades; farming one district must be able to pay for something", single)
		}
	})
	t.Run("a ward halves a building's decay", func(t *testing.T) {
		s := start(t, 2)
		flat(&s, 80, 80)
		s.Materials = rich
		s, _ = do(t, s, BuyUpgrade(content.Ward, content.Market, 1))
		mustFail(t, s, BuyUpgrade(content.Ward, content.Market, 1), ErrAlreadyDone)
		s, _ = do(t, s, BuyUpgrade(content.Ward, content.Docks, 0)) // a second ward is fine
		s.decay(&log{})
		m := s.Districts[content.Market].Buildings
		if m[1].Condition != 80-tn.BuildingDecay/tn.WardDivisor || m[0].Condition != 80-tn.BuildingDecay {
			t.Errorf("warded building at %d, unwarded at %d", m[1].Condition, m[0].Condition)
		}
	})
	t.Run("gear changes every citizen", func(t *testing.T) {
		s := start(t, 2)
		s.Materials = rich
		before := make([]content.Stats, len(s.Citizens))
		for i := range s.Citizens {
			before[i] = s.Citizens[i].Stats()
		}
		for _, id := range []content.UpgradeID{content.BoneCharms, content.AshHoods, content.BoneBlades} {
			s, _ = do(t, s, BuyUpgrade(id, content.NoDistrict, -1))
			mustFail(t, s, BuyUpgrade(id, content.NoDistrict, -1), ErrAlreadyDone)
		}
		for i := range s.Citizens {
			c := &s.Citizens[i]
			got, was := c.Stats(), before[i]
			if got.HP != was.HP+tn.GearCharms.HP || got.Guard != was.Guard+tn.GearHoods.Guard || got.Attack != was.Attack+tn.GearBlades.Attack || c.HP != got.HP {
				t.Errorf("%s: %+v, was %+v", c.Name, got, was)
			}
		}
	})
	t.Run("seals, bells and mortar slow one need's decay", func(t *testing.T) {
		for id, need := range map[content.UpgradeID]content.Need{content.GranarySeal: content.Supply, content.CurfewBells: content.Order, content.BlackMortar: content.Upkeep} {
			s := start(t, 2)
			s.Materials = rich
			was := s.NeedDecay(content.Docks, need)
			s, _ = do(t, s, BuyUpgrade(id, content.NoDistrict, -1))
			for n := content.Need(0); n < content.NumNeeds; n++ {
				want := tn.NeedBaseDecay + tn.NeedDecayAdj[content.Docks][n]
				if n == need {
					want = was - tn.DecayEase
				}
				if got := s.NeedDecay(content.Docks, n); got != want {
					t.Errorf("%s: %s decays by %d, want %d", content.Upgrades[id].Name, n, got, want)
				}
			}
		}
	})
	t.Run("refusals", func(t *testing.T) {
		s := start(t, 2)
		mustFail(t, s, BuyUpgrade(content.BoneCharms, content.NoDistrict, -1), ErrCannotAfford)
		s.Materials = rich
		mustFail(t, s, BuyUpgrade(content.NumUpgrades, content.NoDistrict, -1), ErrBadArgument)
		mustFail(t, s, BuyUpgrade(content.Ward, content.NoDistrict, 0), ErrBadArgument)
		mustFail(t, s, BuyUpgrade(content.Ward, content.Watch, 5), ErrBadArgument)
		s.Districts[content.Watch].Turned = true
		mustFail(t, s, BuyUpgrade(content.Ward, content.Watch, 0), ErrTurned)
	})
}

// nightEvents plays one known night and returns its events: a party walks the
// Docks' hollow to the heart and clears it.
func nightEvents(t testing.TB, needs [3]int, conds ...int) (State, []Event) {
	t.Helper()
	s := docksHollow(t, needs, conds...)
	mighty(&s, 0, 1, 2, 3)
	return walk(t, s, content.Docks, nil, 0, 1, 2, 3)
}

func TestLedgerRecordsWhatWasMet(t *testing.T) {
	_, ev := nightEvents(t, [3]int{10, 60, 60}, 5, 50, 50)
	l := Ledger{}.Record(ev)
	if !l.KnowsFamily(content.Hunger) || l.KnowsFamily(content.Riot) || l.KnowsFamily(content.Rot) {
		t.Errorf("families on record: hunger %v riot %v rot %v", l.KnowsFamily(content.Hunger), l.KnowsFamily(content.Riot), l.KnowsFamily(content.Rot))
	}
	if !l.Knows(LedgerBoss, "docks-boss/supply") {
		t.Error("the boss is not on record")
	}
	for _, e := range l.Entries {
		if e.Seen < 1 || e.Tenure != 1 || e.District != content.Docks {
			t.Errorf("entry %+v", e)
		}
		// Every entry carries the cause that produced it.
		switch e.Kind {
		case LedgerEnemy:
			if e.Cause.Source != FromNeed || e.Cause.ID != "docks/supply" || e.Cause.ValueAtGeneration != 10 {
				t.Errorf("%s: cause %+v", e.ID, e.Cause)
			}
		case LedgerBoss:
			if e.Cause.Source != FromFigure {
				t.Errorf("%s: cause %+v", e.ID, e.Cause)
			}
		case LedgerHazard:
			if e.Cause.Source != FromBuilding || e.Cause.ValueAtGeneration != 5 {
				t.Errorf("%s: cause %+v", e.ID, e.Cause)
			}
		}
	}
	// Meeting things again counts them; it does not list them twice.
	again := l.Record(ev)
	if len(again.Entries) != len(l.Entries) || again.Entries[0].Seen != 2*l.Entries[0].Seen {
		t.Errorf("a second meeting: %d entries (was %d), seen %d", len(again.Entries), len(l.Entries), again.Entries[0].Seen)
	}
	if len(l.Entries) == 0 || encode(t, Ledger{}.Record(ev)) != encode(t, l) {
		t.Error("recording is not repeatable")
	}
}

func TestLedgerRecordsHazardsFragmentsAndTenures(t *testing.T) {
	// Walk the first exit: Salt Pier, then the warehouse (ruined) or its haven.
	_, ev := nightEvents(t, [3]int{60, 60, 60}, 5, 50, 50)
	l := Ledger{}.Record(ev)
	if !l.Knows(LedgerHazard, "docks/warehouse") {
		t.Errorf("the hazard is not on record: %+v", l.Entries)
	}
	_, ev = nightEvents(t, [3]int{60, 60, 60}, 95, 50, 50)
	l = l.Record(ev)
	if len(l.Fragments) != 1 || l.Fragments[0] != 0 || l.HasAllFragments() {
		t.Errorf("fragments %v after one safe room", l.Fragments)
	}
	l = l.Record(ev) // the same page twice is one page
	if len(l.Fragments) != 1 {
		t.Errorf("fragments %v", l.Fragments)
	}
	for n := 1; n < len(content.Fragments); n++ {
		l = l.Record([]Event{{Kind: EvFragment, N: n}})
	}
	if !l.HasAllFragments() {
		t.Error("all five pages found, but the closing text stays locked")
	}
	l = l.Record([]Event{{Kind: EvTenureEnded, ID: string(Won), N: 20}, {Kind: EvTenureEnded, ID: string(LostLid), N: 4}})
	if l.Tenures != 2 || l.Won != 1 {
		t.Errorf("tenures %d won %d", l.Tenures, l.Won)
	}
	l = l.Record([]Event{{Kind: EvMet, ID: "gaunt", N: 2, District: content.Market, Causes: []Cause{{FromNeed, "market/supply", 12}}}})
	if e := l.Entries[len(l.Entries)-1]; e.Tenure != 3 {
		t.Errorf("an entry from the third tenure is dated %d", e.Tenure)
	}
}

func TestSafeRoomsGiveFragmentsOnce(t *testing.T) {
	s := docksHollow(t, [3]int{60, 60, 60}, 95, 50, 50)
	s, _ = do(t, s, SendParty(content.Docks, 0))
	s, _ = do(t, s, Advance(1))
	s, ev := do(t, s, Advance(s.room().Exits[0]))
	if count(ev, EvFragment) != 1 || s.Fragments != 1 || !s.room().Searched {
		t.Fatalf("fragments %d", s.Fragments)
	}
	s, _ = do(t, s, GoHome())
	s = answer(t, s)
	calm(&s)
	s.Dreamer = 0
	s, _ = do(t, s, EndDay())
	s, _ = do(t, s, SendParty(content.Docks, 0))
	s, _ = do(t, s, Advance(1))
	s, ev = do(t, s, Advance(s.room().Exits[0]))
	if count(ev, EvFragment) != 0 || s.Fragments != 1 {
		t.Error("the same safe room gave a second page")
	}
}

func TestDuskPreview(t *testing.T) {
	s := docksHollow(t, [3]int{10, 60, 60}, 5, 95, 50)
	h := s.Hollows[content.Docks]
	// An empty Ledger knows nothing: unknowns, each with its cause.
	blank := Preview(&Ledger{}, h)
	text := func(lines []PreviewLine) string {
		var b strings.Builder
		for _, p := range lines {
			b.WriteString(p.String() + "\n")
		}
		return b.String()
	}
	for _, p := range blank {
		if p.Cause.ID == "" {
			t.Errorf("preview line %q has no cause", p.Text)
		}
		if v, ok := p.Cause.Resolve(&s.Districts[content.Docks]); !ok || v != p.Cause.ValueAtGeneration {
			t.Errorf("preview line %q: cause %+v is not real", p.Text, p.Cause)
		}
	}
	bt := text(blank)
	for _, hidden := range []string{"Gaunt", "Gnawer", "Hollow-Belly", "The Starved Harbourmaster", "The Flooded Warehouse"} {
		if strings.Contains(bt, hidden) {
			t.Errorf("an empty Ledger previewed %q:\n%s", hidden, bt)
		}
	}
	for _, want := range []string{"Something not in the Ledger", "Docks Supply was 10", "A hazard not in the Ledger", "someone not in the Ledger", "Safe room: The Crane Yard"} {
		if !strings.Contains(bt, want) {
			t.Errorf("empty-Ledger preview lacks %q:\n%s", want, bt)
		}
	}
	// After a night in there, the Ledger previews what it recorded.
	mighty(&s, 0, 1, 2, 3)
	_, ev := walk(t, s, content.Docks, nil, 0, 1, 2, 3)
	l := Ledger{}.Record(ev)
	s2 := docksHollow(t, [3]int{10, 60, 60}, 5, 95, 50)
	kt := text(Preview(&l, s2.Hollows[content.Docks]))
	for _, want := range []string{"Docks Supply was 10", "At the heart: The Starved Harbourmaster", "Hazard: The Flooded Warehouse"} {
		if !strings.Contains(kt, want) {
			t.Errorf("preview lacks %q:\n%s", want, kt)
		}
	}
	if !strings.Contains(kt, "Gaunt") && !strings.Contains(kt, "Gnawer") && !strings.Contains(kt, "Hollow-Bell") {
		t.Errorf("the Ledger met Hunger but previews none of it:\n%s", kt)
	}
	// The preview is a reading: it changes neither the Ledger nor the hollow.
	before, lb := encode(t, s2), encode(t, l)
	Preview(&l, s2.Hollows[content.Docks])
	if encode(t, s2) != before || encode(t, l) != lb {
		t.Error("previewing changed something")
	}
}

func TestChronicle(t *testing.T) {
	g, err := NewGame(4)
	if err != nil {
		t.Fatal(err)
	}
	if lines := Chronicle(g.Log); len(lines) != 1 || lines[0] != "A new warden took the keys." {
		t.Errorf("a fresh tenure's chronicle: %v", lines)
	}
	// Three quiet nights, then a turning the warden ignores until the lid breaks.
	events := []Event{
		{Kind: EvTenureStarted, Day: 1}, {Kind: EvDayStarted, Day: 1}, {Kind: EvQuietNight, Day: 1},
		{Kind: EvDayStarted, Day: 2}, {Kind: EvQuietNight, Day: 2},
		{Kind: EvDayStarted, Day: 3}, {Kind: EvRuined, Day: 3, District: content.Market, ID: "granary", N: 28},
		{Kind: EvTurned, Day: 3, District: content.Market, N: 61, Causes: []Cause{{FromNeed, "market/supply", 18}}},
		{Kind: EvJoined, Day: 3, Who: "Edda"}, {Kind: EvJoined, Day: 3, Who: "Hale"}, {Kind: EvJoined, Day: 3, Who: "Sten"},
		{Kind: EvHeartCleared, Day: 3, District: content.Market, ID: "market-boss/supply"},
		{Kind: EvDied, Day: 3, Who: "Hale"}, {Kind: EvScarred, Day: 3, Who: "Sten", ID: "limp"},
		{Kind: EvDayStarted, Day: 4}, {Kind: EvQuietNight, Day: 4},
		{Kind: EvDayStarted, Day: 5}, {Kind: EvQuietNight, Day: 5},
		{Kind: EvTenureEnded, Day: 5, ID: string(LostLid), N: 4},
	}
	got := strings.Join(Chronicle(events), "\n")
	want := strings.Join([]string{
		"A new warden took the keys.",
		"Nights 1 to 2 were quiet.",
		"Night 3. The Granary in the Market fell into ruin. The Market turned, at Neglect 61: Market Supply was 18. Edda, Hale and Sten went in and put down The Starved Grain Factor; the Market was a district again. Hale did not wake. Sten came back marked: Limp.",
		"Night 4 was quiet.",
		"On night 5 the lid broke. The warden had kept it 4 nights.",
	}, "\n")
	if got != want {
		t.Errorf("chronicle:\n%s\nwant:\n%s", got, want)
	}
}

func TestSaveAndLoad(t *testing.T) {
	g, err := NewGame(31)
	if err != nil {
		t.Fatal(err)
	}
	// Play into a night, so the save holds a hollow, a party and a fight.
	g.State = answer(t, g.State)
	setNeeds(&g.State, content.Docks, 10)
	for _, c := range []Command{EndDay(), SendParty(content.Docks, 0, 1, 2)} {
		if _, err := g.Do(c); err != nil {
			t.Fatal(err)
		}
	}
	for g.State.Expedition != nil && g.State.Expedition.Fight == nil {
		if _, err := g.Do(Advance(g.State.room().Exits[0])); err != nil {
			t.Fatal(err)
		}
	}
	if g.State.Expedition == nil || g.State.Expedition.Fight == nil {
		t.Fatal("the test never reached a fight")
	}
	data, err := EncodeSave(g)
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodeSave(data)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := EncodeSave(back)
	if string(again) != string(data) {
		t.Fatal("a save changed on the way through load and save")
	}
	// The loaded game plays out the same future, command for command.
	for step := 0; step < 400 && g.State.Phase != PhaseOver; step++ {
		var c Command
		switch {
		case g.State.Phase == PhaseDay && g.State.Card != nil && !g.State.Card.Resolved:
			c = ResolveEvent(1)
			if _, _, err := Apply(g.State, c); err != nil {
				c = ResolveEvent(0)
			}
		case g.State.Phase == PhaseDay:
			c = EndDay()
		case g.State.Phase == PhaseDusk:
			c = StayHome()
		case g.State.Expedition.Fight != nil:
			c = Act(ActDefend, -1)
		default:
			c = GoHome()
		}
		e1, err1 := g.Do(c)
		e2, err2 := back.Do(c)
		if err1 != nil || err2 != nil || encode(t, e1) != encode(t, e2) {
			t.Fatalf("step %d (%s): the loaded game diverged: %v %v", step, c.Kind, err1, err2)
		}
	}
	a, _ := EncodeSave(g)
	b, _ := EncodeSave(back)
	if string(a) != string(b) || g.State.Phase != PhaseOver {
		t.Errorf("after playing on, the two games differ, or the tenure never ended (phase %q)", g.State.Phase)
	}

	t.Run("bad saves are refused", func(t *testing.T) {
		if _, err := DecodeSave([]byte("not json")); err == nil {
			t.Error("garbage loaded")
		}
		if _, err := DecodeSave([]byte(`{"version":999,"state":{},"log":[]}`)); !errors.Is(err, ErrSaveVersion) {
			t.Errorf("a save from another version: %v", err)
		}
	})
	t.Run("the ledger round-trips", func(t *testing.T) {
		l := Ledger{}.Record(g.Log)
		data, err := EncodeLedger(&l)
		if err != nil {
			t.Fatal(err)
		}
		back, err := DecodeLedger(data)
		if err != nil || encode(t, back) != encode(t, l) {
			t.Errorf("ledger changed in the round trip: %v", err)
		}
		if l.Tenures != 1 {
			t.Errorf("the ledger counts %d finished tenures", l.Tenures)
		}
		if _, err := DecodeLedger([]byte(`{"version":0}`)); !errors.Is(err, ErrSaveVersion) {
			t.Errorf("an old ledger: %v", err)
		}
	})
	t.Run("a new tenure starts a new log", func(t *testing.T) {
		if _, err := g.Do(NewTenure(32)); err != nil {
			t.Fatal(err)
		}
		if len(g.Log) == 0 || g.Log[0].Kind != EvTenureStarted || count(g.Log, EvTenureEnded) != 0 {
			t.Errorf("log after a new tenure: %d events", len(g.Log))
		}
	})
}
