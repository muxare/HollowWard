# Open questions

Each has a default. Unless you say otherwise, I go with the default and record it in `DECISIONS.md`.

## Q1. Is the Dreamer meant to rise only?

Nothing in the spec lowers the Dreamer. With the starting values:

- A turned night adds at least 4.0, because a district only turns at Neglect 40 or more.
- A quiet night adds 5. So a quiet night costs more than a mild turning.
- 20 quiet nights add exactly 100, which breaks the lid on night 20. A perfectly kept city loses.
- The best possible tenure adds 80: a district turns at Neglect 40 every night and is cleared every night.
- A district left unstaffed loses 12 on every need each day, so a realistic turning sits at Neglect 50 to 70 and adds 5 to 7. Twenty of those is 100 to 140.

So the starting values are far from the balance targets. That is fine, since M6 tunes them. The structural question matters more, because M6 may only change `tuning.go`.

**Default:** rise-only, as written. I read "quiet costs more than a mild turning" as intended: it is pillar 3 in one number. To keep M6 honest I add one knob to `tuning.go`, `DreamerSootheOnClear`, set to 0. At 0 the game behaves exactly as the spec says. If the targets turn out unreachable without it, I stop and ask before raising it.

## Q2. Where should the repository live?

It is in a cloud workspace now, which does not outlive this session. I am sending you a git bundle of M0.

**Answered:** `github.com/muxare/HollowWard`. Pushing still needs the Claude GitHub App installed on that repository; until then commits stay local and I send git bundles.

## Q3. Who builds for macOS and Windows?

`gd` cross-compiles with zig, which it downloads from `ziglang.org`, and it downloads Godot export templates. This workspace can reach GitHub and nothing else that matters: `proxy.golang.org`, `go.dev`, `ziglang.org` and `graphics.gd` are all blocked. I worked around that for M0 (see `DECISIONS.md`, D10), but I cannot produce or test macOS and Windows builds from here.

**Default:** I build and test Linux here. macOS and Windows builds come from a GitHub Actions workflow that I write in M6, or from `make build` on your machine. Tell me which OS you will play on, so I check that one first.

Alternative: allow those domains for this workspace (Settings > Capabilities on claude.ai), and I do all three here.

## Q4. One expedition per night, into which hollow?

The spec lets districts stay turned across nights, so several hollows can be open at once. It does not say which one the party enters.

**Default:**

- One expedition per night. The player picks any turned district, tonight's or an older one. Sending nobody is allowed.
- A hollow is generated once, at the dusk its district turned, and stays as generated. Its causes stay true.
- Cleared rooms stay cleared. Going home is a retreat, not a reset.
- A turned district is frozen: its needs do not decay, it cannot be staffed, it produces nothing, and it cannot turn again. Its residents can still join expeditions.

## Q5. Who checks the 60 to 120 minute target?

I can count commands in an autoplayed tenure and estimate. Only a person can say whether a new player finishes in that time and can see why each thing exists.

**Default:** I estimate from command counts in M6 and flag the number as an estimate. You playtest the M5 build and the M6 build.
