package sim

import (
	"fmt"
	"sort"

	"hollowward/internal/content"
)

// Fight is a fight in the party's current room. Between commands it always
// rests on a citizen's turn: enemies act on their own.
type Fight struct {
	Round int     `json:"round"`
	Order []Actor `json:"order"` // this round's initiative, fastest first
	Turn  int     `json:"turn"`  // index into Order of who acts now
}

// Actor is one place in the initiative order.
type Actor struct {
	Enemy bool `json:"enemy"`
	Index int  `json:"index"` // a party slot, or an enemy's index in the room
}

// Acting is the party slot whose turn it is, or -1 when there is no fight.
func (s *State) Acting() int {
	if s.Expedition == nil || s.Expedition.Fight == nil {
		return -1
	}
	f := s.Expedition.Fight
	if f.Turn >= len(f.Order) || f.Order[f.Turn].Enemy {
		return -1
	}
	return f.Order[f.Turn].Index
}

func (s *State) startFight(l *log) {
	exp := s.Expedition
	exp.Fight = &Fight{}
	for i := range exp.Party {
		m := &exp.Party[i]
		m.AttackMod, m.GuardMod, m.Delayed, m.Burning, m.Defending, m.Holding, m.Cooldown = 0, 0, false, 0, false, -1, 0
		if c := &s.Citizens[m.Citizen]; c.Alive && !c.Downed {
			m.Fought = true
		}
	}
	s.resetRoom(s.room())
	l.add(EvFightStarted).in(exp.District).n(exp.Room)
	s.beginRound(l)
}

// beginRound opens a round: enemies show their intents, initiative is set by
// Speed, and play runs on until a citizen must choose.
func (s *State) beginRound(l *log) {
	t := content.T
	exp, f, r := s.Expedition, s.Expedition.Fight, s.room()
	f.Round++
	if f.Round > t.RoundCap {
		// Every fight ends: at the cap the hollow drives the party out.
		l.add(EvDrivenOut).in(exp.District).n(exp.Room)
		s.endNight(l)
		return
	}
	l.add(EvRound).in(exp.District).n(f.Round)

	// Enemies show their hand: this round's action was announced last round,
	// and next round's is announced now.
	for i := range r.Enemies {
		if e := &r.Enemies[i]; e.HP > 0 {
			if f.Round == 1 {
				e.Next = s.intent(e)
			}
			e.Intent, e.Next, e.Acted = e.Next, s.intent(e), false
		}
	}

	// Initiative: Speed, fastest first. The delayed and the hooked go last.
	// Ties go to citizens, then to the lower slot or index, so the order
	// needs no dice.
	type entry struct {
		Actor
		speed int
		late  bool
	}
	var order []entry
	for _, slot := range s.standing() {
		m := &exp.Party[slot]
		order = append(order, entry{Actor{false, slot}, s.Citizens[m.Citizen].Stats().Speed, m.Delayed})
		m.Delayed = false
	}
	for i := range r.Enemies {
		if e := &r.Enemies[i]; e.HP > 0 {
			order = append(order, entry{Actor{true, i}, e.Speed, e.Hooked})
			e.Hooked = false
		}
	}
	sort.SliceStable(order, func(a, b int) bool {
		x, y := order[a], order[b]
		if x.late != y.late {
			return !x.late
		}
		if x.speed != y.speed {
			return x.speed > y.speed
		}
		if x.Enemy != y.Enemy {
			return !x.Enemy
		}
		return x.Index < y.Index
	})
	f.Order = f.Order[:0]
	for _, e := range order {
		f.Order = append(f.Order, e.Actor)
	}
	f.Turn = 0
	s.runFight(l)
}

