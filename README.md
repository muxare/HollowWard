# Hollow Ward

A dark fantasy management game with roguelike expeditions. By day you run a walled city with too few stewards. At dusk the most neglected district becomes a dungeon built from what it lacked.

Written in Go on the Godot engine, joined by [graphics.gd](https://github.com/quaadgras/graphics.gd). This repository is the playable vertical slice. Status: see `PROGRESS.md`.

## You need

- Go 1.27
- golangci-lint 2.15 or newer
- A C compiler (cgo)

`gd`, the graphics.gd build tool, is pinned in `go.mod` and runs as `go tool gd`. It downloads Godot 4.7 on first use.

## Commands

| Command | What it does |
| --- | --- |
| `make check` | `go vet`, lint and race tests. Must pass before every commit. |
| `make run` | Play in a window. |
| `make editor` | Open the Godot editor on `graphics/`. |
| `make smoke` | Start headless and autoplay through the real UI path. |
| `make shots` | Play the smoke script in a window on a virtual display and save a PNG of every screen (Linux, needs xvfb). |
| `make balance` | 1,000 tenures per scripted policy, with the balance targets. |
| `go run ./cmd/simcli play -policy farm -nights 0` | Play a whole tenure in the terminal. |
| `go run ./cmd/simcli generate -seed 7` | Print a hollow as text. |
| `make build` | Desktop builds into `releases/`. |

## Layout

```
main.go              entry point; hands over to game/
cmd/simcli/          headless play and generation from a seed and a policy
cmd/balance/         balance harness
internal/sim/        pure, deterministic game rules
internal/content/    districts, enemies, abilities, events and text as Go data
internal/content/tuning.go   every number
internal/policy/     scripted players
game/                the Godot presentation; the only package that imports graphics.gd
graphics/            Godot project and assets, managed by gd
```

## Documents

`DESIGN.md` the design. `PLAN.md` tasks per milestone. `QUESTIONS.md` open questions. `DECISIONS.md` choices made where the spec was silent. `IDEAS.md` out-of-scope ideas. `PROGRESS.md` done, next, known issues.
