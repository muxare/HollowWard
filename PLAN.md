# Plan

Tasks per milestone. Each milestone ends with `make check` passing, `PROGRESS.md` updated and a commit. M1 to M4 are pure Go.

## M0: plan and toolchain spike (done, waiting for go-ahead)

- [x] `DESIGN.md`, `PLAN.md`, `QUESTIONS.md`, `DECISIONS.md`, `IDEAS.md`, `PROGRESS.md`
- [x] Go module with graphics.gd and the `gd` tool pinned in `go.mod`
- [x] `Makefile` and `.golangci.yml` (depguard and forbidigo rules proven against deliberate violations)
- [x] Package layout with stubs
- [x] Spike: a Go-defined node draws a rectangle and a label in a Godot window
- [x] Headless run exits 0, and a failing run exits 1

## M1: city core

- `sim.State`, `Command`, `Event` and `Apply`; JSON round-trip of the empty and the starting state
- RNG streams: PCG per subsystem, stored in state, serialised with the save
- `content`: six districts with decay modifiers, buildings and key figures; citizen jobs, names and traits; 3 stewards; `tuning.go` with every starting value from the prompt
- New tenure from a seed: shuffled citizens, key figures and starting needs
- Day commands: assign steward (need and building, with defaults), issue decree, answer event, end day
- Day tick: need decay, building decay, steward effects, production of coin, grain and timber from working citizens and need levels
- Decrees: a short list, one per day, with costs
- Events: 12 seeded cards, each with a choice and effects
- Neglect with its causes, largest first
- Tests: Neglect table, steward defaults, decree costs, production, event determinism, an illegal-command table, stream independence (drawing from one stream leaves the others unchanged)

## M2: hollow generator

- `Cause{Source, ID, ValueAtGeneration}` and the hollow types: room graph, rooms, enemy groups, hazards, safe rooms, heart
- Six layout templates, one per district, each yielding 6 to 10 rooms
- Needs to enemy families, with count and strength scaled by how far below 50 the need is
- Buildings to hazard rooms and safe rooms
- Key figure to boss; grievance to ability set
- `simcli generate -seed N` prints a hollow as text, causes included
- Tests: legibility fixtures (Supply 10, others 80 gives Hunger only, and so on), 500 fixed seeds for cause coverage, `FuzzGenerateHollow`, graph invariants (connected, one heart, heart reachable, 6 to 10 rooms)

## M3: expedition and combat, headless

- Dusk: pick the turned district (Neglect 40 or more, seeded tie-break), or a quiet night
- Party selection, room-to-room movement, go deeper or go home
- Combat: initiative by Speed, enemy intents shown a turn ahead, job abilities, traits, hazards, safe-room rest, round cap
- Night end: Downed resolution (25% death, Scars), exhaustion, clearing restores needs to 50, loot
- Policies: `random`, `repair-worst`, `farm`, written against the command interface
- `simcli play -seed N -policy P` plays one night or a whole tenure and prints the event log
- Tests: HP never negative, turn order follows Speed, every fight ends (property test over seeds), intent shown equals action taken, Downed and Scar rules

## M4: full tenure

- The Dreamer: nightly rise, the +15 per district still turned, the dream line per night, loss at 100
- Win after night 20; loss when no citizen can fight
- Materials and upgrades (wards, gear), including single-material upgrades
- The Ledger: built from the event log, persisted separately from the tenure save, feeds the dusk preview
- The chronicle: generated from the event log
- Save and load in `sim` (encode and decode only)
- `cmd/balance`: parallel tenures, one state per goroutine; win rate, mean nights survived, turn counts per district
- Tests: determinism over 3 seeds for a full tenure (identical state and event log), save and load round-trip, resume-then-play equals play-through, Ledger never changes a stat, balance harness is race-clean

## M5: Godot presentation

- Scene shell and a Godot Theme; Polygon2D shapes and text only
- City view; district panel with the Neglect meter and its causes
- Steward assignment, decree and event UI
- Dusk screen with the Dreamer line and the Ledger preview
- Hollow node map; combat view with enemy intents; cause tooltips on every generated thing
- After-action report, Ledger screen, tenure summary with the chronicle
- `game/` writes saves and the Ledger to `user://`
- `make smoke`: headless autoplay through the real UI command path (start tenure, assign stewards, end day, enter hollow, fight one room, go home), non-zero exit on any failure
- Visual check of each screen from window screenshots (`make shot`)

## M6: tuning and finish

- Meet the balance targets by changing `tuning.go` only: `random` under 10%, `repair-worst` 30 to 50%, `farm` at least 80% of `repair-worst`
- Story text: the letter, five fragments, the closing text, the Dreamer's lines
- First-tenure hints
- Save and load wired into the UI
- Desktop builds for Linux, macOS and Windows
- `PROGRESS.md` lists the next steps toward the full game

## Risks I am watching

1. **Balance by numbers alone.** M6 may only touch `tuning.go`, so the rules have to be right by M4. I will run the harness at the end of M4 and report early if the targets look out of reach (see `QUESTIONS.md`, Q1).
2. **graphics.gd churn.** Pinned by commit. All engine code sits behind `game/`.
3. **Legibility.** The cause tests cover the data. Whether a player actually reads it needs a human playtest in M5.
