# Progress

## Done

**M0: plan and toolchain spike** (2026-11-26)

- Design, plan, open questions and decisions written.
- Go module scaffolded; graphics.gd and the `gd` tool pinned in `go.mod`.
- `make check` passes: `go vet`, golangci-lint (depguard and forbidigo rules proven), `go test -race`.
- Toolchain proven on Linux with Go 1.27.4, Godot 4.7.2 and graphics.gd `1b2edc46bdb6`:
  - a Go-defined `Node2D` draws a rectangle and a label in a Godot window (`make shot`, screenshot checked by eye);
  - `make smoke` runs headless and exits 0; a failing autoplay exits 1;
  - `gd test` also runs Go tests inside the engine.
- graphics.gd could do everything the spike asked. No blocker found.

**M1: city core** (2026-11-26)

- `sim.Apply(state, cmd) (State, []Event, error)` with plain JSON state; a rejected command changes nothing.
- Six seeded PCG streams kept in state (setup, events, dusk, generation, combat, fate).
- New tenure from a seed: six districts, 20 citizens with jobs, two traits and jittered stats, shuffled key figures and starting needs.
- Day rules: steward posts with defaults, five decrees, 12 event cards, production of coin, grain and timber, need and building decay.
- Neglect with its causes, largest first.
- `content` holds the data; every number is in `tuning.go`.
- Tests: Neglect table, steward defaults, each decree, every card and choice, the deck, decay, output, day-tick order, illegal commands, input never mutated, JSON round trip with identical future, stream independence, determinism over 3 seeds for 20 days.

**M2: hollow generator** (2026-11-26)

- `sim.GenerateHollow`: district type to layout, needs to enemy families, buildings to hazard and safe rooms, key figure to boss.
- Every room, enemy and boss carries a `Cause`; `Cause.Resolve` checks it against district state and `Cause.Describe` renders it ("Market Supply was 18").
- Content: nine enemies in three families, six layouts, three boss twists, enemy moves.
- `simcli generate -seed N` prints a hollow as text. `-needs` and `-conds` set a district by hand.
- Tests: 11 legibility fixtures, "lower need means more and stronger", 500 fixed seeds, districts from real play, `FuzzGenerateHollow` (about a million inputs clean), graph invariants.

**M3: expedition and combat, headless** (2026-11-26)

- Dusk: the most neglected district turns at Neglect 40 or more, ties broken by a seeded stream; its hollow is generated then and frozen with the district.
- Expeditions: parties of 1 to 4, room-to-room movement, go deeper or go home; hazard and safe rooms; cleared rooms stay cleared.
- Combat: initiative by Speed, seven job abilities, 17 enemy and boss moves, intents shown two actions ahead, a round cap.
- Night's end: Downed citizens die (25%) or take a Scar; those who fought are exhausted; clearing the heart restores the district and drops loot.
- Policies `random`, `repair-worst` and `farm`; `simcli play -seed N -policy P -nights K` prints the event log as text.
- Tests: dusk selection and ties, party rules, hazards and safe rooms, death and Scar rates over 2,000 seeds, exhaustion, initiative, every ability, every enemy move, the round cap, wipes. A property test plays 36 tenures and checks after every command that HP stays in range, causes still match the frozen district, every enemy action was announced, and every fight ends.

**M4: full tenure** (2026-11-26)

- The Dreamer: rises every dusk by the turned district's Neglect over ten, or 5 on a quiet night, plus 15 per district still turned; one dream line per night, escalating; the lid breaks at 100.
- Endings: won after 20 nights, lost to the lid, lost with no citizen left. A new tenure starts from a new seed.
- Materials buy seven upgrades; six take a single material.
- The Ledger, built from the event log; the dusk preview reads it.
- The chronicle, generated from the event log.
- Save and load: `sim.EncodeSave` and `DecodeSave` over the state and the log; a loaded game plays out the same future.
- `cmd/balance`: 1,000 tenures per policy in parallel (about 7 seconds), with win rate, mean nights survived, how often each district turned, and the three targets.
- Tests: the Dreamer's arithmetic, each ending, each upgrade, Ledger recording and round trip, the preview with an empty and a filled Ledger, the chronicle against a fixed text, save and load mid-fight, determinism over full tenures for 3 seeds and 3 policies, the harness under the race detector.

Balance with the starting values (not yet tuned):

| Policy | Wins | Mean nights | Target |
| --- | --- | --- | --- |
| `random` | 0.0% | 3.3 | under 10% |
| `repair-worst` | 98.8% | 20.0 | 30 to 50% |
| `farm` | 0.0% | 15.7 | at least 80% of `repair-worst` |

**M5: Godot presentation** (2026-11-26)

