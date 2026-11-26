package sim

import (
	"fmt"
	"testing"

	"hollowward/internal/content"
)

// foe makes an enemy with exact numbers.
func foe(kind content.EnemyID, hp, attack, guard, speed int) Enemy {
	return Enemy{Kind: kind, Name: content.Enemies[kind].Name, HP: hp, MaxHP: hp, Attack: attack, Guard: guard, Speed: speed,
		Cause: Cause{FromNeed, "docks/supply", 10}}
}

// arena puts citizens with the given jobs into a fight with the given
// enemies. The citizens have no traits that touch combat, so their stats are
// their job's.
func arena(t testing.TB, jobs []content.JobID, enemies ...Enemy) State {
	t.Helper()
	s := start(t, 1)
	s.Citizens = nil
	var ids []int
	for i, j := range jobs {
		c := Citizen{ID: i, Name: fmt.Sprintf("C%d", i), Job: j, Home: content.Docks, Alive: true,
			Traits: [2]content.TraitID{content.Tireless, content.Lucky}}
		c.HP = c.Stats().HP
		s.Citizens = append(s.Citizens, c)
		ids = append(ids, i)
	}
	bossFoe := Enemy{Kind: -1, Boss: true, Name: "The Starved Harbourmaster", HP: 30, MaxHP: 30, Attack: 5, Guard: 1, Speed: 4,
		Cause: Cause{FromFigure, "docks/figure", 10}}
	h := Hollow{District: content.Docks, Night: 1, Grievance: content.Supply, Entry: 0, Heart: 1, Rooms: []Room{
		{ID: 0, Kind: RoomEntry, Name: "Arena", Exits: []int{1}, Building: -1, Enemies: enemies, Cause: Cause{FromDistrict, "docks", int(content.Docks)}},
		{ID: 1, Kind: RoomHeart, Name: "Heart", Depth: 1, Building: -1, Enemies: []Enemy{bossFoe}, Cause: Cause{FromDistrict, "docks", int(content.Docks)}},
	}}
	s.Districts[content.Docks].Turned = true
	s.Hollows[content.Docks] = &h
	s.Phase, s.Dusk = PhaseDusk, &DuskReport{Turned: content.NoDistrict}
	s, _ = do(t, s, SendParty(content.Docks, ids...))
	return s
}

func speeds(s *State) (out []int, enemy []bool) {
	r := s.room()
	for _, a := range s.Expedition.Fight.Order {
		if a.Enemy {
			out = append(out, r.Enemies[a.Index].Speed)
		} else {
			out = append(out, s.Citizens[s.Expedition.Party[a.Index].Citizen].Stats().Speed)
		}
		enemy = append(enemy, a.Enemy)
	}
	return out, enemy
}

func TestInitiativeFollowsSpeed(t *testing.T) {
	// Priest 5, dockhand 5, watchman 4, miller 3; enemies at 9, 4 and 1.
	s := arena(t, []content.JobID{content.Miller, content.Watchman, content.Priest, content.Dockhand},
		foe(content.Gaunt, 50, 1, 0, 9), foe(content.Gnawer, 50, 1, 0, 4), foe(content.Sump, 50, 1, 0, 1))
	sp, enemy := speeds(&s)
	want := []int{9, 5, 5, 4, 4, 3, 1}
	if fmt.Sprint(sp) != fmt.Sprint(want) {
		t.Fatalf("order by speed %v, want %v", sp, want)
	}
	// At speed 4 the watchman (a citizen) goes before the Gnawer.
	if enemy[3] || !enemy[4] {
		t.Errorf("a tie in Speed must go to the citizen: %v", enemy)
	}
	// The fastest enemy has acted; play rests on the first citizen, the priest (slot 2).
	if s.Acting() != 2 {
		t.Errorf("acting slot %d, want the priest in slot 2", s.Acting())
	}
	// Every round keeps the rule.
	for range 3 {
		for s.Expedition != nil && s.Expedition.Fight != nil && s.Expedition.Fight.Round < 4 {
			round := s.Expedition.Fight.Round
			s, _ = do(t, s, Act(ActDefend, -1))
			if s.Expedition.Fight.Round != round {
				sp, _ := speeds(&s)
				for i := 1; i < len(sp); i++ {
					if sp[i] > sp[i-1] {
						t.Fatalf("round %d order %v is not fastest first", s.Expedition.Fight.Round, sp)
					}
				}
				break
			}
		}
	}
}