// runFight plays enemy turns until a citizen has to act or the fight ends.
func (s *State) runFight(l *log) {
	for {
		exp := s.Expedition
		if exp == nil || exp.Fight == nil {
			return // the night ended under us
		}
		f, r := exp.Fight, s.room()
		if r.Alive() == 0 {
			s.winFight(l)
			return
		}
		if len(s.standing()) == 0 {
			l.add(EvWiped).in(exp.District)
			s.endNight(l)
			return
		}
		if f.Turn >= len(f.Order) {
			s.beginRound(l)
			return
		}
		a := f.Order[f.Turn]
		if a.Enemy {
			if r.Enemies[a.Index].HP > 0 {
				s.enemyActs(a.Index, l)
			}
			f.Turn++
			continue
		}
		m := &exp.Party[a.Index]
		c := &s.Citizens[m.Citizen]
		if c.Downed {
			f.Turn++
			continue
		}
		// The start of a citizen's turn.
		m.Defending, m.Holding = false, -1
		if m.Cooldown > 0 {
			m.Cooldown--
		}
		if m.Burning > 0 {
			burn := m.Burning
			m.Burning = 0
			s.hurt(a.Index, burn, l)
			if c.Downed {
				f.Turn++
				continue
			}
		}
		return // wait for the player
	}
}

func (s *State) winFight(l *log) {
	exp, h, r := s.Expedition, s.hollow(), s.room()
	exp.Fight = nil
	r.Cleared = true
	l.add(EvRoomCleared).in(h.District).id(r.Name).n(r.ID)
	for _, slot := range s.standing() {
		c := &s.Citizens[exp.Party[slot].Citizen]
		if c.HasTrait(func(t content.TraitDef) bool { return t.Scavenger }) {
			// A scavenger finds one more of whatever fell here.
			for i := range r.Enemies {
				if e := &r.Enemies[i]; !e.Boss {
					f := content.Enemies[e.Kind].Family
					s.Materials[f]++
					exp.Loot[f]++
					l.add(EvLoot).in(h.District).id(f.Material()).n(1).who(c)
					break
				}
			}
		}
	}
	if r.Kind == RoomHeart {
		s.clearHeart(l)
	}
}

// intent picks what an enemy will do this round and at whom. The player sees
// it before anyone acts.
func (s *State) intent(e *Enemy) Intent {
	var moves []content.MoveID
	if e.Boss {
		tw := content.Twists[s.hollow().Grievance]
		moves = []content.MoveID{content.MoveStrike, tw.Moves[0], content.MoveStrike, tw.Moves[1]}
	} else {
		moves = content.EnemyMoves[e.Kind]
	}
	move := moves[e.NextMove%len(moves)]
	e.NextMove++

	exp := s.Expedition
	standing := s.standing()
	pick := func(better func(a, b *Citizen) bool) int {
		best := standing[0]
		for _, slot := range standing[1:] {
			if better(&s.Citizens[exp.Party[slot].Citizen], &s.Citizens[exp.Party[best].Citizen]) {
				best = slot
			}
		}
		return exp.Party[best].Citizen
	}
	switch move {
	case content.MoveSurge, content.MoveSpore, content.MoveGather, content.MoveWither,
		content.MoveRally, content.MoveBreakRanks, content.MoveCollapse, content.MoveSpores:
		return Intent{move, -1}
	case content.MoveLunge:
		return Intent{move, pick(func(a, b *Citizen) bool { return a.HP < b.HP })}
	case content.MoveMaul:
		return Intent{move, pick(func(a, b *Citizen) bool { return a.HP > b.HP })}
	}
	return Intent{move, exp.Party[standing[s.RNG.Combat.IntN(len(standing))]].Citizen}
}

// slotOf finds a citizen's party slot, or -1.
func (s *State) slotOf(citizen int) int {
	for i, m := range s.Expedition.Party {
		if m.Citizen == citizen {
			return i
		}
	}
	return -1
}

// strike lands an enemy's blow on a citizen. A watchman holding the target
// takes it instead. It returns who was hit and for how much.
func (s *State) strike(target, raw int, pierce bool, l *log) (hit, dealt int) {
	t := content.T
	exp := s.Expedition
	slot := s.slotOf(target)
	if slot < 0 || s.Citizens[target].Downed {
		standing := s.standing()
		if len(standing) == 0 {
			return -1, 0
		}
		slot = standing[0] // the one it wanted is down: it turns on the next
	}
	bonus := 0
	for i, m := range exp.Party {
		if m.Holding == exp.Party[slot].Citizen && i != slot && !s.Citizens[m.Citizen].Downed {
			slot, bonus = i, t.Ability.HoldGuard
			break
		}
	}
	m := &exp.Party[slot]
	c := &s.Citizens[m.Citizen]
	dmg := raw
	if !pierce {
		guard := max(c.Stats().Guard+m.GuardMod, 0) + bonus
		if m.Defending {
			guard += t.DefendGuard
		}
		dmg = max(raw-guard, 1)
	}
	return c.ID, s.hurt(slot, dmg, l)
}

