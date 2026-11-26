package game

import (
	"fmt"
	"strings"

	"graphics.gd/classdb/Control"
	"graphics.gd/classdb/Line2D"
	"graphics.gd/variant/Color"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

// --- Dusk ---

func (a *App) duskScreen() node {
	s := &a.game.State
	t := content.T
	rep := s.Dusk

	// What the Dreamer did, and why.
	dream := vbox(4, para("“"+rep.Line+"”", 19, colDream, 360),
		text(fmt.Sprintf("The Dreamer rises from %s to %s of %d.", tenths(rep.Before), tenths(rep.After), t.DreamerMax/10), 15, colText))
	for _, r := range rep.Rise {
		add(dream, para("  "+riseText(r), 13, colDim, 360))
	}

	// What turned, and why.
	turned := vbox(3)
	if rep.Turned.Valid() {
		add(turned, text("THE "+strings.ToUpper(rep.Turned.String())+" "+rep.Turned.Verb("HAS", "HAVE")+" TURNED", 20, colBlood),
			para(fmt.Sprintf("It was the most neglected district, at Neglect %d. Why, largest first:", rep.Neglect), 14, colText, 360))
		for _, c := range rep.Causes {
			add(turned, text("  "+causeLine(c), 14, colText))
		}
	} else {
		add(turned, text("A QUIET NIGHT", 20, colGood),
			para(fmt.Sprintf("No district's Neglect reached %d. Nothing turns.", t.TurnThreshold), 14, colText, 360))
	}

	// The open hollows: pick one.
	hollows := vbox(6, text("Open hollows. One party may go in tonight.", 14, colDim))
	open := 0
	for id, h := range s.Hollows {
		if h == nil {
			continue
		}
		open++
		d := content.DistrictID(id)
		if !a.target.Valid() || s.Hollows[a.target] == nil {
			a.target = d
		}
		label := fmt.Sprintf("The %s (turned night %d, %d rooms)", d, h.Night, len(h.Rooms))
		pick := a.button("hollow:"+d.Slug(), label, "Choose this hollow for tonight's party.", true, func() { a.target = d })
		border := colLine
		if d == a.target {
			border = colGold
		}
		lines := vbox(1, pick, text("The Ledger says:", 12, colDim))
		for _, p := range sim.Preview(&a.ledger, h) {
			add(lines, a.previewLine(p))
		}
		add(hollows, framed(border, lines))
	}
	if open == 0 {
		add(hollows, text("None. Everyone sleeps.", 14, colGood))
	}

	// The party.
	picked := map[int]bool{}
	var clean []int
	for _, id := range a.party {
		if id < len(s.Citizens) && s.Citizens[id].Alive && !picked[id] && len(clean) < t.MaxParty {
			picked[id] = true
			clean = append(clean, id)
		}
	}
	a.party = clean
	people := grid(2, 4)
	for i := range s.Citizens {
		c := &s.Citizens[i]
		if !c.Alive {
			continue
		}
		label := fmt.Sprintf("%s, %s (%s)   %s", c.Name, content.Jobs[c.Job].Name, c.Home, statsText(c))
		if picked[c.ID] {
			label = "[in] " + label
		}
		can := open > 0 && (picked[c.ID] || len(a.party) < t.MaxParty)
		id := c.ID
		b := a.button(fmt.Sprintf("pick:%d", id), label, traitsTip(c)+"\nWhoever fights tonight works at half strength tomorrow.", can, func() {
			if picked[id] {
				var rest []int
				for _, p := range a.party {
					if p != id {
						rest = append(rest, p)
					}
				}
				a.party = rest
			} else {
				a.party = append(a.party, id)
			}
		})
		if picked[c.ID] {
			b.AsControl().AddThemeColorOverride("font_color", colGold)
		}
		add(people, sized(b, 590, 0))
	}
	roster := vbox(4, text(fmt.Sprintf("Party: %d of %d picked. They work at half strength tomorrow; the Downed may die.", len(a.party), t.MaxParty), 14, colDim),
		scroll(1210, 205, people))

	var go1 node
	if open > 0 && a.target.Valid() {
		c := sim.SendParty(a.target, a.party...)
		ok, why := a.offers(c)
		if len(a.party) == 0 {
			why = "Pick 1 to 4 citizens first."
		}
		go1 = a.button(actionID(c), "Send the party into the "+a.target.String(), why, ok, func() { a.send(c) })
	} else {
		go1 = gap(0, 0)
	}
	stay := "Let the night pass"
	if open > 0 {
		stay = "Send nobody tonight"
	}
	return vbox(8,
		a.topBar(),
		a.hint("a turned district is rebuilt from what it lacked. Pick up to four citizens and send them in, or the hollow stays open and feeds the Dreamer."),
		hbox(12, panel(scroll(380, 230, vbox(12, dream, turned))), panel(scroll(800, 230, hollows))),
		panel(roster),
		a.footer(go1, a.command(stay, fmt.Sprintf("Nobody goes into a hollow tonight. Each open hollow adds %s to the Dreamer at every dusk.", tenths(t.DreamerPerOpenHollow)), sim.StayHome())),
	)
}

// --- The node map ---

func roomColour(r *sim.Room, here bool) Color.RGBA {
	c := colLine
	switch r.Kind {
	case sim.RoomHeart:
		c = colBlood
	case sim.RoomHazard:
		c = colWarn
	case sim.RoomHaven:
		c = colGood
	case sim.RoomEntry, sim.RoomFight:
		c = colDim
	}
	if here {
		c = colGold
	}
	return c
}

// roomShape draws a room on the map: a square for an ordinary room, a diamond
// for a hazard, a rounded shape for a safe room, a spike for the heart.
func roomShape(r *sim.Room, col Color.RGBA) Control.Instance {
	switch r.Kind {
	case sim.RoomHazard:
		return shape(30, 30, col, pt(15, 0), pt(30, 15), pt(15, 30), pt(0, 15))
	case sim.RoomHaven:
		return shape(30, 30, col, pt(8, 0), pt(22, 0), pt(30, 8), pt(30, 22), pt(22, 30), pt(8, 30), pt(0, 22), pt(0, 8))
	case sim.RoomHeart:
		return shape(30, 30, col, pt(15, 0), pt(20, 10), pt(30, 12), pt(22, 20), pt(25, 30), pt(15, 24), pt(5, 30), pt(8, 20), pt(0, 12), pt(10, 10))
	}
	return shape(30, 30, col, pt(2, 2), pt(28, 2), pt(28, 28), pt(2, 28))
}

// nodeMap draws the hollow's graph, shallow rooms on the left.
func (a *App) nodeMap(h *sim.Hollow, here int, w, ht int) node {
	canvas := gap(w, ht)
	depth := 0
	perDepth := map[int][]int{}
	for i := range h.Rooms {
		depth = max(depth, h.Rooms[i].Depth)
		perDepth[h.Rooms[i].Depth] = append(perDepth[h.Rooms[i].Depth], i)
	}
	pos := make([][2]float64, len(h.Rooms))
	for d, rooms := range perDepth {
		for k, i := range rooms {
			x := 50 + float64(d)*float64(w-110)/float64(max(depth, 1))
			y := float64(ht) * float64(k+1) / float64(len(rooms)+1)
			pos[i] = [2]float64{x, y}
		}
	}
	for i := range h.Rooms {
		for _, x := range h.Rooms[i].Exits {
			line := Line2D.New()
			line.AddPoint(pt(pos[i][0], pos[i][1]))
			line.AddPoint(pt(pos[x][0], pos[x][1]))
			line.SetWidth(2)
			col := colLine
			if i == here {
				col = colGold
			}
			line.SetDefaultColor(col)
			add(canvas, line)
		}
	}
	for i := range h.Rooms {
		r := &h.Rooms[i]
		col := roomColour(r, i == here)
		if r.Cleared && i != here {
			col.A = 0.45
		}
		box := roomShape(r, col)
		box.SetPosition(pt(pos[i][0]-15, pos[i][1]-15))
		box.SetMouseFilter(Control.MouseFilterPass)
		tip(box, roomTip(r))
		name := text(r.Name, 11, colDim)
		name.AsControl().SetPosition(pt(max(pos[i][0]-50, 4), pos[i][1]+17))
		name.AsControl().SetMouseFilter(Control.MouseFilterIgnore)
		add(canvas, box, name)
	}
	return canvas
}

func (a *App) partyList(compact bool) node {
	s := &a.game.State
	exp := s.Expedition
	acting := s.Acting()
	box := vbox(4)
	for slot := range exp.Party {
		m := &exp.Party[slot]
		c := &s.Citizens[m.Citizen]
		top := c.Stats().HP
		col := colText
		head := fmt.Sprintf("%s, %s", c.Name, content.Jobs[c.Job].Name)
		switch {
		case c.Downed:
			col, head = colBlood, head+"  DOWNED"
		case slot == acting:
			col, head = colGold, "▶ "+head
		}
		row := vbox(2, hbox(8, sized(text(head, 15, col), 210, 0), meter(c.HP, top, colGood, 120, 12), text(statsText(c), 13, colDim)))
		if st := memberStatus(m); st != "" && !compact {
			add(row, text("   "+st, 12, colWarn))
		}
		add(box, tip(row, traitsTip(c)))
	}
	return box
}

func (a *App) mapScreen() node {
	s := &a.game.State
	exp := s.Expedition
	h := s.Hollows[exp.District]
	r := &h.Rooms[exp.Room]

	here := vbox(4, text(strings.ToUpper(r.Name), 18, colGold), tip(text("Why this room exists: "+r.Cause.Describe()+".", 14, colText), roomTip(r)))
	switch r.Kind {
	case sim.RoomHazard:
		add(here, text(fmt.Sprintf("A hazard. It cost everyone %d HP.", r.Hazard), 14, colWarn))
	case sim.RoomHaven:
		add(here, text("A safe room. The party has rested.", 14, colGood))
	}
	for _, g := range r.Groups() {
		add(here, text("Dead here: "+g.String(), 13, colDim))
	}
	for _, e := range a.recent {
		switch e.Kind {
		case sim.EvHazard, sim.EvHaven, sim.EvFragment, sim.EvLoot, sim.EvRoomCleared, sim.EvDowned:
			add(here, para(strings.TrimSpace(e.Text()), 13, colText, 480))
		}
	}

	ways := vbox(6, text("Go deeper, or go home.", 14, colDim))
	for _, x := range r.Exits {
		nx := &h.Rooms[x]
		label := "Deeper: " + nx.Name
		switch {
		case nx.Kind == sim.RoomHeart:
			label += " (the heart)"
		case nx.Kind == sim.RoomHazard && !nx.Cleared:
			label += fmt.Sprintf(" (hazard: -%d HP each)", nx.Hazard)
		case nx.Kind == sim.RoomHaven:
			label += " (safe room)"
		case nx.Cleared:
			label += " (cleared)"
		}
		add(ways, tip(a.command(label, "", sim.Advance(x)), roomTip(nx)))
	}
	add(ways, a.command("Go home", "The district stays turned. Cleared rooms stay cleared.", sim.GoHome()))

	return vbox(8,
		a.topBar(),
		text("THE HOLLOW OF THE "+strings.ToUpper(exp.District.String()), 20, colBlood),
		a.hint("squares are rooms, diamonds hazards, round ones safe rooms, the spiked one the heart. Hover over a room to see why it exists."),
		hbox(12, panel(a.nodeMap(h, exp.Room, 680, 330)), panel(sized(vbox(10, here, ways), 500, 330))),
		panel(a.partyList(true)),
		a.footer(),
	)
}

// --- Combat ---

func (a *App) combatScreen() node {
	s := &a.game.State
	exp := s.Expedition
	h := s.Hollows[exp.District]
	r := &h.Rooms[exp.Room]
	acting := s.Acting()
	me := &s.Citizens[exp.Party[acting].Citizen]
	ability := content.Abilities[content.Jobs[me.Job].Ability]
	ready := exp.Party[acting].Cooldown == 0
	onAlly := ability.ID == content.Hold || ability.ID == content.Mend
	onAll := ability.ID == content.Heave

	// Enemies, each with its cause and what it has announced.
	foes := vbox(6)
	for i := range r.Enemies {
		e := &r.Enemies[i]
		if e.HP <= 0 {
			add(foes, text(e.Name+": dead", 13, colDim))
			continue
		}
		col := colText
		if e.Boss {
			col = colBlood
		}
		card := vbox(2,
			hbox(8, sized(text(e.Name, 16, col), 250, 0), meter(e.HP, e.MaxHP, colBlood, 120, 12),
				text(fmt.Sprintf("HP %d/%d  Atk %d  Grd %d  Spd %d", e.HP, e.MaxHP, max(e.Attack+e.AttackMod, 1), e.Guard, e.Speed), 13, colDim)),
			text("Why it exists: "+e.Cause.Describe()+".", 13, colHunger),
			para(intentText(s, e), 14, colWarn, 620),
		)
		if st := enemyStatus(e); st != "" {
			add(card, text(st, 12, colGood))
		}
		acts := hbox(6, a.command("Strike", "A plain blow: Attack minus its Guard.", sim.Act(sim.ActAttack, i)))
		if !onAlly && !onAll {
			add(acts, a.command(ability.Name, ability.Text, sim.Act(sim.ActAbility, i)))
		}
		add(card, acts)
		add(foes, tip(framed(colLine, card), "Why it exists: "+e.Cause.Describe()+"."))
	}

	// The party, with ally-targeted abilities on each card.
	friends := vbox(6, a.partyList(false))
	mine := hbox(6, text(me.Name+"'s turn.", 15, colGold), a.command("Brace", fmt.Sprintf("Take %d less from every blow until your next turn.", content.T.DefendGuard), sim.Act(sim.ActDefend, -1)))
	if onAll {
		add(mine, a.command(ability.Name, ability.Text, sim.Act(sim.ActAbility, -1)))
	}
	if onAlly {
		for _, m := range exp.Party {
			c := &s.Citizens[m.Citizen]
			if c.Downed || (ability.ID == content.Hold && c.ID == me.ID) {
				continue
			}
			add(mine, a.command(ability.Name+": "+c.Name, ability.Text, sim.Act(sim.ActAbility, c.ID)))
		}
	}
	if !ready {
		add(mine, text(ability.Name+" is resting this turn.", 12, colDim))
	}
	add(friends, mine)

	// What just happened.
	logBox := vbox(1)
	lines := 0
	for i := len(a.recent) - 1; i >= 0 && lines < 7; i-- {
		if l := strings.TrimSpace(a.recent[i].Text()); l != "" {
			logBox.AsNode().AddChild(text(l, 13, colDim).AsNode())
			logBox.AsNode().MoveChild(logBox.AsNode().GetChild(lines), 0)
			lines++
		}
	}

	return vbox(8,
		a.topBar(),
		text(fmt.Sprintf("%s, IN THE HOLLOW OF THE %s. ROUND %d OF %d", strings.ToUpper(r.Name), strings.ToUpper(exp.District.String()), exp.Fight.Round, content.T.RoundCap), 18, colBlood),
		a.hint("every enemy shows what it will do next, and after that. Fastest acts first. A citizen at 0 HP is Downed and may not wake."),
		hbox(12, panel(scroll(660, 440, foes)), vbox(8, panel(sized(friends, 540, 0)), panel(scroll(540, 170, logBox)))),
		a.footer(),
	)
}