func TestHookAndShoveMoveTurnOrder(t *testing.T) {
	// A dockhand hooks the fast enemy: next round it acts last.
	s := arena(t, []content.JobID{content.Dockhand}, foe(content.Gaunt, 50, 1, 0, 9), foe(content.Gnawer, 50, 1, 0, 2))
	s, _ = do(t, s, Act(ActAbility, 0))
	f := s.Expedition.Fight
	if f.Round != 2 || !f.Order[len(f.Order)-1].Enemy || f.Order[len(f.Order)-1].Index != 0 {
		t.Errorf("round %d order %+v: the hooked Gaunt should act last", f.Round, f.Order)
	}
	if s.Acting() != 0 || f.Turn != 0 {
		t.Errorf("the dockhand should now act first: turn %d", f.Turn)
	}
	// An Ash-Mask's shove puts a citizen at the back of the next round.
	s = arena(t, []content.JobID{content.Merchant, content.Miller}, foe(content.AshMask, 50, 3, 0, 1))
	s.room().Enemies[0].Intent = Intent{content.MoveShove, 0} // aim it at the fast merchant
	s, _ = do(t, s, Act(ActDefend, -1))
	s, _ = do(t, s, Act(ActDefend, -1))
	f = s.Expedition.Fight
	if f.Round != 2 || f.Order[0].Enemy || f.Order[0].Index != 1 {
		t.Fatalf("round %d order %+v: the miller should lead now that the merchant is shoved back", f.Round, f.Order)
	}
	last := f.Order[len(f.Order)-1]
	if last.Enemy || last.Index != 0 {
		t.Errorf("order %+v: the shoved merchant should act last", f.Order)
	}
}

