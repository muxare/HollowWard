package sim

import "hollowward/internal/content"

// GenerateHollow remakes a district as a hollow. The district's state is the
// whole input:
//
//   - the district type picks the layout;
//   - each need below the threshold breeds its enemy family, more and stronger
//     the lower the need;
//   - each ruined building becomes a hazard room and each well-kept one a haven;
//   - the key figure becomes the boss, twisted by the district's lowest need.
//
// Every room and enemy carries the Cause that produced it. The stream only
// decides where in the hollow each enemy waits.
func GenerateHollow(d *District, night int, rng *Stream) Hollow {
	t := content.T
	layout := content.Layouts[d.ID]
	h := Hollow{
		District:  d.ID,
		Night:     night,
		Neglect:   Neglect(d).Neglect,
		Grievance: d.Grievance(),
		Entry:     0,
		Heart:     len(layout.Rooms) - 1,
	}

	// The district type sets the layout.
	for i, name := range layout.Rooms {
		kind := RoomFight
		switch i {
		case h.Entry:
			kind = RoomEntry
		case h.Heart:
			kind = RoomHeart
		}
		h.Rooms = append(h.Rooms, Room{ID: i, Kind: kind, Name: name, Building: -1, Cause: districtCause(d)})
	}
	for _, e := range layout.Edges {
		h.Rooms[e[0]].Exits = append(h.Rooms[e[0]].Exits, e[1])
	}

	// Buildings pick the special rooms.
	for b := range d.Buildings {
		cond := d.Buildings[b].Condition
		def := content.Districts[d.ID].Buildings[b]
		room := Room{ID: len(h.Rooms), Building: b, Cause: buildingCause(d, b)}
		switch {
		case cond < t.RuinThreshold:
			room.Kind, room.Name = RoomHazard, def.Hazard
			room.Hazard = t.HazardDamageBase + (t.RuinThreshold-cond)/t.HazardDamagePer
		case cond >= t.HavenThreshold:
			room.Kind, room.Name = RoomHaven, def.Haven
		default:
			continue
		}
		// Splice it into every way out of its slot.
		slot := layout.Slots[b]
		room.Exits = h.Rooms[slot].Exits
		h.Rooms[slot].Exits = []int{room.ID}
		h.Rooms = append(h.Rooms, room)
	}
	h.setDepths()

	// Needs pick the enemies.
	var fights []int
	for i := range h.Rooms {
		if h.Rooms[i].Kind == RoomEntry || h.Rooms[i].Kind == RoomFight {
			fights = append(fights, i)
		}
	}
	for n := content.Need(0); n < content.NumNeeds; n++ {
		severity := t.EnemyNeedBelow - d.Needs[n]
		if severity <= 0 {
			continue
		}
		family := content.FamilyOf(n)
		count := t.EnemyCountBase + severity/t.EnemyCountPer
		at := rng.IntN(len(fights))
		for i := 0; i < count; i++ {
			tier := 1
			switch {
			case severity >= t.EnemyTier3At && i == 0:
				tier = 3
			case severity >= t.EnemyTier2At && i%2 == 1:
				tier = 2
			}
			// The next fight room with space.
			for tries := 0; tries < len(fights); tries++ {
				room := &h.Rooms[fights[(at+i+tries)%len(fights)]]
				if len(room.Enemies) < t.MaxEnemiesPerRoom {
					room.Enemies = append(room.Enemies, newEnemy(content.EnemyOf(family, tier), severity, needCause(d, n)))
					break
				}
			}
		}
	}

	// The key figure is the boss, twisted by their grievance.
	heart := &h.Rooms[h.Heart]
	below := t.NeedMax - d.Needs[h.Grievance]
	st := t.BossStats[d.ID]
	st.HP += below / t.BossHPPer
	st.Attack += below / t.BossAttackPer
	heart.Enemies = append(heart.Enemies, Enemy{
		Kind: -1, Boss: true, Name: content.BossName(d.ID, h.Grievance),
		HP: st.HP, MaxHP: st.HP, Attack: st.Attack, Guard: st.Guard, Speed: st.Speed,
		Cause: figureCause(d),
	})
	if severity := t.EnemyNeedBelow - d.Needs[h.Grievance]; severity > 0 {
		escorts := min(severity/t.BossEscortPer, t.BossEscortMax)
		for i := 0; i < escorts; i++ {
			heart.Enemies = append(heart.Enemies,
				newEnemy(content.EnemyOf(content.FamilyOf(h.Grievance), 1), severity, needCause(d, h.Grievance)))
		}
	}
	return h
}

func newEnemy(kind content.EnemyID, severity int, cause Cause) Enemy {
	t := content.T
	def := content.Enemies[kind]
	st := def.Stats()
	st.HP += severity / t.EnemyHPPer
	st.Attack += severity / t.EnemyAttackPer
	return Enemy{
		Kind: kind, Name: def.Name,
		HP: st.HP, MaxHP: st.HP, Attack: st.Attack, Guard: st.Guard, Speed: st.Speed,
		Cause: cause,
	}
}

// setDepths gives every room the length of the longest way to it from the
// entry. The graph has no cycles: every edge leads deeper.
func (h *Hollow) setDepths() {
	var walk func(room, depth int)
	walk = func(room, depth int) {
		if depth < h.Rooms[room].Depth {
			return
		}
		h.Rooms[room].Depth = depth
		for _, next := range h.Rooms[room].Exits {
			walk(next, depth+1)
		}
	}
	walk(h.Entry, 0)
}