// enemyActs carries out the intent the enemy showed.
func (s *State) enemyActs(index int, l *log) {
	mv := content.T.Move
	exp, r := s.Expedition, s.room()
	e := &r.Enemies[index]
	e.Acted = true
	ev := l.add(EvEnemyActed).in(exp.District).n(index)
	ev.Citizen = -1
	if e.Stunned {
		e.Stunned = false
		ev.id("stunned")
		return
	}
	move := e.Intent.Move
	ev.id(content.Moves[move].Name)
	atk := max(e.Attack+e.AttackMod, 1)
	one := func(raw int, pierce bool) int { // hit the shown target
		hit, dealt := s.strike(e.Intent.Target, raw, pierce, l)
		ev = &l.events[indexOf(l, EvEnemyActed, index)]
		if hit >= 0 {
			ev.Citizen, ev.Who, ev.V = hit, s.Citizens[hit].Name, dealt
		}
		return hit
	}
	all := func(raw int, pierce bool) {
		total := 0
		for _, slot := range s.standing() {
			_, dealt := s.strike(exp.Party[slot].Citizen, raw, pierce, l)
			total += dealt
		}
		l.events[indexOf(l, EvEnemyActed, index)].V = total
	}
	each := func(f func(m *Member)) {
		for _, slot := range s.standing() {
			f(&exp.Party[slot])
		}
	}
	switch move {
	case content.MoveStrike, content.MoveLunge:
		one(atk, false)
	case content.MoveGnaw:
		if hit := one(atk-mv.GnawPenalty, false); hit >= 0 {
			m := &exp.Party[s.slotOf(hit)]
			m.GuardMod = max(m.GuardMod-mv.GnawGuard, mv.ModFloor)
		}
	case content.MoveGorge:
		one(atk+mv.GorgeBonus, false)
	case content.MoveShove:
		if hit := one(atk-mv.ShovePenalty, false); hit >= 0 {
			exp.Party[s.slotOf(hit)].Delayed = true
		}
	case content.MoveBrand:
		if hit := one(atk, false); hit >= 0 {
			exp.Party[s.slotOf(hit)].Burning = mv.BurnDamage
		}
	case content.MoveSurge:
		all(atk-mv.SurgePenalty, false)
	case content.MoveSpore:
		each(func(m *Member) { m.AttackMod = max(m.AttackMod-mv.SporeAttack, mv.ModFloor) })
	case content.MoveMaul:
		one(atk+mv.MaulBonus, false)
	case content.MoveEngulf:
		one(atk+mv.EngulfBonus, false)
	case content.MoveGather:
	case content.MoveDevour:
		if hit := one(atk+mv.DevourBonus, false); hit >= 0 {
			e.HP = min(e.HP+mv.DevourHeal, e.MaxHP)
		}
	case content.MoveWither:
		each(func(m *Member) { m.AttackMod = max(m.AttackMod-mv.WitherAttack, mv.ModFloor) })
	case content.MoveRally:
		for i := range r.Enemies {
			if r.Enemies[i].HP > 0 {
				r.Enemies[i].AttackMod += mv.RallyAttack
			}
		}
	case content.MoveBreakRanks:
		all(atk-mv.BreakRanksPenalty, false)
	case content.MoveCollapse:
		all(mv.CollapseDamage, true)
	case content.MoveSpores:
		each(func(m *Member) { m.GuardMod = max(m.GuardMod-mv.SporesGuard, mv.ModFloor) })
	}
}

// indexOf finds the latest event of a kind about an enemy, since hurting a
// citizen may have appended events after it.
func indexOf(l *log, kind EventKind, n int) int {
	for i := len(l.events) - 1; i >= 0; i-- {
		if l.events[i].Kind == kind && l.events[i].N == n {
			return i
		}
	}
	return len(l.events) - 1
}

// wound takes HP from an enemy. HP never goes below zero.
func (s *State) wound(index, amount int, l *log) int {
	exp, h, r := s.Expedition, s.hollow(), s.room()
	e := &r.Enemies[index]
	dealt := min(max(amount, 0), e.HP)
	e.HP -= dealt
	if e.HP == 0 && dealt > 0 {
		l.add(EvEnemyFell).in(h.District).id(e.Slug(h)).n(index).because(e.Cause)
		if !e.Boss {
			f := content.Enemies[e.Kind].Family
			s.Materials[f]++
			exp.Loot[f]++
			l.add(EvLoot).in(h.District).id(f.Material()).n(1)
		}
	}
	return dealt
}