func TestAbilities(t *testing.T) {
	ab, tn := content.T.Ability, content.T
	t.Run("hammer wounds and stuns", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Smith}, foe(content.Gnawer, 50, 4, 1, 1))
		hp := s.Citizens[0].HP
		s, ev := do(t, s, Act(ActAbility, 0))
		e := s.room().Enemies[0]
		if want := 50 - (tn.JobStats[content.Smith].Attack - 1); e.HP != want {
			t.Errorf("enemy HP %d, want %d", e.HP, want)
		}
		acted := false
		for _, x := range ev {
			if x.Kind == EvEnemyActed {
				acted = true
				if x.ID != "stunned" {
					t.Errorf("the stunned enemy still acted: %+v", x)
				}
			}
		}
		if !acted || s.Citizens[0].HP != hp || e.Stunned {
			t.Errorf("stun should cost exactly one action: acted %v, citizen HP %d of %d, still stunned %v", acted, s.Citizens[0].HP, hp, e.Stunned)
		}
	})
	t.Run("a boss cannot be stunned", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Smith})
		s, _ = do(t, s, Advance(1))
		s, ev := do(t, s, Act(ActAbility, 0))
		for _, x := range ev {
			if x.Kind == EvEnemyActed && x.ID == "stunned" {
				t.Error("the boss was stunned")
			}
		}
		if s.Expedition != nil && s.room().Enemies[0].Stunned {
			t.Error("the boss carries a stun")
		}
	})
	t.Run("mend heals an ally, never past full", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Priest, content.Smith}, foe(content.Gnawer, 50, 1, 0, 1))
		s.Citizens[1].HP = 3
		s, _ = do(t, s, Act(ActAbility, 1))
		if s.Citizens[1].HP != 3+ab.MendHeal {
			t.Errorf("mended to %d, want %d", s.Citizens[1].HP, 3+ab.MendHeal)
		}
		s = arena(t, []content.JobID{content.Priest, content.Smith}, foe(content.Gnawer, 50, 1, 0, 1))
		top := s.Citizens[1].Stats().HP
		s.Citizens[1].HP = top - 2
		s, _ = do(t, s, Act(ActAbility, 1))
		if s.Citizens[1].HP != top {
			t.Errorf("mended to %d, want full at %d", s.Citizens[1].HP, top)
		}
		mustFail(t, arena(t, []content.JobID{content.Priest}, foe(content.Gnawer, 50, 1, 0, 1)), Act(ActAbility, 7), ErrBadArgument)
	})
	t.Run("grind ignores guard", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Miller}, foe(content.Sump, 50, 1, 9, 1))
		s, _ = do(t, s, Act(ActAbility, 0))
		if want := 50 - (tn.JobStats[content.Miller].Attack + ab.GrindBonus); s.room().Enemies[0].HP != want {
			t.Errorf("enemy HP %d, want %d", s.room().Enemies[0].HP, want)
		}
		s = arena(t, []content.JobID{content.Miller}, foe(content.Sump, 50, 1, 9, 1))
		s, _ = do(t, s, Act(ActAttack, 0))
		if s.room().Enemies[0].HP != 49 {
			t.Errorf("a plain blow against heavy guard should do 1: HP %d", s.room().Enemies[0].HP)
		}
	})
	t.Run("haggle lowers an enemy's attack, to a floor", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Merchant, content.Merchant, content.Merchant}, foe(content.Sump, 50, 9, 0, 1))
		for range 3 {
			s, _ = do(t, s, Act(ActAbility, 0))
		}
		if got := s.room().Enemies[0].AttackMod; got != ab.HaggleFloor {
			t.Errorf("attack modifier %d after three haggles, want the floor %d", got, ab.HaggleFloor)
		}
	})
	t.Run("heave strikes every enemy", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Labourer}, foe(content.Gaunt, 50, 1, 0, 1), foe(content.Gnawer, 50, 1, 1, 1), foe(content.Sump, 50, 1, 5, 1))
		s, _ = do(t, s, Act(ActAbility, -1))
		base := tn.JobStats[content.Labourer].Attack - ab.HeavePenalty
		for i, guard := range []int{0, 1, 5} {
			if want := 50 - max(base-guard, 1); s.room().Enemies[i].HP != want {
				t.Errorf("enemy %d HP %d, want %d", i, s.room().Enemies[i].HP, want)
			}
		}
	})
	t.Run("hold takes the blow meant for an ally", func(t *testing.T) {
		// Watchman (speed 4) acts before the enemy (speed 1), which aims at the priest.
		s := arena(t, []content.JobID{content.Watchman, content.Priest}, foe(content.Gnawer, 50, 6, 0, 1))
		s.room().Enemies[0].Intent = Intent{content.MoveStrike, 1}
		priest, watch := s.Citizens[1].HP, s.Citizens[0].HP
		s, _ = do(t, s, Act(ActDefend, -1)) // the priest, speed 5
		s, _ = do(t, s, Act(ActAbility, 1)) // the watchman holds the priest
		if s.Citizens[1].HP != priest {
			t.Errorf("the held priest was hit: HP %d of %d", s.Citizens[1].HP, priest)
		}
		if want := watch - max(6-tn.JobStats[content.Watchman].Guard-ab.HoldGuard, 1); s.Citizens[0].HP != want {
			t.Errorf("watchman HP %d, want %d", s.Citizens[0].HP, want)
		}
		mustFail(t, arena(t, []content.JobID{content.Watchman}, foe(content.Gnawer, 50, 1, 0, 1)), Act(ActAbility, 0), ErrBadArgument)
	})
	t.Run("defending takes less", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Merchant}, foe(content.Gnawer, 50, 6, 0, 1))
		s.room().Enemies[0].Intent = Intent{content.MoveStrike, 0}
		hp := s.Citizens[0].HP
		s, _ = do(t, s, Act(ActDefend, -1))
		if want := hp - (6 - tn.DefendGuard); s.Citizens[0].HP != want {
			t.Errorf("HP %d, want %d", s.Citizens[0].HP, want)
		}
	})
	t.Run("an ability needs a rest between uses", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Miller}, foe(content.Sump, 500, 1, 0, 1))
		s, _ = do(t, s, Act(ActAbility, 0))
		mustFail(t, s, Act(ActAbility, 0), ErrNotReady)
		s, _ = do(t, s, Act(ActAttack, 0))
		do(t, s, Act(ActAbility, 0))
	})
	t.Run("bad targets are refused", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Smith}, foe(content.Gaunt, 50, 1, 0, 1))
		mustFail(t, s, Act(ActAttack, 5), ErrBadArgument)
		mustFail(t, s, Act(ActAttack, -1), ErrBadArgument)
		mustFail(t, s, Act("dance", 0), ErrBadArgument)
		mustFail(t, s, GoHome(), ErrInFight)
		mustFail(t, s, Advance(1), ErrInFight)
	})
}

