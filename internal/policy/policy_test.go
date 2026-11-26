package policy

import (
	"encoding/json"
	"testing"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

func encode(t testing.TB, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// invariants holds a state to the rules that must hold after every command.
func invariants(t testing.TB, s *sim.State) {
	t.Helper()
	for i := range s.Citizens {
		c := &s.Citizens[i]
		if c.HP < 0 || c.HP > c.Stats().HP {
			t.Fatalf("day %d: %s has %d HP of %d", s.Day, c.Name, c.HP, c.Stats().HP)
		}
		if c.Downed && (c.HP != 0 || s.Expedition == nil) {
			t.Fatalf("day %d: %s is Downed with %d HP, expedition %v", s.Day, c.Name, c.HP, s.Expedition != nil)
		}
	}
	for d, h := range s.Hollows {
		dist := &s.Districts[d]
		if (h != nil) != dist.Turned {
			t.Fatalf("day %d: %s turned=%v but hollow present=%v", s.Day, dist.ID, dist.Turned, h != nil)
		}
		if h == nil {
			continue
		}
		hearts := 0
		for i := range h.Rooms {
			r := &h.Rooms[i]
			if r.Kind == sim.RoomHeart {
				hearts++
			}
			// A turned district is frozen, so every cause still points at real state.
			if v, ok := r.Cause.Resolve(dist); !ok || v != r.Cause.ValueAtGeneration {
				t.Fatalf("day %d: room %q cause %+v no longer matches the district", s.Day, r.Name, r.Cause)
			}
			for j := range r.Enemies {
				e := &r.Enemies[j]
				if e.HP < 0 || e.HP > e.MaxHP {
					t.Fatalf("day %d: %s has %d HP of %d", s.Day, e.Name, e.HP, e.MaxHP)
				}
				if v, ok := e.Cause.Resolve(dist); !ok || v != e.Cause.ValueAtGeneration {
					t.Fatalf("day %d: %s cause %+v no longer matches the district", s.Day, e.Name, e.Cause)
				}
			}
		}
		if hearts != 1 {
			t.Fatalf("day %d: hollow of %s has %d hearts", s.Day, dist.ID, hearts)
		}
	}
	if exp := s.Expedition; exp != nil {
		if s.Phase != sim.PhaseNight {
			t.Fatalf("an expedition outside the night: phase %q", s.Phase)
		}
		if exp.Fight != nil {
			slot := s.Acting()
			if slot < 0 {
				t.Fatalf("day %d: a fight is waiting, but not on a citizen", s.Day)
			}
			if c := &s.Citizens[exp.Party[slot].Citizen]; c.Downed || !c.Alive {
				t.Fatalf("day %d: it is %s's turn, who cannot act", s.Day, c.Name)
			}
			if exp.Fight.Round < 1 || exp.Fight.Round > content.T.RoundCap {
				t.Fatalf("day %d: round %d", s.Day, exp.Fight.Round)
			}
		}
	} else if s.Phase == sim.PhaseNight {
		t.Fatal("night without an expedition")
	}
	for _, v := range s.Resources {
		if v < 0 {
			t.Fatalf("day %d: resources %v", s.Day, s.Resources)
		}
	}
}

// shown records what every living enemy has announced: the moves it may make
// before the player next chooses.
func shown(s *sim.State) map[int][]content.MoveID {
	out := map[int][]content.MoveID{}
	if s.Expedition == nil || s.Expedition.Fight == nil {
		return nil
	}
	room := &s.Hollows[s.Expedition.District].Rooms[s.Expedition.Room]
	for i := range room.Enemies {
		e := &room.Enemies[i]
		if e.HP <= 0 {
			continue
		}
		if e.Acted {
			out[i] = []content.MoveID{e.Next.Move}
		} else {
			out[i] = []content.MoveID{e.Intent.Move, e.Next.Move}
		}
	}
	return out
}

// Every policy plays whole tenures; after every command the invariants hold,
// every enemy action was announced before the choice that preceded it, and
// every fight ends.
func TestPoliciesPlayByTheRules(t *testing.T) {
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			fights, enemyActions, turned, cleared := 0, 0, 0, 0
			for seed := uint64(1); seed <= 12; seed++ {
				p, err := New(name, seed)
				if err != nil {
					t.Fatal(err)
				}
				s, _, err := sim.Apply(sim.State{}, sim.NewTenure(seed))
				if err != nil {
					t.Fatal(err)
				}
				actsThisFight := 0
				for steps := 0; s.Phase != sim.PhaseOver && s.Day <= content.T.Nights; steps++ {
					if steps > 100000 {
						t.Fatalf("seed %d: no end in sight", seed)
					}
					announced := shown(&s)
					inFight := s.Expedition != nil && s.Expedition.Fight != nil
					cmd := p.Next(&s)
					before := encode(t, s)
					next, events, err := sim.Apply(s, cmd)
					if err != nil {
						t.Fatalf("seed %d day %d: %s sent %s: %v", seed, s.Day, name, cmd.Kind, err)
					}
					if encode(t, s) != before {
						t.Fatalf("seed %d: Apply changed its input on %s", seed, cmd.Kind)
					}
					s = next
					invariants(t, &s)

					if inFight {
						if actsThisFight++; actsThisFight > content.T.RoundCap*content.T.MaxParty {
							t.Fatalf("seed %d: a fight has taken %d actions", seed, actsThisFight)
						}
					} else {
						actsThisFight = 0
					}
					for _, e := range events {
						switch e.Kind {
						case sim.EvFightStarted:
							fights++
							announced = nil // nothing was on show before the door opened
						case sim.EvTurned:
							turned++
						case sim.EvHeartCleared:
							cleared++
						case sim.EvEnemyActed:
							enemyActions++
							if announced == nil || e.ID == "stunned" {
								continue
							}
							moves := announced[e.N]
							if len(moves) == 0 || content.Moves[moves[0]].Name != e.ID {
								t.Fatalf("seed %d day %d: enemy %d did %q, but had announced %v", seed, s.Day, e.N, e.ID, moves)
							}
							announced[e.N] = moves[1:]
						}
					}
				}
			}
			t.Logf("%s: %d districts turned, %d hearts cleared, %d fights, %d enemy actions checked", name, turned, cleared, fights, enemyActions)
			if fights == 0 || enemyActions == 0 || turned == 0 {
				t.Errorf("%s never reached the night: %d turned, %d fights", name, turned, fights)
			}
		})
	}
}

