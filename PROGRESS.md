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

## Next

**M4: full tenure.** See `PLAN.md`.

## Known issues

- macOS and Windows builds cannot be produced or tested in the current build sandbox (Q3). Only Linux is proven.
- `go.sum` was produced without access to the Go checksum database. Run `go mod verify` once on a networked machine (`DECISIONS.md`, D10).
- The first `gd` run on a machine is slow: it builds a patched copy of the Go compiler and the engine bindings. About two minutes here on two cores. Later runs take seconds.
- `gd test` overwrites `graphics/linux_amd64.so` with a test build. The next `gd run` rebuilds it.
- `make balance` is a stub that exits 2 until M4.
- The tenure has no end yet: the Dreamer, winning and losing arrive in M4.
- With the starting values `repair-worst` clears every hollow easily. Balance is M6.