// oneMove sets an enemy's move for round 1 and lets a single slow enemy act
// against citizens who all brace... or not, as given.
func oneMove(t testing.TB, move content.MoveID, target int, jobs []content.JobID, attack int, prep func(s *State)) (State, []Event) {
	t.Helper()
	s := arena(t, jobs, foe(content.Sump, 500, attack, 0, 1), foe(content.Gaunt, 500, 1, 0, 1))
	s.room().Enemies[0].Intent = Intent{move, target}
	s.room().Enemies[1].Intent = Intent{content.MoveGather, -1}
	if prep != nil {
		prep(&s)
	}
	var all []Event
	for range jobs {
		var ev []Event
		s, ev = do(t, s, Act(ActAttack, 1))
		all = append(all, ev...)
	}
	return s, all
}

func TestEnemyMoves(t *testing.T) {
	mv := content.T.Move
	two := []content.JobID{content.Merchant, content.Priest} // guard 0 both; HP 10 and 11
	full := func(s *State, id int) int { return s.Citizens[id].Stats().HP }
	t.Run("strike", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveStrike, 1, two, 4, nil)
		if s.Citizens[1].HP != full(&s, 1)-4 || s.Citizens[0].HP != full(&s, 0) {
			t.Errorf("HP %d, %d", s.Citizens[0].HP, s.Citizens[1].HP)
		}
	})
	t.Run("gorge, maul and engulf hit harder", func(t *testing.T) {
		for move, bonus := range map[content.MoveID]int{content.MoveGorge: mv.GorgeBonus, content.MoveMaul: mv.MaulBonus, content.MoveEngulf: mv.EngulfBonus} {
			s, _ := oneMove(t, move, 0, two, 3, nil)
			if want := full(&s, 0) - 3 - bonus; s.Citizens[0].HP != want {
				t.Errorf("%s: HP %d, want %d", content.Moves[move].Name, s.Citizens[0].HP, want)
			}
		}
	})
	t.Run("gnaw takes guard for the rest of the fight", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveGnaw, 0, two, 4, nil)
		if s.Citizens[0].HP != full(&s, 0)-(4-mv.GnawPenalty) || s.Expedition.Party[0].GuardMod != -mv.GnawGuard {
			t.Errorf("HP %d, guard modifier %d", s.Citizens[0].HP, s.Expedition.Party[0].GuardMod)
		}
	})
	t.Run("brand burns at the start of the next turn", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveBrand, 0, two, 3, nil)
		// Round 2 has begun and the merchant, fastest, has already burned.
		if want := full(&s, 0) - 3 - mv.BurnDamage; s.Citizens[0].HP != want || s.Expedition.Party[0].Burning != 0 {
			t.Errorf("HP %d, want %d after the burn", s.Citizens[0].HP, want)
		}
	})
	t.Run("surge and break ranks hit everyone", func(t *testing.T) {
		for move, penalty := range map[content.MoveID]int{content.MoveSurge: mv.SurgePenalty, content.MoveBreakRanks: mv.BreakRanksPenalty} {
			s, _ := oneMove(t, move, -1, two, 5, nil)
			for id := range two {
				if want := full(&s, id) - (5 - penalty); s.Citizens[id].HP != want {
					t.Errorf("%s: citizen %d HP %d, want %d", content.Moves[move].Name, id, s.Citizens[id].HP, want)
				}
			}
		}
	})
	t.Run("collapse ignores guard", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveCollapse, -1, []content.JobID{content.Watchman, content.Watchman}, 1, nil)
		for id := range 2 {
			if want := full(&s, id) - mv.CollapseDamage; s.Citizens[id].HP != want {
				t.Errorf("citizen %d HP %d, want %d", id, s.Citizens[id].HP, want)
			}
		}
	})
	t.Run("spore and wither sap attack; spores sap guard", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveSpore, -1, two, 5, nil)
		s2, _ := oneMove(t, content.MoveWither, -1, two, 5, nil)
		s3, _ := oneMove(t, content.MoveSpores, -1, two, 5, nil)
		for slot := range two {
			if s.Expedition.Party[slot].AttackMod != -mv.SporeAttack || s2.Expedition.Party[slot].AttackMod != -mv.WitherAttack || s3.Expedition.Party[slot].GuardMod != -mv.SporesGuard {
				t.Errorf("slot %d modifiers: %+v %+v %+v", slot, s.Expedition.Party[slot], s2.Expedition.Party[slot], s3.Expedition.Party[slot])
			}
			if s.Citizens[slot].HP != full(&s, slot) {
				t.Error("spore should not wound")
			}
		}
	})
	t.Run("devour heals the boss", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveDevour, 0, two, 3, func(s *State) { s.room().Enemies[0].HP = 100 })
		if got := s.room().Enemies[0].HP; got != 100+mv.DevourHeal {
			t.Errorf("devourer HP %d, want %d", got, 100+mv.DevourHeal)
		}
		if want := full(&s, 0) - 3 - mv.DevourBonus; s.Citizens[0].HP != want {
			t.Errorf("victim HP %d, want %d", s.Citizens[0].HP, want)
		}
	})
	t.Run("rally makes every enemy hit harder", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveRally, -1, two, 3, nil)
		for i := range s.room().Enemies {
			if s.room().Enemies[i].AttackMod != mv.RallyAttack {
				t.Errorf("enemy %d attack modifier %d", i, s.room().Enemies[i].AttackMod)
			}
		}
	})
	t.Run("lunge seeks the weakest and maul the strongest", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Watchman, content.Merchant, content.Smith})
		e := Enemy{Kind: content.Gaunt}
		if got := s.intent(&e); got.Move != content.MoveLunge || got.Target != 1 {
			t.Errorf("a Gaunt opens with %+v, want a lunge at the merchant", got)
		}
		e = Enemy{Kind: content.MouldHound}
		if got := s.intent(&e); got.Move != content.MoveMaul || got.Target != 0 {
			t.Errorf("a Mould-Hound opens with %+v, want a maul at the watchman", got)
		}
	})
	t.Run("a blow meant for the fallen turns on the next citizen", func(t *testing.T) {
		s, _ := oneMove(t, content.MoveStrike, 0, two, 4, func(s *State) {
			s.Citizens[0].HP = 1                                       // the first enemy fells the merchant...
			s.room().Enemies[1].Intent = Intent{content.MoveStrike, 0} // ...whom the second also meant to hit
		})
		if !s.Citizens[0].Downed || s.Citizens[1].HP != full(&s, 1)-1 {
			t.Errorf("merchant downed %v; priest HP %d, want %d", s.Citizens[0].Downed, s.Citizens[1].HP, full(&s, 1)-1)
		}
	})
}

