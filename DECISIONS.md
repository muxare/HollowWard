# Decisions

Choices made where the spec was silent or contradicted itself. Newest last. Each is cheap to reverse unless it says otherwise.

## M0 (2026-11-26)

**D1. Module path is `hollowward`.** No hosting location is known yet. A rename is one `sed` once the repository has a home.

**D2. The entry point is `main.go` at the module root.** `gd` builds the package that sits beside the `graphics/` folder. The root package only calls `game.Main()`, so `game/` stays the single package that imports graphics.gd.

**D3. graphics.gd is pinned by commit, and so is `gd`.** Upstream has no version tags, so `go.mod` pins the pseudo-version `v0.0.0-20261126075435-1b2edc46bdb6` (head of the `release` branch on 2026-11-26). The spec installs the tool with `go install graphics.gd/cmd/gd@release`, which floats and contradicts "pin its version". Instead `go.mod` carries `tool graphics.gd/cmd/gd`, and the Makefile runs `go tool gd`. Library and tool now move together, by editing one line.

**D4. Godot 4.7.2.** It is the version this `gd` release downloads. `gd` accepts any 4.7.x on `PATH`.

**D5. The smoke test uses Godot's `--headless` flag, not `gd test`.** `make smoke` runs `gd run --headless -- --autoplay`. Both routes work with this toolchain; I tried each. `gd test` builds a Go test binary and runs it inside the engine, which suits unit tests that need engine types. The smoke test has to exercise the real build through the real UI command path, so it runs the real build. The exit code carries through: 0 on success and 1 on failure, both verified.

**D6. Renderer is `gl_compatibility`, window 1280×720, stretch mode `canvas_items`.** The slice is 2D placeholder shapes and text. The compatibility renderer runs on the widest range of machines and on a software rasteriser, which is how `make shot` takes window screenshots on a virtual display.

**D7. No GDScript.** Nothing in M0 needed it.

**D8. How the purity rules are enforced.**

- depguard, non-test files in `internal/sim`: no `os`, no `time`, no `math/rand` (v1). Test files may use `os` for fixtures.
- depguard, all files in `internal/sim`: no graphics.gd, no `hollowward/game`.
- depguard, everything outside `game/`: no graphics.gd.
- forbidigo: `time.Now` is banned everywhere except `cmd/` and `game/`, which may time a run or pick a fresh seed. The global functions of `math/rand/v2` and all of `math/rand` v1 are banned everywhere. Methods on a seeded `*rand.Rand` are allowed.

Each rule was checked against a deliberate violation before the files were removed.

**D9. golangci-lint 2.15.0 or newer.** It must be built with Go 1.27 or it refuses to load a Go 1.27 module.

**D10. How the toolchain was assembled in the build sandbox.** The sandbox reaches GitHub but not `proxy.golang.org`, `go.dev`, `graphics.gd` or `ziglang.org`. So:

- Go 1.27.4 was built from the `go1.27.4` tag of `github.com/golang/go`.
- Godot 4.7.2 came from the official GitHub release, the same URL `gd` uses.
- graphics.gd and its dependencies were cloned from GitHub at the pinned versions and served from a local file-based module proxy.
- All 26 dependency archives matched the checksums in graphics.gd's own upstream `go.sum`. The checksum for graphics.gd itself was computed from the GitHub commit, with no checksum database to compare against.

None of this is in the repository. `go.mod` and `go.sum` are standard. **Action for a networked machine:** run `go mod verify` once. It should pass; if it reports a mismatch for graphics.gd, tell me.

**D11. `make balance` and `simcli` exit 2 until they exist.** A stub that exits 0 would look like a pass.

## Rule defaults planned for M1 to M4

These follow from the questions in `QUESTIONS.md` and take effect when the code that needs them is written.

**D12. The sim uses integers only.** The Dreamer is stored in tenths (0 to 1000), so "Neglect divided by 10" needs no floating point. This keeps results identical across machines and compilers.

**D13. Neglect is not capped at 100.** Three ruined buildings can push it to 130. The meter shows the overflow, since hiding it would hide a cause.

**D14. "No citizen can fight" means no citizen is alive.** Downed citizens recover by morning or die at night's end, so no other state blocks a party.

**D15. Order within a night.** The Dreamer rises at dusk. If it reaches 100 the tenure ends there, before any expedition. The win is checked after night 20 resolves.

**D16. New-tenure seeds come from outside the sim.** `sim` cannot read a clock or OS entropy. `game/` and the CLIs pass the seed in with the command that starts a tenure.
