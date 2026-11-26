# Hollow Ward: design

A dark fantasy management game with roguelike expeditions. This is the vertical slice, restated from the build prompt. Where the two disagree, the prompt wins.

## The loop

You are the warden of a walled city built as a lid over something that dreams. One tenure is 20 nights.

**Day.** Six districts each have three needs (Supply, Order, Upkeep) that fall every day, and two or three buildings that wear down. You have three stewards for six districts, so half the city slips every day. You also issue one decree and answer one event card.

**Dusk.** The most neglected district turns, if it is neglected enough. The dream remakes it as a dungeon, a hollow. The Dreamer meter rises by how badly that district was neglected. At 100 the lid breaks and the tenure ends.

**Night.** You send up to four citizens into the hollow. They move room to room and choose, after each room, to go deeper or go home. Clearing the heart room restores the district. Going home leaves it turned: it produces nothing and feeds the Dreamer every night until someone clears it.

You win by surviving 20 nights. You lose when the Dreamer reaches 100 or no citizen is left to fight.

## The one idea

The hollow is built from the district's state at the moment it turned, and nothing else.

| City state | Becomes |
| --- | --- |
| Supply below 50 | Hunger enemies. Lower Supply means more of them, and stronger. |
| Order below 50 | Riot enemies, scaled the same way. |
| Upkeep below 50 | Rot enemies, scaled the same way. |
| A building below 30 condition | A hazard room themed on that building. |
| A building at 70 or above | A safe room: rest, healing, no enemies. |
| The district's key figure | The boss. Their grievance is the district's lowest need, and it picks the boss's abilities. |
| The district type | The layout of the 6 to 10 rooms. |

Every generated room, enemy and boss carries a `Cause`: where it came from, which thing, and the value at generation. The UI shows it on hover and in the after-action report: "Gaunts ×4: Market Supply was 18."

This is the design's main risk. If the player cannot see why a hollow looks the way it does, the game is two unrelated halves. So the Neglect meter lists its causes largest first, the dusk screen previews what the Ledger already knows, and no generated thing exists without a cause. A fuzz test holds the generator to that.

## Neglect is a lever

Neglect is 100 minus the mean of the three needs, plus 10 per building below 30 condition. It decides which district turns and how fast the Dreamer rises.

Each enemy family drops its own material: hungerbone, ash-masks, blackcap. Materials buy upgrades in the city, and some upgrades need only one material. So starving the Market on purpose to farm hungerbone is sometimes the right play. The `farm` policy in the balance harness exists to prove that this works: it must win at least 80% as often as the policy that always repairs.

## People are the scarce resource

About 20 citizens live and work in the districts. Each has a job, four stats, one job ability and two traits. The same people run the city by day and fight at night.

- A citizen who fought works at half output the next day.
- A citizen at 0 HP is Downed. At night's end each Downed citizen has a 25% chance to die. Survivors carry a permanent Scar.
- The dead stay dead.

Sending your best smith into the hollow costs you the smithy tomorrow. That trade is the game.

## Combat stays small

Turn-based, initiative by Speed, party of 1 to 4. Enemies show their next action before they take it, so a fight is a readable puzzle and not a dice roll. A round cap ends every fight. No grid, no positioning beyond what abilities like Hook and Hold imply. The city layer is equally small: no building placement, no roads, no map editing.

## What persists

The Ledger survives across tenures. It records every enemy family, hazard and boss variant met, with the cause that produced it. It unlocks the dusk preview and grants no stats. A new warden inherits a reshuffled city (citizens, key figures, starting needs, new seed) and a better ability to read it.

## Story

The previous warden's last letter opens the game. Five fragments of their notes turn up as loot in safe rooms; all five unlock a closing text. The Dreamer speaks one line each night, and the lines escalate with the meter. Terse, heavy with dread, no detailed gore.

## How the code is shaped

- **`internal/sim`** is a pure, deterministic core. State is plain structs that marshal to JSON. Every change goes through `Apply(state, cmd) (State, []Event, error)`. It imports no engine, no `os`, no `time`. Lint rules enforce this.
- **Randomness** is `math/rand/v2` PCG streams kept in the state, one per subsystem (events, generation, combat), so changing one subsystem does not reshuffle the others, and a loaded save plays out the same future.
- **The event log** is the single source for the chronicle, the after-action report and the Ledger.
- **`internal/content`** is typed Go data. Every number lives in `tuning.go`, so balance work touches one file.
- **`internal/policy`** holds scripted players. They drive `simcli`, the balance harness and the smoke test.
- **`game/`** is the Godot presentation, written in Go through graphics.gd. It renders state, sends commands and never mutates state. Tooltips map straight from `Cause` values. It is the only package that imports graphics.gd, because that library warns of breaking changes.

Milestones M1 to M4 build and test the whole game headless. Godot arrives in M5 as a view onto a game that already works.

## Out of scope

Art, audio, warden classes or traits, more than six districts, procedural city layouts, localisation, web and mobile, controllers, runtime LLM calls. Ideas go in `IDEAS.md` and stay there.