- Ten screens, all Go through graphics.gd, rebuilt from state after every command: title, the old warden's letter, city, upgrades, dusk, node map, combat, after-action report, Ledger, tenure summary.
- City: six district cards with Neglect and needs; a district panel with Neglect's causes largest first, needs, buildings, "if it turned tonight", steward orders, residents and output; the event card; decrees.
- Dusk: the Dreamer's line and why it rose, what turned and why, the Ledger's preview of every open hollow, the party roster.
- Night: the hollow as a node map with a cause on every room; combat with every enemy's cause and its next two actions.
- Saves and the Ledger in `user://`; "Continue" on the title screen; hints during the first tenure.
- `make smoke`: headless autoplay through the UI's own buttons. It starts a tenure, sends each steward out, ends the day, picks a party, enters the hollow, fights one room, goes home, reads the report, and checks the save on disk matches the game. Exit 0; `--autoplay-fail` exits 1.
- `make shots`: the same script in a real window on a virtual display, saving a PNG of every screen. All ten were checked by eye.
- Tests for the text the screens show (causes, intents, report lines, action ids) run under plain `go test`.

**M6: tuning and finish** (2026-11-26)

- Balance targets met by changing `tuning.go` only (table below; what changed and why is in `DECISIONS.md`, D66).
- Story text: the old warden's letter (7 lines), five fragments found in safe rooms, a closing text unlocked by all five, 22 lines for the Dreamer in six tiers.
- First-tenure hints on every screen; autosave and "Continue".
- Linux desktop build made and smoke-tested. The CI workflow built Linux, Windows and macOS (artifacts on the Actions run); it also passes `go mod verify`, `make check`, `make balance` and `make smoke` on GitHub's runner.
- `make smoke-full`: a whole tenure played through the UI's buttons; passes on nine seeds.

| Policy | Wins | Mean nights | Target | |
| --- | --- | --- | --- | --- |
| `random` | 0.0% | 3.0 | under 10% | met |
| `repair-worst` | 41.0% | 15.9 | 30 to 50% | met |
| `farm` | 57.6% | 17.6 | at least 32.8% (80% of `repair-worst`) | met |

1,000 tenures each, seeds 1 to 1,000. Three other seed ranges gave 37.6 to 41.9% for `repair-worst`.

## Definition of done

| Requirement | State |
| --- | --- |
| A new player can finish a 20-night tenure in a desktop build in 60 to 120 minutes | Estimated at 65 to 89 minutes from counted decisions. Not measured: nobody has played it by hand. |
| Every enemy, hazard and boss shows why it exists | Yes: in the fight, on the map, in the dusk preview, in the report and in the Ledger. Tests hold every generated thing to a cause that matches city state. |
| The balance targets are met | Yes. |
| `make check`, `make balance` and `make smoke` all pass | Yes. |
| `PROGRESS.md` lists the next steps toward the full game | Below. |

## Next steps toward the full game

1. **Play it.** One person, one tenure, start to finish, on a real display. Watch for: text that overflows at other window sizes, whether the cause lines are actually read, and how long it takes.
2. **Start the Windows and macOS builds.** CI produces them, but nobody has run either. Download them from the latest "build" run under Actions.
3. **Mouse and keyboard polish.** Keyboard shortcuts for the common actions; a confirm on "Send nobody tonight" when a hollow is open.
4. **Make `repair-worst` lose for more than one reason.** Nearly all losses are the lid. A tenure that ends because the city ran out of people should be possible more often.
5. **Art and audio**, which the slice left out: district art, enemy portraits, a sound for the Dreamer's line.
6. **More city.** Procedural layouts, more than six districts, warden classes and traits. All parked in the spec as out of scope.
7. **More to learn.** The Ledger can record 42 things (9 enemies, 15 hazards, 18 boss variants). A longer game needs more enemy kinds per family and more than three twists per boss.

## Known issues

- The Windows and macOS builds exist as CI artifacts but have never been started. Only the Linux build has been run.
- The first `gd` run on a machine is slow: it builds a patched copy of the Go compiler and the engine bindings. About two minutes here on two cores. Later runs take seconds.
- `gd test` overwrites `graphics/linux_amd64.so` with a test build. The next `gd run` rebuilds it.
- The smoke tests press buttons by action id; they do not move a mouse (`DECISIONS.md`, D57). Nobody has clicked through the game by hand yet.
- The UI is laid out for 1280 by 720 and has only been seen in screenshots from a software renderer. Later screens in a tenure (several open hollows, a fight with four enemies) were exercised by `make smoke-full` but not looked at.
- `random` loses in about three nights, every time. It meets its target but is a weak floor.
