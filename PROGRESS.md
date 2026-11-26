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

## Next

Waiting for go-ahead, and for answers to `QUESTIONS.md` (Q2, where the repository lives, blocks M1).

Then **M1: city core**. See `PLAN.md`.

## Known issues

- The repository has no remote yet (Q2).
- macOS and Windows builds cannot be produced or tested in the current build sandbox (Q3). Only Linux is proven.
- `go.sum` was produced without access to the Go checksum database. Run `go mod verify` once on a networked machine (`DECISIONS.md`, D10).
- The first `gd` run on a machine is slow: it builds a patched copy of the Go compiler and the engine bindings. About two minutes here on two cores. Later runs take seconds.
- `gd test` overwrites `graphics/linux_amd64.so` with a test build. The next `gd run` rebuilds it.
- `make balance` and `simcli` are stubs that exit 2 until M4 and M2.
