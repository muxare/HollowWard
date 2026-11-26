package sim

import (
	"errors"

	"hollowward/internal/content"
)

// CommandKind names what the player asks for.
type CommandKind string

// The commands of the day. Night commands are in night.go.
const (
	CmdNewTenure     CommandKind = "new-tenure"     // Seed
	CmdAssignSteward CommandKind = "assign-steward" // Steward, District, Need, Building
	CmdRecallSteward CommandKind = "recall-steward" // Steward
	CmdIssueDecree   CommandKind = "issue-decree"   // Decree, District (if targeted)
	CmdResolveEvent  CommandKind = "resolve-event"  // Choice
	CmdEndDay        CommandKind = "end-day"
)

// Command is one request from the player (or a scripted policy). It is plain
// data, so a list of commands is a replay.
type Command struct {
	Kind     CommandKind        `json:"kind"`
	Seed     uint64             `json:"seed,omitempty"`
	Steward  int                `json:"steward,omitempty"`
	District content.DistrictID `json:"district"`
	Need     content.Need       `json:"need"`
	Building int                `json:"building"`
	Decree   content.DecreeID   `json:"decree,omitempty"`
	Choice   int                `json:"choice,omitempty"`
	Citizens []int              `json:"citizens,omitempty"`
	Room     int                `json:"room,omitempty"`
	Actor    int                `json:"actor,omitempty"`
	Target   int                `json:"target,omitempty"`
	Action   ActionKind         `json:"action,omitempty"`
	Upgrade  int                `json:"upgrade,omitempty"`
}

// ActionKind is what a citizen does on their turn in combat.
type ActionKind string

// NewTenure starts a tenure from a seed.
func NewTenure(seed uint64) Command { return Command{Kind: CmdNewTenure, Seed: seed} }

// AssignSteward sends a steward to a district. NoNeed and building -1 let the
// steward pick the lowest need and the worst building.
func AssignSteward(steward int, d content.DistrictID, n content.Need, building int) Command {
	return Command{Kind: CmdAssignSteward, Steward: steward, District: d, Need: n, Building: building}
}

// RecallSteward leaves a steward idle.
func RecallSteward(steward int) Command {
	return Command{Kind: CmdRecallSteward, Steward: steward, District: content.NoDistrict}
}

// IssueDecree issues today's decree. Untargeted decrees ignore the district.
func IssueDecree(id content.DecreeID, target content.DistrictID) Command {
	return Command{Kind: CmdIssueDecree, Decree: id, District: target}
}

// ResolveEvent answers today's event card.
func ResolveEvent(choice int) Command { return Command{Kind: CmdResolveEvent, Choice: choice} }

// EndDay ends the day: stewards work, districts produce, needs decay.
func EndDay() Command { return Command{Kind: CmdEndDay} }

// Errors Apply returns. A rejected command leaves the state untouched.
var (
	ErrWrongPhase   = errors.New("not possible in this phase")
	ErrBadArgument  = errors.New("bad argument")
	ErrCannotAfford = errors.New("cannot afford it")
	ErrAlreadyDone  = errors.New("already done today")
	ErrEventPending = errors.New("today's event is not answered")
	ErrTurned       = errors.New("that district has turned")
	ErrUnknown      = errors.New("unknown command")
)
