package sim

import (
	"fmt"

	"hollowward/internal/content"
)

// What can happen from dusk to dawn.
const (
	EvTurned       EventKind = "turned"        // District; N: Neglect; Causes: why, largest first
	EvQuietNight   EventKind = "quiet-night"   // nothing turned
	EvPartySent    EventKind = "party-sent"    // District; N: party size
	EvJoined       EventKind = "joined"        // Citizen: went in
	EvStayedHome   EventKind = "stayed-home"   // nobody went in
	EvRoomEntered  EventKind = "room-entered"  // District; ID: room name; N: room; Causes: why the room exists
	EvMet          EventKind = "met"           // District; ID: enemy slug; N: how many; V: 1 for a boss; Causes: why
	EvHazard       EventKind = "hazard"        // District; ID: building slug; N: HP lost by each; Causes: why
	EvHaven        EventKind = "haven"         // District; ID: building slug; N: HP healed in all; Causes: why
	EvFightStarted EventKind = "fight-started" // N: room
	EvRound        EventKind = "round"         // N: round
	EvCitizenActed EventKind = "citizen-acted" // Citizen; ID: action or ability slug; N: target; V: damage or healing
	EvEnemyActed   EventKind = "enemy-acted"   // ID: move name; N: enemy index; Citizen: target or -1; V: damage
	EvEnemyFell    EventKind = "enemy-fell"    // District; ID: enemy slug; N: enemy index
	EvDowned       EventKind = "downed"        // Citizen
	EvRoomCleared  EventKind = "room-cleared"  // District; ID: room name; N: room
	EvLoot         EventKind = "loot"          // ID: material or resource; N: amount
	EvHeartCleared EventKind = "heart-cleared" // District; ID: boss slug
	EvRestored     EventKind = "restored"      // District: it is a district again
	EvWentHome     EventKind = "went-home"     // District: still turned
	EvWiped        EventKind = "wiped"         // District: the whole party went down
	EvDrivenOut    EventKind = "driven-out"    // District: the fight ran to the round cap
	EvDied         EventKind = "died"          // Citizen
	EvScarred      EventKind = "scarred"       // Citizen; ID: scar slug
	EvExhausted    EventKind = "exhausted"     // Citizen: works less tomorrow
	EvNightEnded   EventKind = "night-ended"
)

// DuskReport is what happened at dusk tonight.
type DuskReport struct {
	Turned  content.DistrictID `json:"turned"` // NoDistrict on a quiet night
	Neglect int                `json:"neglect"`
	Causes  []NeglectCause     `json:"causes"`
	Before  int                `json:"before"` // the Dreamer before tonight, in tenths
	After   int                `json:"after"`  // and after
	Rise    []Rise             `json:"rise"`   // why it rose
	Line    string             `json:"line"`   // what the Dreamer says tonight
}

// Rise is one reason the Dreamer rose tonight.
type Rise struct {
	Reason   RiseReason         `json:"reason"`
	District content.DistrictID `json:"district"`
	Tenths   int                `json:"tenths"`
}

// RiseReason says why the Dreamer rose.
type RiseReason string

// The reasons.
const (
	RiseTurned RiseReason = "turned" // a district turned tonight: its Neglect
	RiseQuiet  RiseReason = "quiet"  // nothing turned: it stirs anyway
	RiseOpen   RiseReason = "open"   // a district is still turned from an earlier night
)

// Expedition is a party inside a hollow.
type Expedition struct {
	District content.DistrictID       `json:"district"`
	Party    []Member                 `json:"party"`
	Room     int                      `json:"room"`
	Fight    *Fight                   `json:"fight"`
	Loot     [content.NumFamilies]int `json:"loot"` // materials taken tonight
}

// Member is a citizen on an expedition, with what the fight has done to them.
type Member struct {
	Citizen   int  `json:"citizen"`
	Fought    bool `json:"fought"`     // was in a fight tonight
	AttackMod int  `json:"attack_mod"` // lasts the fight
	GuardMod  int  `json:"guard_mod"`  // lasts the fight
	Delayed   bool `json:"delayed"`    // acts last next round
	Burning   int  `json:"burning"`    // HP lost at the start of their next turn
	Defending bool `json:"defending"`  // until their next turn
	Holding   int  `json:"holding"`    // the citizen they guard until their next turn, or -1
	Cooldown  int  `json:"cooldown"`   // rounds until the job ability is ready
}

func (e Expedition) clone() Expedition {
	n := e
	n.Party = append([]Member(nil), e.Party...)
	if e.Fight != nil {
		f := *e.Fight
		f.Order = append([]Actor(nil), e.Fight.Order...)
		n.Fight = &f
	}
	return n
}

