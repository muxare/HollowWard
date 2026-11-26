# Decisions

Choices made where the spec was silent or contradicted itself. Newest last. Each is cheap to reverse unless it says otherwise.

## M0 (2026-11-26)

**D1. Module path is `hollowward`.** The repository is `github.com/muxare/HollowWard`. The module keeps the short path: it is an application, nothing imports it, and a mixed-case module path is awkward in Go.

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

None of this is in the repository. `go.mod` and `go.sum` are standard. The one open point, whether the locally computed checksum for graphics.gd matches the real one, was closed in M6 (D76).

**D11. `make balance` and `simcli` exit 2 until they exist.** A stub that exits 0 would look like a pass.

## Rule defaults planned for M1 to M4

These follow from the questions in `QUESTIONS.md` and take effect when the code that needs them is written.

**D12. The sim uses integers only.** The Dreamer is stored in tenths (0 to 1000), so "Neglect divided by 10" needs no floating point. This keeps results identical across machines and compilers.

**D13. Neglect is not capped at 100.** Three ruined buildings can push it to 130. The meter shows the overflow, since hiding it would hide a cause.

**D14. "No citizen can fight" means no citizen is alive.** Downed citizens recover by morning or die at night's end, so no other state blocks a party.

**D15. Order within a night.** The Dreamer rises at dusk. If it reaches 100 the tenure ends there, before any expedition. The win is checked after night 20 resolves.

**D16. New-tenure seeds come from outside the sim.** `sim` cannot read a clock or OS entropy. `game/` and the CLIs pass the seed in with the command that starts a tenure.

## M1 (2026-11-26)

**D17. The day tick runs in a fixed order: stewards work, districts produce, needs and buildings decay.** A steward's visit therefore lifts today's output, and decay is the last thing that happens before dusk reads Neglect. Decrees and event choices apply the moment they are chosen.

**D18. The day cannot end with the event card unanswered.** Every card has at least one free choice (a test holds the deck to that), so a card can never block the day. The deck is dealt without replacement and reshuffled when empty: 12 cards over 20 nights means every card is seen once before any repeats.

**D19. Stewards keep their posts across days, and may share a district.** A post with no need or building named resolves to the lowest need and the worst building at the moment the steward works, not when the order is given.

**D20. Output.** A district yields `yield x labour x need factor`. Labour is the sum of its residents' work percentages (traits and scars shift it; a citizen who fought last night works at 50%; the dead do no work). The need factor runs from 50% at mean need 0 to 100% at mean need 100. The spec says only "according to working citizens and need levels".

**D21. Resources are spent on decrees, paid event choices and (from M4) upgrades.** There is no upkeep or consumption. That keeps the city layer small (pillar 6) while still making the choice of decree depend on what the districts produced.

**D22. A turned district is out of reach.** Stewards, decrees and event effects do nothing there, its needs and buildings do not decay, and it produces nothing. This is the Q4 default.

**D23. "Every number in tuning.go" includes the tables.** Job stats, trait and scar modifiers, decree costs and the sizes event cards deal in all live in `tuning.go`. Event cards name a size (`e.M`, `e.CostLow`), never a number. Card text and structure stay in `events.go`.

**D24. The event log is returned by `Apply` and kept by the caller, not stored inside `State`.** `Apply` copies the state it is given so that it never modifies its input. Carrying a growing log inside the state would make every command cost more as the tenure goes on. A save holds the state and the log side by side (M4).

**D25. Five more decisions the spec left open.** Citizens get two different traits that are not opposites (nobody is both Sturdy and Frail). Starting needs are drawn from 55 to 85 and building conditions from 45 to 90. Neglect rounds to the nearest point, and its causes are held in thirds of a point so the three needs share "100 minus the mean" exactly. The grievance and the worst building break ties toward the first in the fixed order.

## M2 (2026-11-26)

**D26. What each cause value means.** A need cause records the need's value. A building cause records its condition. The figure cause records the value of the district's lowest need, which is the grievance. The district cause records the district's ID. `Cause.Resolve` looks any of them up in a district, and the tests use it to prove that every cause points at real state.

**D27. A need at or above 50 breeds nothing.** A district can turn at Neglect 40 with every need above 50, for example on ruined buildings. Its hollow then has empty rooms, hazards and the boss, and the empty rooms say why. This also means farming a material requires actually starving the need below 50.