func TestEveryFightEnds(t *testing.T) {
	tn := content.T
	t.Run("the round cap drives the party out", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Watchman}, foe(content.Sump, 100000, 1, 0, 1))
		var ev []Event
		acts := 0
		for s.Phase == PhaseNight {
			var more []Event
			s, more = do(t, s, Act(ActAttack, 0))
			ev = append(ev, more...)
			if acts++; acts > tn.RoundCap {
				t.Fatalf("still fighting after %d actions", acts)
			}
		}
		if acts != tn.RoundCap || count(ev, EvDrivenOut) != 1 || s.Phase != PhaseDay {
			t.Errorf("%d actions, %d driven-out events, phase %q", acts, count(ev, EvDrivenOut), s.Phase)
		}
		e := s.Hollows[content.Docks].Rooms[0].Enemies[0]
		if e.HP != 100000-tn.RoundCap*tn.JobStats[content.Watchman].Attack {
			t.Errorf("the enemy's wounds should stay: HP %d", e.HP)
		}
		if e.NextMove != 0 || e.Acted || e.AttackMod != 0 || !s.Citizens[0].Exhausted {
			t.Errorf("the broken-off fight left its moods behind: %+v", e)
		}
	})
	t.Run("a party that falls is carried out", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Merchant, content.Priest}, foe(content.HollowBelly, 500, 40, 0, 9))
		var ev []Event
		for s.Phase == PhaseNight {
			var more []Event
			s, more = do(t, s, Act(ActAttack, 0))
			ev = append(ev, more...)
		}
		// The enemy is faster, so the first citizen fell before the first command.
		if count(ev, EvWiped) != 1 || count(ev, EvDowned) != 1 || count(ev, EvDied)+count(ev, EvScarred) != 2 {
			t.Errorf("wiped %d, downed %d, died %d, scarred %d", count(ev, EvWiped), count(ev, EvDowned), count(ev, EvDied), count(ev, EvScarred))
		}
		for _, c := range s.Citizens {
			if c.Downed || (c.Alive && c.HP != c.Stats().HP) {
				t.Errorf("%s after the night: %+v", c.Name, c)
			}
		}
		if !s.Districts[content.Docks].Turned {
			t.Error("a wipe must leave the district turned")
		}
	})
	t.Run("HP stops at zero on both sides", func(t *testing.T) {
		s := arena(t, []content.JobID{content.Miller}, foe(content.Gaunt, 1, 1, 0, 1), foe(content.Gnawer, 3, 1, 0, 1))
		s, ev := do(t, s, Act(ActAbility, 0)) // a grind far larger than 1 HP
		if s.room().Enemies[0].HP != 0 || count(ev, EvEnemyFell) != 1 || count(ev, EvLoot) != 1 {
			t.Errorf("enemy HP %d", s.room().Enemies[0].HP)
		}
		mustFail(t, s, Act(ActAttack, 0), ErrBadArgument) // the dead are no target
	})
}

func TestKillsDropTheFamilysMaterial(t *testing.T) {
	s := arena(t, []content.JobID{content.Miller, content.Miller}, foe(content.Gaunt, 1, 1, 0, 1), foe(content.AshMask, 1, 1, 0, 1), foe(content.Blackcap, 1, 1, 0, 1))
	before := s.Materials
	for target := range 3 {
		s, _ = do(t, s, Act(ActAttack, target))
	}
	for f := content.Family(0); f < content.NumFamilies; f++ {
		if s.Materials[f] != before[f]+1 || s.Expedition.Loot[f] != 1 {
			t.Errorf("%s: %d, want one more than %d", f.Material(), s.Materials[f], before[f])
		}
	}
	if s.Expedition.Fight != nil || !s.room().Cleared {
		t.Error("the room should be cleared with every enemy dead")
	}
}