// dusk falls after the day tick: the most neglected district turns, if it is
// neglected enough.
func (s *State) dusk(l *log) {
	t := content.T
	s.Phase = PhaseDusk
	rep := DuskReport{Turned: content.NoDistrict}
	worst := -1
	var tied []content.DistrictID
	for id := range s.Districts {
		d := &s.Districts[id]
		if d.Turned {
			continue // it is already a hollow
		}
		switch n := Neglect(d).Neglect; {
		case n > worst:
			worst, tied = n, []content.DistrictID{d.ID}
		case n == worst:
			tied = append(tied, d.ID)
		}
	}
	if worst >= t.TurnThreshold {
		pick := tied[0]
		if len(tied) > 1 {
			pick = tied[s.RNG.Dusk.IntN(len(tied))]
		}
		d := &s.Districts[pick]
		neglect := Neglect(d)
		rep.Turned, rep.Neglect, rep.Causes = pick, neglect.Neglect, neglect.Causes
		h := GenerateHollow(d, s.Day, &s.RNG.Gen)
		s.Hollows[pick] = &h
		d.Turned = true
		for i := range s.Stewards {
			if s.Stewards[i].District == pick {
				s.Stewards[i] = Post{District: content.NoDistrict, Need: content.NoNeed, Building: -1}
			}
		}
		ev := l.add(EvTurned).in(pick).n(neglect.Neglect)
		for _, c := range neglect.Causes {
			ev.because(c.Cause)
		}
	} else {
		l.add(EvQuietNight)
	}
	s.Dusk = &rep
	s.duskClock(l)
}

func (s *State) sendParty(c Command, l *log) error {
	t := content.T
	if !c.District.Valid() || s.Hollows[c.District] == nil {
		return fmt.Errorf("%w: there is no hollow there", ErrBadArgument)
	}
	if len(c.Citizens) < 1 || len(c.Citizens) > t.MaxParty {
		return fmt.Errorf("%w: a party is 1 to %d citizens", ErrBadArgument, t.MaxParty)
	}
	seen := map[int]bool{}
	for _, id := range c.Citizens {
		if id < 0 || id >= len(s.Citizens) || !s.Citizens[id].Alive || seen[id] {
			return fmt.Errorf("%w: citizen %d cannot go", ErrBadArgument, id)
		}
		seen[id] = true
	}
	exp := &Expedition{District: c.District}
	l.add(EvPartySent).in(c.District).n(len(c.Citizens))
	for _, id := range c.Citizens {
		exp.Party = append(exp.Party, Member{Citizen: id, Holding: -1})
		l.add(EvJoined).in(c.District).who(&s.Citizens[id])
	}
	s.Expedition = exp
	s.Phase = PhaseNight
	s.enter(s.Hollows[c.District].Entry, l)
	return nil
}

func (s *State) stayHome(l *log) error {
	l.add(EvStayedHome)
	s.endNight(l)
	return nil
}

// hollow is the hollow the party is in.
func (s *State) hollow() *Hollow { return s.Hollows[s.Expedition.District] }

// room is the room the party is in.
func (s *State) room() *Room { return &s.hollow().Rooms[s.Expedition.Room] }

// standing lists the party slots of citizens still on their feet.
func (s *State) standing() []int {
	var out []int
	for i, m := range s.Expedition.Party {
		if c := &s.Citizens[m.Citizen]; c.Alive && !c.Downed {
			out = append(out, i)
		}
	}
	return out
}

// enter moves the party into a room and resolves what waits there.
func (s *State) enter(room int, l *log) {
	t := content.T
	exp, h := s.Expedition, s.hollow()
	exp.Room = room
	r := &h.Rooms[room]
	r.Visited = true
	l.add(EvRoomEntered).in(h.District).id(r.Name).n(room).because(r.Cause)

	switch r.Kind {
	case RoomHazard:
		if !r.Cleared { // a hazard only catches a party once
			slug := content.Districts[h.District].Buildings[r.Building].Slug
			l.add(EvHazard).in(h.District).id(slug).n(r.Hazard).because(r.Cause)
			for _, slot := range s.standing() {
				s.hurt(slot, r.Hazard, l)
			}
			r.Cleared = true
		}
		if len(s.standing()) == 0 {
			l.add(EvWiped).in(h.District)
			s.endNight(l)
		}
	case RoomHaven:
		slug := content.Districts[h.District].Buildings[r.Building].Slug
		healed := 0
		for _, slot := range s.standing() {
			c := &s.Citizens[exp.Party[slot].Citizen]
			top := c.Stats().HP
			gain := min((top*t.HavenHealPct+99)/100, top-c.HP)
			c.HP += gain
			healed += gain
		}
		l.add(EvHaven).in(h.District).id(slug).n(healed).because(r.Cause)
		r.Cleared = true
		s.searchHaven(r, l)
	default:
		if r.Alive() == 0 {
			r.Cleared = true
			return
		}
		// Log what the party meets, kind by kind, with why it exists.
		var order []string
		count, cause, isBoss := map[string]int{}, map[string]Cause{}, map[string]int{}
		for i := range r.Enemies {
			e := &r.Enemies[i]
			if e.HP <= 0 {
				continue
			}
			slug := e.Slug(h)
			if count[slug] == 0 {
				order = append(order, slug)
				cause[slug] = e.Cause
				if e.Boss {
					isBoss[slug] = 1
				}
			}
			count[slug]++
		}
		for _, slug := range order {
			l.add(EvMet).in(h.District).id(slug).n(count[slug]).v(isBoss[slug]).because(cause[slug])
		}
		s.startFight(l)
	}
}

