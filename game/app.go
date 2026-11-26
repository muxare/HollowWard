package game

import (
	"fmt"
	"os"
	"time"

	"graphics.gd/classdb/Button"
	"graphics.gd/classdb/ColorRect"
	"graphics.gd/classdb/Control"
	"graphics.gd/classdb/FileAccess"
	"graphics.gd/classdb/RenderingServer"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/classdb/VBoxContainer"
	"graphics.gd/classdb/Viewport"
	"graphics.gd/variant/Float"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

// App is the root of the game's UI. It holds the tenure in play and the
// Ledger, and rebuilds the screen from them after every command.
type App struct {
	Control.Extension[App] `gd:"HollowWardApp"`

	opts   options
	game   *sim.Game  // nil on the title screen
	ledger sim.Ledger // crosses tenures

	// What the player is looking at. None of this is game state.
	overlay  string             // "", "letter", "ledger", "report", "upgrades"
	selected content.DistrictID // district open in the city view
	target   content.DistrictID // hollow chosen at dusk
	party    []int              // citizens picked at dusk
	report   []sim.Event        // last night, for the after-action report
	recent   []sim.Event        // what the last command caused
	message  string             // why the last command was refused

	body    VBoxContainer.Instance
	actions map[string]func() // every enabled button on screen, by action id
	screen  string            // the screen now showing, for autoplay and tests
	dirty   bool
	frames  int
	auto    *autoplay
}

// Ready builds the frame the screens are drawn into.
func (a *App) Ready() {
	RenderingServer.SetDefaultClearColor(colBack)
	a.AsControl().SetAnchorsPreset(Control.PresetFullRect)
	a.AsControl().SetTheme(newTheme())
	back := ColorRect.New()
	back.SetColor(colBack)
	back.AsControl().SetAnchorsPreset(Control.PresetFullRect)
	back.AsControl().SetMouseFilter(Control.MouseFilterIgnore)
	a.body = vbox(8)
	frame := margin(14, a.body)
	frame.AsControl().SetAnchorsPreset(Control.PresetFullRect)
	add(a, back, frame)

	a.selected, a.target = content.Market, content.NoDistrict
	a.loadLedger()
	if a.opts.autoplay {
		a.auto = newAutoplay(a)
	}
	a.render()
}

// Process redraws after a change and drives autoplay.
func (a *App) Process(delta Float.X) {
	a.frames++
	if a.dirty {
		a.dirty = false
		a.render()
	}
	if a.opts.screenshot != "" && a.frames == 12 {
		code := 0
		if err := a.snap(a.opts.screenshot); err != nil {
			fmt.Fprintln(os.Stderr, "screenshot:", err)
			code = 1
		}
		if !a.opts.autoplay {
			a.quit(code)
		}
	}
	if a.auto != nil {
		a.auto.step()
	}
}

func (a *App) quit(code int) { SceneTree.Get(a.AsNode()).MoreArgs().Quit(code) }

// snap saves what the window shows to a PNG.
func (a *App) snap(path string) error {
	return Viewport.Get(a.AsNode()).GetTexture().AsTexture2D().GetImage().SavePng(path)
}

// --- Commands ---

// send is the one way the UI changes the game: it hands a command to the sim
// and redraws from whatever state comes back.
func (a *App) send(c sim.Command) {
	a.message = ""
	if a.game == nil {
		a.game = &sim.Game{}
	}
	events, err := a.game.Do(c)
	if err != nil {
		a.message = capital(err.Error()) + "."
		return
	}
	a.recent = events
	a.ledger = a.ledger.Record(events)
	for _, e := range events {
		switch e.Kind {
		case sim.EvNightEnded:
			a.report = a.game.Tonight()
			a.overlay = "report"
			a.party, a.target = nil, content.NoDistrict
			a.saveLedger()
		case sim.EvTenureEnded:
			a.saveLedger()
		case sim.EvTurned:
			a.target = e.District
		}
	}
	a.saveGame()
}

// newTenure starts a tenure from a fresh seed. The sim cannot read a clock,
// so the seed comes from here.
func (a *App) newTenure() {
	seed := a.opts.seed
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	a.game = &sim.Game{}
	a.send(sim.NewTenure(seed))
	a.overlay, a.selected, a.party, a.target = "letter", content.Market, nil, content.NoDistrict
}

// offers reports whether the rules would accept a command now, and why not.
func (a *App) offers(c sim.Command) (bool, string) {
	if a.game == nil {
		return false, "No tenure."
	}
	if _, _, err := sim.Apply(a.game.State, c); err != nil {
		return false, capital(err.Error()) + "."
	}
	return true, ""
}

// --- Buttons ---

// button makes a button with an action id. An enabled button is recorded, so
// that autoplay can press exactly what the player could.
func (a *App) button(id, label, hint string, enabled bool, fn func()) Button.Instance {
	b := rawButton(label, enabled)
	if hint != "" {
		tip(b, hint)
	}
	if enabled {
		a.actions[id] = fn
		b.AsBaseButton().OnPressed(func() { a.press(id) })
	}
	return b
}

// command makes a button that sends a command. It is enabled only if the
// rules accept the command now; otherwise its hover text says why not.
func (a *App) command(label, hint string, c sim.Command) Button.Instance {
	ok, why := a.offers(c)
	if !ok {
		hint = why
	}
	return a.button(actionID(c), label, hint, ok, func() { a.send(c) })
}

// press runs the action behind a button id, as a click on it would.
func (a *App) press(id string) bool {
	fn, ok := a.actions[id]
	if !ok {
		return false
	}
	fn()
	a.dirty = true
	return true
}

// --- Drawing ---

// render rebuilds the screen from the state.
func (a *App) render() {
	for _, k := range a.body.AsNode().GetChildren() {
		a.body.AsNode().RemoveChild(k)
		k.QueueFree()
	}
	a.actions = map[string]func(){}
	switch {
	case a.overlay == "ledger":
		a.screen = "ledger"
		add(a.body, a.ledgerScreen())
	case a.game == nil:
		a.screen = "title"
		add(a.body, a.titleScreen())
	case a.overlay == "letter":
		a.screen = "letter"
		add(a.body, a.letterScreen())
	case a.overlay == "report":
		a.screen = "report"
		add(a.body, a.reportScreen())
	case a.overlay == "upgrades" && a.game.State.Phase == sim.PhaseDay:
		a.screen = "upgrades"
		add(a.body, a.upgradeScreen())
	default:
		s := &a.game.State
		switch s.Phase {
		case sim.PhaseDay:
			a.screen = "city"
			add(a.body, a.cityScreen())
		case sim.PhaseDusk:
			a.screen = "dusk"
			add(a.body, a.duskScreen())
		case sim.PhaseNight:
			if s.Expedition != nil && s.Expedition.Fight != nil {
				a.screen = "combat"
				add(a.body, a.combatScreen())
			} else {
				a.screen = "map"
				add(a.body, a.mapScreen())
			}
		default:
			a.screen = "summary"
			add(a.body, a.summaryScreen())
		}
	}
}

// firstTenure reports whether no warden has finished a tenure yet: hints show.
func (a *App) firstTenure() bool { return a.ledger.Tenures == 0 }

// hint is a line of guidance shown during the first tenure only.
func (a *App) hint(s string) node {
	if !a.firstTenure() {
		return gap(0, 0)
	}
	return text("Hint: "+s, 13, colGold)
}

// --- Files in user:// ---

func (a *App) savePath() string {
	if a.opts.autoplay {
		return "user://autoplay-save.json" // never touch the player's own save
	}
	return "user://save.json"
}

func (a *App) ledgerPath() string {
	if a.opts.autoplay {
		return "user://autoplay-ledger.json"
	}
	return "user://ledger.json"
}

func writeFile(path string, data []byte) error {
	f := FileAccess.Open(path, FileAccess.Write)
	if f == FileAccess.Nil {
		return fmt.Errorf("cannot write %s: %w", path, FileAccess.GetOpenError())
	}
	f.StoreString(string(data))
	f.Flush()
	return nil
}

func (a *App) saveGame() {
	if a.game == nil {
		return
	}
	data, err := sim.EncodeSave(a.game)
	if err == nil {
		err = writeFile(a.savePath(), data)
	}
	if err != nil {
		a.message = "Could not save: " + err.Error()
	}
}

func mustRead(path string) []byte { return []byte(FileAccess.GetFileAsString(path)) }

func (a *App) hasSave() bool { return FileAccess.FileExists(a.savePath()) }

func (a *App) loadGame() {
	g, err := sim.DecodeSave([]byte(FileAccess.GetFileAsString(a.savePath())))
	if err != nil {
		a.message = "Could not load the saved tenure: " + err.Error() + "."
		return
	}
	a.game, a.overlay, a.message = g, "", ""
	a.selected, a.party, a.target = content.Market, nil, content.NoDistrict
}

func (a *App) saveLedger() {
	data, err := sim.EncodeLedger(&a.ledger)
	if err == nil {
		err = writeFile(a.ledgerPath(), data)
	}
	if err != nil {
		a.message = "Could not save the Ledger: " + err.Error()
	}
}

func (a *App) loadLedger() {
	if a.opts.autoplay || !FileAccess.FileExists(a.ledgerPath()) {
		return // autoplay always starts from an empty Ledger
	}
	l, err := sim.DecodeLedger([]byte(FileAccess.GetFileAsString(a.ledgerPath())))
	if err != nil {
		a.message = "The Ledger could not be read and was left alone: " + err.Error() + "."
		return
	}
	a.ledger = *l
}
