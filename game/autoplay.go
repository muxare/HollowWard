package game

import (
	"fmt"
	"os"
	"path/filepath"

	"hollowward/internal/content"
	"hollowward/internal/policy"
	"hollowward/internal/sim"
)

// autoplay is the smoke test. A script plays through the real UI: it presses
// the same action ids the buttons on screen carry, and fails if the screen
// does not offer the action it needs. It starts a tenure, assigns stewards,
// ends days until a district turns, sends a party in, fights one room, goes
// home, and reads the report.
type autoplay struct {
	app    *App
	stage  int
	wait   int // frames to let the UI settle between presses
	settle int // frames waited before a screenshot
	steps  int
	shot   map[string]bool
	fought bool
	seen   map[string]bool // screens visited
	done   bool
}

func newAutoplay(a *App) *autoplay {
	if a.opts.seed == 0 {
		a.opts.seed = 20261126 // a fixed city, so the smoke test is repeatable
	}
	return &autoplay{app: a, shot: map[string]bool{}, seen: map[string]bool{}}
}

func (p *autoplay) fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "autoplay: FAIL on screen %q: %s\n", p.app.screen, fmt.Sprintf(format, args...))
	p.done = true
	p.app.quit(1)
}

// press presses a button by its action id, and fails if no such button is on
// screen and enabled.
func (p *autoplay) press(id string) bool {
	if !p.app.press(id) {
		p.fail("the screen offers no enabled button %q", id)
		return false
	}
	p.wait = 2
	return true
}

// send presses the button that sends a command.
func (p *autoplay) send(c sim.Command) bool { return p.press(actionID(c)) }

func (p *autoplay) step() {
	a := p.app
	if p.done || a.dirty {
		return // let the redraw happen first
	}
	if p.wait > 0 {
		p.wait--
		return
	}
	if a.opts.fail {
		p.fail("asked to fail")
		return
	}
	if p.steps++; p.steps > 3000 {
		p.fail("no end after %d steps", p.steps)
		return
	}
	p.seen[a.screen] = true
	if a.opts.shots != "" && !p.shot[a.screen] {
		if a.frames < 8 || p.settle < 3 {
			p.settle++ // give the window a few frames to draw the new screen
			return
		}
		p.settle = 0
		p.shot[a.screen] = true
		if err := a.snap(filepath.Join(a.opts.shots, a.screen+".png")); err != nil {
			p.fail("screenshot: %v", err)
			return
		}
	}
	if a.message != "" && a.screen != "title" {
		p.fail("the UI reported: %s", a.message)
		return
	}

	if a.opts.full && a.screen != "title" && a.screen != "letter" {
		p.whole()
		return
	}
	switch a.screen {
	case "title":
		if p.stage > 0 {
			p.fail("back on the title screen")
			return
		}
		p.stage = 1
		p.press("new-tenure")
	case "letter":
		p.press("read-letter")
	case "city":
		p.city()
	case "dusk":
		p.dusk()
	case "map":
		p.hollow()
	case "combat":
		p.fought = true
		s := &a.game.State
		c := policy.RepairWorst{}.Next(s) // a sound combat move
		p.send(c)
	case "report":
		if p.stage == 4 {
			if !p.finish() {
				return
			}
			if a.opts.shots == "" {
				p.done = true
				a.quit(0)
				return
			}
			p.stage = 5 // taking pictures: go on to the Ledger and the end of the tenure
		}
		p.press("close-report")
	case "upgrades":
		p.press("close-upgrades")
	case "ledger":
		p.stage = 6
		p.press("close-ledger")
	case "summary":
		if p.stage < 5 {
			p.fail("the tenure ended before the script finished (outcome %s)", a.game.State.Outcome)
			return
		}
		p.done = true
		a.quit(0)
	default:
		p.fail("unknown screen")
	}
}

// city plays a day through the buttons: on day 1 every steward is sent out;
// after that the city is left alone, so that a district turns soon.
func (p *autoplay) city() {
	a := p.app
	s := &a.game.State
	switch p.stage {
	case 5: // pictures: the upgrades, then the Ledger
		if !p.seen["upgrades"] {
			p.press("upgrades")
		} else {
			p.press("ledger")
		}
		return
	case 6: // pictures: play out the tenure with a scripted warden, then the summary
		for s.Phase != sim.PhaseOver {
			a.send(policy.RepairWorst{}.Next(s))
			s = &a.game.State
		}
		a.overlay, a.dirty, p.wait = "", true, 3
		return
	}
	if s.Day == 1 {
		for i, post := range s.Stewards {
			want := content.DistrictID(i)
			if post.District != want {
				if a.selected != want {
					p.press("select:" + want.Slug())
					return
				}
				p.send(sim.AssignSteward(i, want, content.NoNeed, -1))
				return
			}
		}
		p.stage = 2
	} else {
		for i, post := range s.Stewards {
			if post.District.Valid() {
				if a.selected != post.District {
					p.press("select:" + post.District.Slug())
					return
				}
				p.send(sim.RecallSteward(i))
				return
			}
		}
	}
	if s.Card != nil && !s.Card.Resolved {
		for i, ch := range content.EventCards[s.Card.Card].Choices {
			if ch.Cost.IsZero() {
				p.send(sim.ResolveEvent(i))
				return
			}
		}
	}
	if s.Day > 12 {
		p.fail("twelve days and no district has turned")
		return
	}
	p.send(sim.EndDay())
}