func (s *State) advance(c Command, l *log) error {
	if s.Expedition.Fight != nil {
		return ErrInFight
	}
	r := s.room()
	for _, x := range r.Exits {
		if x == c.Room {
			s.enter(c.Room, l)
			return nil
		}
	}
	return fmt.Errorf("%w: room %d is not a way deeper from %q", ErrBadArgument, c.Room, r.Name)
}

func (s *State) goHome(l *log) error {
	if s.Expedition.Fight != nil {
		return ErrInFight
	}
	l.add(EvWentHome).in(s.Expedition.District)
	s.endNight(l)
	return nil
}

// hurt takes HP from the party member in a slot. HP never goes below zero; at
// zero the citizen is Downed.
func (s *State) hurt(slot, amount int, l *log) int {
	c := &s.Citizens[s.Expedition.Party[slot].Citizen]
	dealt := min(max(amount, 0), c.HP)
	c.HP -= dealt
	if c.HP == 0 && !c.Downed {
		c.Downed = true
		l.add(EvDowned).in(s.Expedition.District).who(c)
	}
	return dealt
}

// clearHeart ends the hollow: the district is a district again.
func (s *State) clearHeart(l *log) {
	t := content.T
	h := s.hollow()
	d := &s.Districts[h.District]
	l.add(EvHeartCleared).in(h.District).id(content.BossSlug(h.District, h.Grievance))
	family := content.FamilyOf(h.Grievance)
	s.Materials[family] += t.HeartMaterials
	s.Expedition.Loot[family] += t.HeartMaterials
	l.add(EvLoot).in(h.District).id(family.Material()).n(t.HeartMaterials)
	s.Resources[content.Coin] += t.HeartCoin
	l.add(EvLoot).in(h.District).id(content.Coin.String()).n(t.HeartCoin)
	for n := range d.Needs {
		d.Needs[n] = max(d.Needs[n], t.ClearRestoreNeed)
	}
	d.Turned = false
	s.Hollows[h.District] = nil
	l.add(EvRestored).in(d.ID)
	if t.DreamerSootheOnClear > 0 {
		s.Dreamer = max(s.Dreamer-t.DreamerSootheOnClear, 0)
		l.add(EvDreamerSoothed).n(t.DreamerSootheOnClear).v(s.Dreamer)
	}
	s.endNight(l)
}

// endNight brings everyone home: the Downed live or die, the living are
// tired, and the next day begins.
func (s *State) endNight(l *log) {
	t := content.T
	if exp := s.Expedition; exp != nil {
		for _, m := range exp.Party {
			c := &s.Citizens[m.Citizen]
			if c.Downed {
				chance := t.DownedDeathPct
				if c.HasTrait(func(tr content.TraitDef) bool { return tr.Lucky }) {
					chance = t.LuckyDeathPct
				}
				c.Downed = false
				if s.RNG.Fate.Chance(chance) {
					c.Alive = false
					l.add(EvDied).in(exp.District).who(c)
					continue
				}
				scar := content.ScarID(s.RNG.Fate.IntN(int(content.NumScars)))
				c.Scars = append(c.Scars, scar)
				l.add(EvScarred).in(exp.District).who(c).id(content.Scars[scar].Slug)
			}
			if m.Fought {
				c.Exhausted = true
				l.add(EvExhausted).who(c)
			}
		}
		if exp.Fight != nil { // driven out or wiped mid-fight: the room keeps its wounds, not its moods
			s.resetRoom(s.room())
		}
	}
	for i := range s.Citizens {
		if c := &s.Citizens[i]; c.Alive {
			c.HP = c.Stats().HP // a night's sleep
		}
	}
	s.Expedition, s.Dusk = nil, nil
	l.add(EvNightEnded)
	if s.dawn(l) {
		return
	}
	s.Day++
	s.startDay(l)
}

// resetRoom clears what a broken-off fight left on the enemies of a room.
func (s *State) resetRoom(r *Room) {
	for i := range r.Enemies {
		e := &r.Enemies[i]
		e.Stunned, e.Hooked, e.Acted, e.AttackMod, e.NextMove = false, false, false, 0, 0
		e.Intent, e.Next = Intent{Target: -1}, Intent{Target: -1}
	}
}