// act is the acting citizen's move.
func (s *State) act(c Command, l *log) error {
	t := content.T
	exp := s.Expedition
	if exp.Fight == nil {
		return ErrNoFight
	}
	slot := s.Acting()
	if slot < 0 {
		return ErrNoFight
	}
	m := &exp.Party[slot]
	who := &s.Citizens[m.Citizen]
	r := s.room()
	st := who.Stats()
	atk := max(st.Attack+m.AttackMod, 1)

	enemy := func() (*Enemy, error) {
		if c.Target < 0 || c.Target >= len(r.Enemies) || r.Enemies[c.Target].HP <= 0 {
			return nil, fmt.Errorf("%w: no living enemy %d", ErrBadArgument, c.Target)
		}
		return &r.Enemies[c.Target], nil
	}
	ally := func() (*Citizen, error) {
		at := s.slotOf(c.Target)
		if at < 0 || s.Citizens[c.Target].Downed {
			return nil, fmt.Errorf("%w: citizen %d is not standing in this party", ErrBadArgument, c.Target)
		}
		return &s.Citizens[c.Target], nil
	}

	switch c.Action {
	case ActDefend:
		m.Defending = true
		l.add(EvCitizenActed).in(exp.District).who(who).id("defend").n(-1)
	case ActAttack:
		e, err := enemy()
		if err != nil {
			return err
		}
		dealt := s.wound(c.Target, max(atk-e.Guard, 1), l)
		s.note(l, who, "attack", c.Target, dealt)
	case ActAbility:
		if m.Cooldown > 0 {
			return ErrNotReady
		}
		ab := content.Jobs[who.Job].Ability
		slug := content.Abilities[ab].Slug
		switch ab {
		case content.Hold:
			a, err := ally()
			if err != nil {
				return err
			}
			if a.ID == who.ID {
				return fmt.Errorf("%w: hold someone else", ErrBadArgument)
			}
			m.Holding = a.ID
			s.note(l, who, slug, a.ID, 0)
		case content.Mend:
			a, err := ally()
			if err != nil {
				return err
			}
			gain := min(t.Ability.MendHeal, a.Stats().HP-a.HP)
			a.HP += gain
			s.note(l, who, slug, a.ID, gain)
		case content.Hammer:
			e, err := enemy()
			if err != nil {
				return err
			}
			if !e.Boss { // a boss cannot be stunned
				e.Stunned = true
			}
			s.note(l, who, slug, c.Target, s.wound(c.Target, max(atk-e.Guard, 1), l))
		case content.Hook:
			e, err := enemy()
			if err != nil {
				return err
			}
			e.Hooked = true
			s.note(l, who, slug, c.Target, s.wound(c.Target, max(atk-e.Guard, 1), l))
		case content.Grind:
			if _, err := enemy(); err != nil {
				return err
			}
			s.note(l, who, slug, c.Target, s.wound(c.Target, atk+t.Ability.GrindBonus, l))
		case content.Haggle:
			e, err := enemy()
			if err != nil {
				return err
			}
			e.AttackMod = max(e.AttackMod-t.Ability.HaggleAmount, t.Ability.HaggleFloor)
			s.note(l, who, slug, c.Target, 0)
		case content.Heave:
			total := 0
			for i := range r.Enemies {
				if r.Enemies[i].HP > 0 {
					total += s.wound(i, max(atk-t.Ability.HeavePenalty-r.Enemies[i].Guard, 1), l)
				}
			}
			s.note(l, who, slug, -1, total)
		}
		m.Cooldown = t.AbilityCooldown
	default:
		return fmt.Errorf("%w: unknown action %q", ErrBadArgument, c.Action)
	}
	exp.Fight.Turn++
	s.runFight(l)
	return nil
}

// note logs a citizen's action before the events it caused.
func (s *State) note(l *log, who *Citizen, slug string, target, amount int) {
	l.add(EvCitizenActed).in(s.Expedition.District).who(who).id(slug).n(target).v(amount)
}