**D28. How many, and how strong.** A need below 50 has a severity of 50 minus its value. It breeds 1 enemy plus 1 per 10 points of severity (so Supply 18 gives four, as in the spec's example). From severity 17 every second one is the common kind; from 34 the first is the greater kind. Every enemy also gains HP and Attack with severity. The stream decides only which room each enemy waits in, never what exists.

**D29. Building rooms sit on the path.** Each building has a slot in its district's layout. A hazard or safe room is spliced into every way out of that slot's room, so a ruined building always costs the party something and a kept one always helps. The spec allows moving only deeper or home, so a side room would be a dead end.

**D30. Room counts.** Layouts have 6 to 8 core rooms and districts have 2 or 3 buildings, so every hollow has 6 to 10 rooms whatever the buildings became.

**D31. Boss variants.** Six bosses, one per district's key figure, each twisted one of three ways by the grievance: Starved (Supply), Unpaid (Order), Crumbling (Upkeep). The twist picks the boss's two special moves. The boss grows with how low the grievance is, and brings up to two escorts of the grievance's family when that need is below 50.

**D32. A hazard room costs HP.** Everyone who enters loses 2 HP, plus 1 per 10 points the building is below 30. Hazards hold no enemies.

## M3 (2026-11-26)

**D33. Every night passes through a dusk phase.** After the day tick the game stops at dusk, even on a quiet night. The player sends a party into any open hollow or lets the night pass. This is the Q4 default: one expedition per night, into a hollow of the player's choosing.

**D34. Enemies show two actions: this round's and the next round's.** With only one shown, a fast enemy could act at the end of one round and again at the start of the next before any citizen chose anything, and its second action would never have been on show. With two, every enemy action that can happen before a citizen's next choice is visible when they choose. A test checks this for every enemy action in whole tenures. The one exception is unavoidable: an enemy faster than every citizen acts in round 1 before anyone has chosen.

**D35. Initiative uses no dice.** Fastest first. A tie goes to the citizen, then to the lower party slot or enemy index. Hook and Shove move their target to the back of the next round.

**D36. No fleeing in the middle of a fight.** The spec offers "go deeper or go home" after each room. Inside a fight the only exits are winning, the whole party going down, or the round cap.

**D37. The round cap is 12.** A fight that reaches round 13 ends with the hollow driving the party out, which counts as going home. Enemies keep the wounds they took.

**D38. A wiped party is carried out.** If every party member is Downed the night ends. Each then rolls for death like any other Downed citizen.

**D39. Damage is fixed, not rolled.** A blow does Attack minus Guard, at least 1. The combat stream only picks targets for enemy moves that have no rule for it. Fights stay readable and the intent on show is what happens.

**D40. Abilities need a turn of rest between uses.** A boss cannot be stunned.

**D41. Everyone wakes at full HP.** Wounds do not carry over to the next night. The lasting costs of a night are the ones the spec names: exhaustion, Scars and deaths. Only citizens who were in a fight are exhausted; walking through empty rooms tires nobody.

**D42. Clearing the heart lifts needs to 50; it never lowers one.** A need already above 50 stays where it was.

**D43. Materials drop one per kill,** plus two of the grievance's family and 8 coin from the heart. A Scavenger in the party finds one more per room cleared.

**D44. A hazard room catches a party once.** After that it is cleared like any other room. Cleared rooms stay cleared across nights (Q4 default).

**D45. The scripted wardens look one step ahead.** `repair-worst` and `farm` answer cards and choose decrees by applying each option to a copy of the state and comparing the result. They send the four hardiest citizens and go home when the party is badly hurt. `random` tries random orders and sends whichever the rules accept.

## M4 (2026-11-26)

**D46. The Dreamer is kept in tenths of a point (D12) and only rises (Q1 default).** `DreamerSootheOnClear` exists in `tuning.go` and is 0. Each dusk records why the meter rose, part by part, so the dusk screen can show it.

**D47. When the tenure ends.** The lid breaks at dusk, before any expedition, and nights survived does not count that night. The win is checked at dawn after night 20. Losing every citizen is checked at dawn and takes precedence over the win.

**D48. Seven upgrades.** A ward for one building (3 blackcap, repeatable per building) halves its decay. Three kinds of gear change every citizen: Bone Charms (+3 HP, 6 hungerbone), Ash Hoods (+1 Guard, 6 ash-masks), Bone-and-Ash Blades (+1 Attack, 4 hungerbone and 4 ash-masks). Three slow one need's decay by 2 everywhere: Granary Seals (5 hungerbone), Curfew Bells (5 ash-masks), Black Mortar (5 blackcap). Six of the seven take a single material. Upgrades last one tenure; nothing but the Ledger crosses to the next (pillar 5).

**D49. What the Ledger holds.** One entry per enemy kind, per hazard (district and building) and per boss variant (district and grievance), each with the cause that produced it the first time, how often it has been met and in which tenure it was first met. It also holds the fragments found and the count of tenures ended and won. It is built by `Ledger.Record(events)`, which reads the event log and nothing else.

**D50. The Ledger cannot grant stats, by construction.** `Apply` does not take a Ledger. The rules cannot read one.

**D51. The dusk preview hides names, never causes.** For each thing in the hollow the preview shows its cause, since the cause is the city's own state and the warden can already see it. The name of the enemy, hazard or boss appears only if the Ledger records it; otherwise the line reads "Something not in the Ledger". Safe rooms are always named: the warden knows which buildings they kept.

**D52. Fragments.** The first visit to each safe room turns up the next of the five notes, at most five per tenure. The Ledger keeps the set across tenures; all five unlock the closing text.

**D53. A save is the state and the event log, with a version number.** A save from another version is refused with a clear error, not guessed at. `sim` encodes and decodes bytes; it never opens a file.

**D54. The chronicle merges quiet nights** ("Nights 1 to 3 were quiet") and gives every other night one paragraph: what fell into ruin, what turned and why, who went in and how it ended, who died and who came back marked.

**D55. `make balance` prints the targets but does not fail on them yet.** `balance -check` exits 1 when a target is missed; the Makefile starts using it in M6.

## M5 (2026-11-26)

**D56. The screen is rebuilt from the state after every command.** `game/` keeps no copy of game state. It holds the `sim.Game`, the Ledger and a few view choices (which district is open, who is picked for the party), and redraws everything when any of them changes. It can therefore never show a state the sim is not in.

**D57. What "the real UI command path" means for the smoke test.** Every button has an action id (`end-day`, `assign:0:market:-1:-1`, `act:attack:2`). A click runs `App.press(id)`. Autoplay also runs `App.press(id)`, with the ids of the commands its script wants. If the screen does not carry an enabled button with that id, the smoke test fails. So the test proves that each step is offered by the screen the player would be on, and that the button's own code sends the command. It does not prove pixel hit-testing: nothing moves a mouse.

**D58. A button is enabled only if the rules accept its command now.** The UI asks the sim, by applying the command to a copy of the state. A disabled button's hover text is the rule's own reason ("Cannot afford it: Curfew."). The UI has no second copy of the rules to drift out of step.

**D59. The district panel shows "if it turned tonight".** It runs the real generator on the district as it stands, with a throwaway stream, and shows the result through the Ledger preview. The player sees the mapping from city state to hollow before dusk, not only after.

**D60. Tooltips come from `Cause` values.** Enemies, rooms, preview lines, Neglect causes and Ledger entries all render `Cause.Describe()`. The after-action report is the night's events as text, each with its cause; the blow-by-blow of fights is left out.

**D61. Saving.** The game autosaves to `user://save.json` after every command and keeps the Ledger in `user://ledger.json`. The title screen offers "Continue". Autoplay uses its own two files, an empty Ledger and a fixed seed, so it never touches a player's save and always plays the same city.

**D62. Hints show during the first tenure only,** which is until the Ledger records a finished tenure. One line per screen.

**D63. Upgrades have their own screen; wards are bought beside the building.** The city screen had no room for seven more buttons at 1280 by 720.

**D64. Placeholder art.** One Polygon2D per district as its icon. On the node map: squares for rooms, diamonds for hazards, octagons for safe rooms, a spiked shape for the heart. Colour comes from one Theme and a small palette.

**D65. District names agree with their verbs.** "The Market turns", "the Docks turn".

## M6 (2026-11-26)

**D66. The balance was met by changing `tuning.go` only.** The rules, the policies and the harness are as they were at the end of M5. Five groups of numbers moved:

| Number | Spec's starting value | Now | Why |
| --- | --- | --- | --- |
| Building decay per day | 4 | 2 | At 4 a district left to starve had every building in ruins within two weeks. Each ruin adds 10 Neglect, so the farmed district turned every single night at Neglect 58 and the `farm` policy never won. This one number made farming viable. |
| Decree costs | 12 grain, 12 coin, 10 timber, 4 coin + 6 timber, 8 coin + 8 grain | 30 grain, 36 coin, 20 timber, 8 coin + 12 timber, 16 coin + 16 grain | At the old prices the city could afford a decree every day, and nothing ever slipped far. Now it is one every two or three days. |
| Needs after a heart is cleared | 50 | 65 | At 50 a cleared district was back at Neglect 60 by the next dusk and turned again at once. |
| Boss HP and Attack | 24 to 30 HP, 4 to 5 Attack; +1 HP per 6 points of grievance | 54 to 64 HP, 7 to 8 Attack; +1 HP per 3 points | The old bosses fell in three rounds and never downed anyone. Now a heart is a real fight, parties sometimes come home without clearing it, and that is what loses tenures. |
| Enemy growth with severity | +1 HP per 10, +1 Attack per 25 | +1 HP per 6, +1 Attack per 15 | Deeply starved needs breed enemies worth fearing. |

The spec's own numbers for the Dreamer are unchanged: Neglect over ten per turning, 5 on a quiet night, 15 per district still turned, the lid at 100. So are the Neglect formula, the turn threshold of 40, need decay of 12, the steward's +30 and +15, the 25% death chance and exhaustion at 50%. Enemy counts are unchanged too, so Supply 18 still breeds four, as in the spec's example.

**D67. Results, 1,000 tenures per policy, seeds 1 to 1,000.** `random` 0.0%, `repair-worst` 41.0%, `farm` 57.6% (the bar is 80% of `repair-worst`, which is 32.8%). On three other ranges of 1,000 seeds `repair-worst` came out at 37.6%, 41.8% and 41.9%, and `farm` at 58.6%, 59.2% and 59.4%. `make balance` now runs with `-check` and fails if a target is missed.

**D68. Q1 is settled: the Dreamer only rises.** `DreamerSootheOnClear` stayed at 0. The targets were reachable without it. With the spec's values a quiet night (5) costs about what a mild turning does (4 to 5), so a controlled turning costs nothing extra on the meter. That is what makes neglect a lever, and it is why those numbers were left alone.

**D69. `farm` wins more often than `repair-worst`.** The spec asks for at least 80%; it gets about 140%. Leaving one district alone frees the three stewards for five districts, and its hollows pay for gear and seals. A stronger `repair-worst` would close the gap, but the policies are frozen with the rules. A human can do what `farm` does.

**D70. Two test setups changed with the tuning, and no assertion was loosened.** One test issued a decree with the starting stores, which no longer cover one; it now sets the stores it needs. The test helper that makes citizens who "cannot lose" gave them only HP, and a tougher boss could outlast them to the round cap; it now gives them Attack as well. Both tests check what they checked before.

**D71. Fragments: two per tenure.** Collecting all five now takes at least three tenures, so the closing text is a reward for handing the Ledger on, not for one good night.

**D72. Desktop builds.** `make build-linux` works in the build sandbox and the resulting binary passes the smoke script, headless and in a window. macOS and Windows builds need the zig compiler from `ziglang.org`, which the sandbox cannot reach, so they come from CI. `.github/workflows/build.yml` runs `go mod verify`, `make check`, `make balance` and `make smoke` on every push, and builds all three targets on a tag or on demand. It has run: the checks pass on GitHub's runner, and one run produced Linux, Windows and macOS builds as artifacts. The Windows and macOS builds have not been started by anyone.

**D76. The checksum caveat in D10 is closed.** `go mod verify` passes on GitHub's runner, against the real module proxy and checksum database. The `go.sum` assembled in the sandbox is correct.

**D77. Cross-builds resolve the host's `gd` first.** `GOOS=macos go tool gd build` made `go` cross-compile the `gd` tool itself, which then could not run. The build targets call `go tool -n gd` to get the host's binary and pass it the target.

**D73. Playing time is estimated, not measured.** The harness counts the decisions in each full tenure and prices them (45 seconds for a day in the city, 7 for a combat turn, and so on): about 65 minutes as `repair-worst` plays, about 89 as `farm` plays. Both are inside the 60 to 120 minute target. Only a person playing can confirm it.

**D74. A second smoke test plays a whole tenure through the UI.** `make smoke-full` has the careful warden play to the end of a tenure by pressing the UI's buttons, walking to whichever screen offers each command: the right district panel, the upgrades screen, the right hollow, the right citizens. It passed on nine seeds (three wins, six losses). It is the nearest thing to a playtest the sandbox allows.

**D75. `graphics/graphics.gen.go` is ignored by git.** `gd build` writes it into its own folder: a generated Go file that lists the Godot project's resources and imports graphics.gd. Nothing imports it, the build regenerates it, and it would be a second package importing graphics.gd. `game/` remains the only hand-written one.