func TestPlayIsDeterministic(t *testing.T) {
	for _, name := range Names() {
		run := func() (string, string) {
			p, _ := New(name, 5)
			s, _, _ := sim.Apply(sim.State{}, sim.NewTenure(5))
			s, ev, err := Play(s, p, func(s *sim.State) bool { return s.Day > 8 })
			if err != nil {
				t.Fatal(err)
			}
			return encode(t, s), encode(t, ev)
		}
		s1, e1 := run()
		s2, e2 := run()
		if s1 != s2 || e1 != e2 {
			t.Errorf("%s: two runs from one seed differ", name)
		}
	}
}

func TestNew(t *testing.T) {
	for _, name := range Names() {
		if p, err := New(name, 1); err != nil || p.Name() != name {
			t.Errorf("New(%q) = %v, %v", name, p, err)
		}
	}
	if _, err := New("pacifist", 1); err == nil {
		t.Error("an unknown policy must be an error")
	}
}

// The farm policy never staffs its field, and repair-worst always staffs the
// three most neglected districts.
func TestWardensStaffAsPromised(t *testing.T) {
	s, _, _ := sim.Apply(sim.State{}, sim.NewTenure(3))
	farm := Farm{Field: content.Market}
	for s.Phase == sim.PhaseDay {
		cmd := farm.Next(&s)
		if cmd.Kind == sim.CmdAssignSteward && cmd.District == content.Market {
			t.Fatal("the farm policy staffed its own field")
		}
		if cmd.Kind == sim.CmdIssueDecree && content.Decrees[cmd.Decree].Targeted && cmd.District == content.Market {
			t.Fatal("the farm policy sent relief to its own field")
		}
		s, _, _ = sim.Apply(s, cmd)
	}

	s, _, _ = sim.Apply(sim.State{}, sim.NewTenure(3))
	want := byNeglect(&s)[:content.NumStewards]
	rw := RepairWorst{}
	for range content.NumStewards {
		s, _, _ = sim.Apply(s, rw.Next(&s))
	}
	for i, post := range s.Stewards {
		if post.District != want[i] {
			t.Errorf("steward %d at %s, want %s", i, post.District, want[i])
		}
	}
}
