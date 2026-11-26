package game

import (
	"fmt"
	"strings"

	"graphics.gd/classdb/Control"
	"graphics.gd/variant/Color"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

// topBar is the strip every in-tenure screen starts with: the day, the
// Dreamer, the stores.
func (a *App) topBar() node {
	s := &a.game.State
	t := content.T
	dream := meter(s.Dreamer, t.DreamerMax, colDream, 240, 16)
	tip(dream, fmt.Sprintf("The Dreamer: %s of %d.\nIt rises every dusk: by the turned district's Neglect over ten, or by %s on a quiet night, plus %s for every district still turned from an earlier night.\nAt %d the lid breaks and the tenure ends.",
		tenths(s.Dreamer), t.DreamerMax/10, tenths(t.DreamerQuiet), tenths(t.DreamerPerOpenHollow), t.DreamerMax/10))
	stores := text(fmt.Sprintf("Coin %d   Grain %d   Timber %d", s.Resources[content.Coin], s.Resources[content.Grain], s.Resources[content.Timber]), 15, colText)
	tip(stores, "Districts make these each day, by how many residents are working and how well the district's needs are met.\nThey pay for decrees and for some answers to events.")
	mats := text(fmt.Sprintf("Hungerbone %d   Ash-masks %d   Blackcap %d", s.Materials[content.Hunger], s.Materials[content.Riot], s.Materials[content.Rot]), 15, colHunger)
	tip(mats, "Dropped in the hollows: hungerbone by Hunger (bred by low Supply), ash-masks by Riot (low Order), blackcap by Rot (low Upkeep).\nThey buy upgrades. To get more of one, let its need fall.")
	return panel(hbox(18,
		text(fmt.Sprintf("Night %d of %d", s.Day, t.Nights), 17, colGold),
		hbox(6, text("Dreamer "+tenths(s.Dreamer)+" / "+fmt.Sprint(t.DreamerMax/10), 15, colDream), dream),
		stores, mats, spacer(),
		a.button("ledger", "Ledger", "What earlier wardens learned: which neglect breeds which horror.", true, func() { a.overlay = "ledger" }),
	))
}

// footer shows why the last command was refused, if it was.
func (a *App) footer(kids ...node) node {
	row := hbox(12)
	if a.message != "" {
		add(row, text(a.message, 14, colWarn))
	}
	add(row, spacer())
	add(row, kids...)
	return row
}

func neglectColour(n int) Color.RGBA {
	switch {
	case n >= content.T.TurnThreshold:
		return colBlood
	case n >= content.T.TurnThreshold-10:
		return colWarn
	}
	return colGood
}

func needColour(v int) Color.RGBA {
	switch {
	case v < content.T.EnemyNeedBelow:
		return colBlood
	case v < content.T.EnemyNeedBelow+15:
		return colWarn
	}
	return colGood
}

// districtIcon is the district's placeholder art: one polygon each.
func districtIcon(d content.DistrictID, col Color.RGBA) Control.Instance {
	switch d {
	case content.Market: // an awning
		return shape(28, 28, col, pt(2, 11), pt(6, 3), pt(22, 3), pt(26, 11), pt(23, 11), pt(23, 26), pt(5, 26), pt(5, 11))
	case content.Watch: // a tower
		return shape(28, 28, col, pt(7, 26), pt(7, 6), pt(11, 6), pt(11, 2), pt(17, 2), pt(17, 6), pt(21, 6), pt(21, 26))
	case content.Docks: // a hull
		return shape(28, 28, col, pt(2, 12), pt(26, 12), pt(21, 24), pt(7, 24))
	case content.Temple: // a pediment
		return shape(28, 28, col, pt(14, 2), pt(26, 12), pt(22, 12), pt(22, 26), pt(6, 26), pt(6, 12), pt(2, 12))
	case content.Mills: // a wheel
		return shape(28, 28, col, pt(10, 2), pt(18, 2), pt(26, 10), pt(26, 18), pt(18, 26), pt(10, 26), pt(2, 18), pt(2, 10))
	}
	// Tenements: stacked floors
	return shape(28, 28, col, pt(3, 26), pt(3, 12), pt(10, 12), pt(10, 6), pt(18, 6), pt(18, 2), pt(25, 2), pt(25, 26))
}

// neglectTip is the Neglect meter's hover text: the causes, largest first.
func neglectTip(d *sim.District) string {
	rep := sim.Neglect(d)
	var b strings.Builder
	fmt.Fprintf(&b, "Neglect %d = 100 minus the mean of the three needs, plus %d for each ruined building.\nAt dusk the most neglected district turns, if its Neglect is %d or more.\n", rep.Neglect, content.T.NeglectPerRuin, content.T.TurnThreshold)
	if len(rep.Causes) > 0 {
		b.WriteString("\nWhy, largest first:")
	}
	for _, c := range rep.Causes {
		b.WriteString("\n" + causeLine(c))
	}
	return b.String()
}

func (a *App) districtCard(d *sim.District) node {
	s := &a.game.State
	rep := sim.Neglect(d)
	col := neglectColour(rep.Neglect)
	head := hbox(6, districtIcon(d.ID, col),
		a.button("select:"+d.ID.Slug(), d.ID.String(), "Open the "+d.ID.String()+".", true, func() { a.selected = d.ID }))
	var posted []string
	for i, p := range s.Stewards {
		if p.District == d.ID {
			posted = append(posted, strings.TrimPrefix(content.StewardNames[i], "Steward "))
		}
	}
	if len(posted) > 0 {
		add(head, text(strings.Join(posted, " "), 12, colGold))
	}
	body := vbox(3, head)
	if d.Turned {
		add(body, text("TURNED", 16, colBlood), para("A hollow since night "+fmt.Sprint(s.Hollows[d.ID].Night)+". It makes nothing and feeds the Dreamer every dusk.", 12, colDim, 170))
	} else {
		nm := meter(rep.Neglect, 100, col, 110, 12)
		add(body, tip(hbox(6, text(fmt.Sprintf("Neglect %d", rep.Neglect), 14, col), nm), neglectTip(d)))
		for n := content.Need(0); n < content.NumNeeds; n++ {
			v := d.Needs[n]
			add(body, hbox(6, sized(text(n.String(), 12, colDim), 50, 0), meter(v, 100, needColour(v), 90, 8), text(fmt.Sprint(v), 12, needColour(v))))
		}
	}
	border := colLine
	switch {
	case d.ID == a.selected:
		border = colGold
	case d.Turned:
		border = colBlood
	}
	return sized(framed(border, body), 198, 128)
}

// wouldBecome lists what a district would turn into tonight, as far as the
// Ledger can name it. It is the same mapping the generator uses, shown ahead
// of time.
func (a *App) wouldBecome(d *sim.District) []sim.PreviewLine {
	throwaway := sim.NewStream(0, 0)
	h := sim.GenerateHollow(d, a.game.State.Day, &throwaway)
	return sim.Preview(&a.ledger, &h)
}

func (a *App) districtPanel() node {
	s := &a.game.State
	t := content.T
	d := &s.Districts[a.selected]
	def := content.Districts[d.ID]
	rep := sim.Neglect(d)
	col := neglectColour(rep.Neglect)

	box := vbox(6)
	add(box, hbox(10, districtIcon(d.ID, col), text("THE "+strings.ToUpper(def.Name), 20, colText),
		text(def.FigureTitle+" "+d.Figure, 14, colDim)))
	if d.Turned {
		add(box, para(fmt.Sprintf("TURNED on night %d. It is a hollow now: stewards, decrees and events cannot reach it, it makes nothing, and it adds %s to the Dreamer every dusk until its heart is cleared. Send a party in at dusk.",
			s.Hollows[d.ID].Night, tenths(t.DreamerPerOpenHollow)), 14, colBlood, 550))
		add(box, text("What is in there, as far as the Ledger can say:", 14, colDim))
		for _, p := range sim.Preview(&a.ledger, s.Hollows[d.ID]) {
			add(box, a.previewLine(p))
		}
		return box
	}

	add(box, tip(hbox(8, text(fmt.Sprintf("Neglect %d", rep.Neglect), 16, col), meter(rep.Neglect, 100, col, 260, 14),
		text(fmt.Sprintf("turns at %d or more", t.TurnThreshold), 12, colDim)), neglectTip(d)))
	why := vbox(1, text("Why, largest first:", 13, colDim))
	for _, c := range rep.Causes {
		add(why, hbox(6, sized(rect(max(c.Thirds/3, 1)*2, 8, col), 70, 8), text(causeLine(c), 13, colText)))
	}
	if len(rep.Causes) == 0 {
		add(why, text("Nothing. Every need is full and every building stands.", 13, colGood))
	}
	add(box, why)

	// Needs.
	needs := vbox(2)
	for n := content.Need(0); n < content.NumNeeds; n++ {
		v := d.Needs[n]
		note := fmt.Sprintf("-%d a day", s.NeedDecay(d.ID, n))
		if v < t.EnemyNeedBelow {
			note += ", breeds " + content.FamilyOf(n).String()
		}
		if n == d.Grievance() {
			note += ", the grievance"
		}
		row := hbox(8, sized(text(n.String(), 14, colText), 60, 0), meter(v, 100, needColour(v), 180, 12), sized(text(fmt.Sprint(v), 14, needColour(v)), 28, 0), text(note, 12, colDim))
		add(needs, tip(row, fmt.Sprintf("%s falls by %d every day here. A steward adds %d.\nBelow %d it breeds %s in the hollow: the lower, the more and the stronger. They drop %s.\nThe lowest need is the %s's grievance: it decides what the boss can do.",
			n, s.NeedDecay(d.ID, n), t.StewardNeed, t.EnemyNeedBelow, content.FamilyOf(n), content.FamilyOf(n).Material(), def.FigureTitle)))
	}
	add(box, needs)

	// Buildings.
	builds := vbox(2)
	for i, b := range d.Buildings {
		note, c := "", colText
		switch {
		case b.Condition < t.RuinThreshold:
			note, c = "ruined: a hazard room", colBlood
		case b.Condition >= t.HavenThreshold:
			note, c = "kept: a safe room", colGood
		}
		row := hbox(8, sized(text(def.Buildings[i].Name, 14, c), 110, 0), meter(b.Condition, 100, c, 110, 10), sized(text(fmt.Sprint(b.Condition), 14, c), 28, 0), sized(text(note, 12, colDim), 140, 0))
		tip(row, fmt.Sprintf("The %s is at %d and loses %d a day. A steward adds %d.\nBelow %d it is ruined: +%d Neglect, and it becomes a hazard room in the hollow (%s).\nAt %d or more it becomes a safe room (%s).",
			def.Buildings[i].Name, b.Condition, t.BuildingDecay, t.StewardBuilding, t.RuinThreshold, t.NeglectPerRuin, def.Buildings[i].Hazard, t.HavenThreshold, def.Buildings[i].Haven))
		if b.Warded {
			add(row, text("warded", 12, colRot))
		} else if s.Materials[content.Rot] > 0 || s.CanBuy(content.Ward) {
			add(row, a.command("Ward ("+matsText(content.Upgrades[content.Ward].Cost())+")", content.Upgrades[content.Ward].Text, sim.BuyUpgrade(content.Ward, d.ID, i)))
		}
		add(builds, row)
	}
	add(box, builds)

	// If it turned tonight.
	would := vbox(1, text("If it turned tonight, the hollow would hold:", 13, colDim))
	for _, p := range a.wouldBecome(d) {
		add(would, a.previewLine(p))
	}
	add(box, would)

	// Stewards.
	stew := vbox(3, text(fmt.Sprintf("Stewards (+%d to one need, +%d to one building, at day's end):", t.StewardNeed, t.StewardBuilding), 13, colDim))
	for i, post := range s.Stewards {
		name := content.StewardNames[i]
		row := hbox(6, sized(text(name, 14, colText), 110, 0))
		if post.District != d.ID {
			where := "idle"
			if post.District.Valid() {
				where = "in the " + post.District.String()
			}
			add(row, a.command("Send here", name+" is "+where+".", sim.AssignSteward(i, d.ID, content.NoNeed, -1)), text(where, 12, colDim))
		} else {
			pick := func(label string, need content.Need, building int, on bool) node {
				b := a.command(label, "", sim.AssignSteward(i, d.ID, need, building))
				if on {
					b.AsControl().AddThemeColorOverride("font_color", colGold)
				}
				return b
			}
			add(row, text("need:", 12, colDim), pick("lowest", content.NoNeed, post.Building, post.Need == content.NoNeed))
			for n := content.Need(0); n < content.NumNeeds; n++ {
				add(row, pick(n.String(), n, post.Building, post.Need == n))
			}
			add(row, a.command("Recall", "", sim.RecallSteward(i)))
			brow := hbox(6, gap(110, 0), text("building:", 12, colDim), pick("worst", post.Need, -1, post.Building == -1))
			for bi := range d.Buildings {
				add(brow, pick(def.Buildings[bi].Name, post.Need, bi, post.Building == bi))
			}
			add(stew, row)
			row = brow
		}
		add(stew, row)
	}
	add(box, stew)

	// Residents and output.
	res := vbox(1, text("Residents (they work here by day and can fight by night):", 13, colDim))
	for i := range s.Citizens {
		c := &s.Citizens[i]
		if c.Home != d.ID {
			continue
		}
		line, cc := fmt.Sprintf("%s, %s. %s. Works at %d%%.", c.Name, content.Jobs[c.Job].Name, traitsText(c), c.WorkPct()), colText
		switch {
		case !c.Alive:
			line, cc = c.Name+", "+content.Jobs[c.Job].Name+". Dead.", colDim
		case c.Exhausted:
			cc = colWarn
		}
		add(res, tip(para(line, 13, cc, 560), traitsTip(c)))
	}
	out := s.Output(d.ID)
	add(res, text("Makes today: "+strings.ReplaceAll(costText(out), "free", "nothing")+".", 13, colGold))
	add(box, res)
	return box
}

// previewLine draws one line of a preview: named if the Ledger knows it, and
// always with its cause.
func (a *App) previewLine(p sim.PreviewLine) node {
	col := colText
	if !p.Known {
		col = colDim
	}
	return tip(para("  "+p.String(), 13, col, 540), "Why: "+p.Cause.Describe()+".")
}

func (a *App) eventPanel() node {
	s := &a.game.State
	box := vbox(4)
	if s.Card == nil {
		return panel(box)
	}
	card := content.EventCards[s.Card.Card]
	add(box, text("EVENT: "+strings.ToUpper(card.Title), 15, colGold),
		para(strings.ReplaceAll(card.Text, "{district}", s.Card.District.String()), 14, colText, 580))
	if s.Card.Resolved {
		ch := card.Choices[s.Card.Choice]
		add(box, para("You chose: "+ch.Label+". "+ch.Result, 13, colDim, 580))
		return panel(box)
	}
	row := hbox(8)
	for i, ch := range card.Choices {
		add(row, a.command(ch.Label+" ("+costText(ch.Cost)+")", "", sim.ResolveEvent(i)))
	}
	add(box, row)
	return framed(colGold, box)
}

func (a *App) decreePanel() node {
	s := &a.game.State
	head := "DECREE (one a day)"
	if s.DecreeIssued {
		head = "DECREE: today's is issued"
	}
	box := vbox(4, text(head, 14, colDim))
	row := hbox(6)
	for id := content.DecreeID(0); id < content.NumDecrees; id++ {
		dec := content.Decrees[id]
		label := dec.Name
		if dec.Targeted {
			label += ": " + a.selected.String()
		}
		b := a.command(label+" ("+costText(dec.Cost)+")", dec.Text, sim.IssueDecree(id, a.selected))
		b.AsControl().AddThemeFontSizeOverride("font_size", 13)
		add(row, b)
		if id == content.Corvee {
			add(box, row)
			row = hbox(6)
		}
	}
	add(box, row)
	return panel(box)
}

// upgradeScreen lists what the city can make from what the hollows drop.
func (a *App) upgradeScreen() node {
	s := &a.game.State
	box := vbox(8)
	for id := content.UpgradeID(0); id < content.NumUpgrades; id++ {
		up := content.Upgrades[id]
		row := hbox(12, sized(text(up.Name, 16, colText), 220, 0), sized(text(matsText(up.Cost()), 14, colHunger), 230, 0))
		switch {
		case up.Building:
			add(row, sized(text("on a building", 13, colDim), 150, 0))
		case s.Owned[id]:
			add(row, sized(text("made", 14, colGood), 150, 0))
		default:
			add(row, sized(a.command("Make it", up.Text, sim.BuyUpgrade(id, content.NoDistrict, -1)), 150, 0))
		}
		add(row, para(up.Text, 13, colDim, 520))
		add(box, row)
	}
	return vbox(10,
		a.topBar(),
		text("UPGRADES", 22, colGold),
		para("Made from what the hollows drop: hungerbone from Hunger (low Supply), ash-masks from Riot (low Order), blackcap from Rot (low Upkeep). Most take a single material, so letting one need fall, on purpose, is a way to pay for them. Wards are painted on a building: open its district and use the Ward button beside it. Upgrades last this tenure only.", 14, colDim, 1150),
		panel(box),
		a.footer(a.button("close-upgrades", "Back to the city", "", true, func() { a.overlay = "" })),
	)
}

// affordable counts the city-wide upgrades the city could make now.
func (a *App) affordable() int {
	s := &a.game.State
	n := 0
	for id := content.UpgradeID(0); id < content.NumUpgrades; id++ {
		if !content.Upgrades[id].Building && !s.Owned[id] && s.CanBuy(id) {
			n++
		}
	}
	return n
}

func (a *App) cityScreen() node {
	s := &a.game.State
	cards := grid(3, 8)
	for d := range s.Districts {
		add(cards, a.districtCard(&s.Districts[d]))
	}
	left := sized(vbox(8, cards, a.eventPanel(), a.decreePanel()), 620, 0)
	right := scroll(580, 520, margin(4, a.districtPanel()))
	end := a.command("End the day", "Stewards work, districts produce, needs decay. Then dusk.", sim.EndDay())
	label := "Upgrades"
	if n := a.affordable(); n > 0 {
		label = fmt.Sprintf("Upgrades (%d you can make)", n)
	}
	ups := a.button("upgrades", label, "What the city can make from hungerbone, ash-masks and blackcap.", true, func() { a.overlay = "upgrades" })
	return vbox(8,
		a.topBar(),
		a.hint("three stewards, six districts. Open a district, send a steward, answer the event, then end the day. Hover over anything to see why."),
		hbox(12, left, panel(right)),
		a.footer(ups, end),
	)
}
