package game

import (
	"fmt"
	"strings"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

func (a *App) titleScreen() node {
	box := vbox(12,
		gap(0, 90),
		text("HOLLOW WARD", 54, colText),
		para("You are the warden of a walled city built as a lid over something that dreams.", 18, colDim, 700),
		gap(0, 20),
		sized(a.button("new-tenure", "Take the keys: a new tenure", "Twenty nights. A reshuffled city.", true, a.newTenure), 360, 44),
		sized(a.button("continue", "Continue the saved tenure", "", a.hasSave(), a.loadGame), 360, 44),
		sized(a.button("ledger", "Open the Ledger", "What earlier wardens learned.", true, func() { a.overlay = "ledger" }), 360, 44),
		gap(0, 10),
		text(fmt.Sprintf("Wardens before you: %d. Of those, %d kept the lid on.", a.ledger.Tenures, a.ledger.Won), 14, colDim),
	)
	if a.message != "" {
		add(box, text(a.message, 14, colWarn))
	}
	return margin(60, box)
}

func (a *App) letterScreen() node {
	box := vbox(10, text("THE LAST WARDEN'S LETTER", 22, colGold), gap(0, 6))
	for _, line := range content.OpeningLetter {
		add(box, para(line, 17, colText, 760))
	}
	add(box, gap(0, 12), sized(a.button("read-letter", "Take the keys", "", true, func() { a.overlay = "" }), 240, 44))
	return margin(70, box)
}

func (a *App) reportScreen() node {
	s := &a.game.State
	night := s.Day - 1
	if s.Phase == sim.PhaseOver {
		night = s.Nights
	}
	box := vbox(2)
	for _, line := range reportLines(a.report) {
		col := colText
		switch {
		case strings.Contains(line, "does not wake"), strings.Contains(line, "DOWNED"), strings.Contains(line, "TURNS"):
			col = colBlood
		case strings.HasPrefix(line, "Loot"), strings.Contains(line, "scavenges"), strings.Contains(line, "district again"):
			col = colGood
		case strings.Contains(line, "Dreamer"), strings.HasPrefix(line, "\""):
			col = colDream
		}
		add(box, para(line, 14, col, 1150))
	}
	return vbox(10,
		a.topBar(),
		text(fmt.Sprintf("NIGHT %d: WHAT HAPPENED, AND WHY", night), 20, colGold),
		a.hint("everything the night made is listed with the city state that made it."),
		panel(scroll(1200, 500, box)),
		a.footer(a.button("close-report", "Morning", "", true, func() { a.overlay = "" })),
	)
}

func (a *App) ledgerScreen() node {
	l := &a.ledger
	box := vbox(3)
	section := func(title string, kind sim.LedgerKind, name func(e sim.LedgerEntry) string) {
		add(box, gap(0, 6), text(title, 16, colGold))
		n := 0
		for _, e := range l.Entries {
			if e.Kind != kind {
				continue
			}
			n++
			add(box, tip(para(fmt.Sprintf("  %s First met in the %s, tenure %d, because %s. Met %s.",
				name(e), e.District, e.Tenure, e.Cause.Describe(), times(e.Seen)), 14, colText, 1150), "Why: "+e.Cause.Describe()+"."))
		}
		if n == 0 {
			add(box, text("  Nothing recorded yet.", 14, colDim))
		}
	}
	section("ENEMIES (what low needs breed)", sim.LedgerEnemy, func(e sim.LedgerEntry) string {
		for _, def := range content.Enemies {
			if def.Slug == e.ID {
				return fmt.Sprintf("%s (%s, from low %s): %s", def.Name, def.Family, content.NeedOf(def.Family), def.Text)
			}
		}
		return e.ID + "."
	})
	section("HAZARDS (what ruined buildings become)", sim.LedgerHazard, func(e sim.LedgerEntry) string {
		for _, b := range content.Districts[max(e.District, 0)].Buildings {
			if e.District.Slug()+"/"+b.Slug == e.ID {
				return b.Hazard + "."
			}
		}
		return e.ID
	})
	section("AT THE HEART (what key figures become)", sim.LedgerBoss, func(e sim.LedgerEntry) string {
		for n := content.Need(0); n < content.NumNeeds; n++ {
			if content.BossSlug(e.District, n) == e.ID {
				return content.BossName(e.District, n) + ". " + content.Twists[n].Text
			}
		}
		return e.ID
	})
	add(box, gap(0, 6), text(fmt.Sprintf("THE OLD WARDEN'S NOTES (%d of %d found, in safe rooms)", len(l.Fragments), len(content.Fragments)), 16, colGold))
	for _, f := range l.Fragments {
		add(box, para("  “"+content.Fragments[f%len(content.Fragments)]+"”", 14, colText, 1100))
	}
	if l.HasAllFragments() {
		add(box, gap(0, 6), text("THE LAST PAGE", 16, colDream))
		for _, line := range content.ClosingText {
			add(box, para("  "+line, 14, colText, 1100))
		}
	}
	return vbox(10,
		text("THE LEDGER", 26, colText),
		para(fmt.Sprintf("Passed from warden to warden. It records what each neglect breeds, and it is the only thing that passes on. It grants nothing but knowledge. Wardens so far: %d; %d kept the lid on.", l.Tenures, l.Won), 14, colDim, 1150),
		panel(scroll(1200, 520, box)),
		hbox(8, spacer(), a.button("close-ledger", "Close the Ledger", "", true, func() { a.overlay = "" })),
	)
}

func (a *App) summaryScreen() node {
	s := &a.game.State
	head, col := "THE LID BROKE", colBlood
	switch s.Outcome {
	case sim.Won:
		head, col = "THE LID HELD", colGood
	case sim.LostNoParty:
		head, col = "NO ONE IS LEFT TO SEND", colBlood
	}
	box := vbox(4)
	for _, line := range sim.Chronicle(a.game.Log) {
		add(box, para(line, 14, colText, 1130))
	}
	return vbox(10,
		text(head, 34, col),
		text(fmt.Sprintf("Nights survived: %d of %d.   The Dreamer: %s of %d.   Citizens alive: %d of %d.",
			s.Nights, content.T.Nights, tenths(s.Dreamer), content.T.DreamerMax/10, s.Living(), len(s.Citizens)), 16, colText),
		text("THE CHRONICLE OF THIS TENURE", 16, colGold),
		panel(scroll(1200, 440, box)),
		para("The next warden inherits a reshuffled city and the Ledger. Nothing else.", 14, colDim, 1100),
		hbox(8, spacer(),
			a.button("ledger", "Open the Ledger", "", true, func() { a.overlay = "ledger" }),
			a.button("new-tenure", "Hand on the keys: a new tenure", "", true, a.newTenure)),
	)
}