func (p *autoplay) dusk() {
	a := p.app
	s := &a.game.State
	if !a.target.Valid() || s.Hollows[a.target] == nil {
		p.send(sim.StayHome()) // a quiet night
		return
	}
	// Pick the four hardiest through the roster buttons, then send them.
	want := 0
	for i := range s.Citizens {
		if s.Citizens[i].Alive && want < content.T.MaxParty {
			want++
			picked := false
			for _, id := range a.party {
				picked = picked || id == i
			}
			if !picked {
				p.press(fmt.Sprintf("pick:%d", i))
				return
			}
		}
	}
	p.stage = 3
	p.send(sim.SendParty(a.target, a.party...))
}

// hollow walks deeper until one room has been fought, then goes home.
func (p *autoplay) hollow() {
	a := p.app
	s := &a.game.State
	if p.fought {
		p.stage = 4
		p.send(sim.GoHome())
		return
	}
	r := &s.Hollows[s.Expedition.District].Rooms[s.Expedition.Room]
	if len(r.Exits) == 0 {
		p.fail("reached the end of the hollow without a fight")
		return
	}
	p.send(sim.Advance(r.Exits[0]))
}

// finish checks the session did what a smoke test promises. It reports
// whether it did.
func (p *autoplay) finish() bool {
	a := p.app
	s := &a.game.State
	for _, want := range []string{"title", "letter", "city", "dusk", "map", "combat", "report"} {
		if !p.seen[want] {
			p.fail("never saw the %s screen", want)
			return false
		}
	}
	turned := 0
	for _, h := range s.Hollows {
		if h != nil {
			turned++
		}
	}
	switch {
	case !p.fought:
		p.fail("no room was fought")
	case turned == 0:
		p.fail("went home, but no district is still turned")
	case len(a.report) == 0 || len(reportLines(a.report)) == 0:
		p.fail("the after-action report is empty")
	case len(a.ledger.Entries) == 0:
		p.fail("a fight was fought and the Ledger recorded nothing")
	case !a.hasSave():
		p.fail("no save was written to user://")
	default:
		loaded, err := sim.DecodeSave(mustRead(a.savePath()))
		if err != nil {
			p.fail("the save does not load: %v", err)
			return false
		}
		want, _ := sim.EncodeSave(a.game)
		if got, _ := sim.EncodeSave(loaded); string(got) != string(want) {
			p.fail("the save on disk is not the game in play")
			return false
		}
		fmt.Printf("autoplay: ok. %d steps, day %d, %d district turned, %d Ledger entries, Dreamer %s\n",
			p.steps, s.Day, turned, len(a.ledger.Entries), tenths(s.Dreamer))
		return true
	}
	return false
}

// whole plays an entire tenure through the buttons, with the careful warden
// choosing. Every command the warden wants must be on offer on some screen,
// and this walks the UI to it: it opens the right district, the upgrades
// screen, the right hollow, and ticks the right citizens.
func (p *autoplay) whole() {
	a := p.app
	s := &a.game.State
	switch a.screen {
	case "report":
		p.press("close-report")
		return
	case "ledger":
		p.press("close-ledger")
		return
	case "summary":
		if len(sim.Chronicle(a.game.Log)) < 3 {
			p.fail("the tenure ended with no chronicle")
			return
		}
		if a.ledger.Tenures != 1 {
			p.fail("the Ledger counts %d finished tenures, want 1", a.ledger.Tenures)
			return
		}
		fmt.Printf("autoplay: ok. A whole tenure through the UI: %d steps, %s after %d nights, screens %d\n", p.steps, s.Outcome, s.Nights, len(p.seen))
		p.done = true
		a.quit(0)
		return
	}
	c := policy.RepairWorst{}.Next(s)
	open := func(d content.DistrictID) bool { // make sure the district's panel is showing
		if a.screen == "upgrades" {
			p.press("close-upgrades")
			return false
		}
		if a.selected != d {
			p.press("select:" + d.Slug())
			return false
		}
		return true
	}
	switch c.Kind {
	case sim.CmdAssignSteward:
		if !open(c.District) {
			return
		}
	case sim.CmdRecallSteward:
		if !open(s.Stewards[c.Steward].District) {
			return
		}
	case sim.CmdIssueDecree:
		if content.Decrees[c.Decree].Targeted && !open(c.District) {
			return
		}
	case sim.CmdBuyUpgrade:
		if content.Upgrades[c.Upgrade].Building {
			if !open(c.District) {
				return
			}
		} else if a.screen != "upgrades" {
			p.press("upgrades")
			return
		}
	case sim.CmdSendParty:
		if a.target != c.District {
			p.press("hollow:" + c.District.Slug())
			return
		}
		want := map[int]bool{}
		for _, id := range c.Citizens {
			want[id] = true
		}
		for _, id := range a.party { // untick anyone the warden does not want
			if !want[id] {
				p.press(fmt.Sprintf("pick:%d", id))
				return
			}
		}
		if len(a.party) < len(c.Citizens) {
			have := map[int]bool{}
			for _, id := range a.party {
				have[id] = true
			}
			for _, id := range c.Citizens {
				if !have[id] {
					p.press(fmt.Sprintf("pick:%d", id))
					return
				}
			}
		}
		c = sim.SendParty(c.District, a.party...) // the same people, in the order they were ticked
	}
	if a.screen == "upgrades" && c.Kind != sim.CmdBuyUpgrade {
		p.press("close-upgrades")
		return
	}
	p.send(c)
	p.wait = 0
}
