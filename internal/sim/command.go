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
	CmdBuyUpgrade    CommandKind = "buy-upgrade" // Upgrade, District and Building (for a ward)

	// Dusk and night.
	CmdSendParty CommandKind = "send-party" // District, Citizens
	CmdStayHome  CommandKind = "stay-home"  // let the night pass
	CmdAdvance   CommandKind = "advance"    // Room: go deeper
	CmdGoHome    CommandKind = "go-home"    // leave the hollow
	CmdAct       CommandKind = "act"        // Action, Target: the citizen whose turn it is acts
)

// What a citizen can do on their turn.
const (
	ActAttack  ActionKind = "attack"  // Target: an enemy's index in the room
	ActAbility ActionKind = "ability" // Target: an enemy's index, or a citizen's ID for Hold and Mend
	ActDefend  ActionKind = "defend"  // brace until your next turn
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
	Upgrade  content.UpgradeID  `json:"upgrade,omitempty"`
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

// BuyUpgrade spends materials on an upgrade. A ward names its building.
func BuyUpgrade(id content.UpgradeID, d content.DistrictID, building int) Command {
	return Command{Kind: CmdBuyUpgrade, Upgrade: id, District: d, Building: building}
}

// SendParty sends citizens into a turned district's hollow.
func SendParty(d content.DistrictID, citizens ...int) Command {
	return Command{Kind: CmdSendParty, District: d, Citizens: citizens}
}

// StayHome lets the night pass with nobody in a hollow.
func StayHome() Command { return Command{Kind: CmdStayHome} }

// Advance moves the party deeper, into one of the current room's exits.
func Advance(room int) Command { return Command{Kind: CmdAdvance, Room: room} }

// GoHome ends the expedition. The district stays turned.
func GoHome() Command { return Command{Kind: CmdGoHome} }

// Act is the acting citizen's move in a fight.
func Act(action ActionKind, target int) Command {
	return Command{Kind: CmdAct, Action: action, Target: target}
}

// Errors Apply returns. A rejected command leaves the state untouched.
var (
	ErrWrongPhase   = errors.New("not possible in this phase")
	ErrBadArgument  = errors.New("bad argument")
	ErrCannotAfford = errors.New("cannot afford it")
	ErrAlreadyDone  = errors.New("already done today")
	ErrEventPending = errors.New("today's event is not answered")
	ErrTurned       = errors.New("that district has turned")
	ErrInFight      = errors.New("not in the middle of a fight")
	ErrNoFight      = errors.New("there is no fight")
	ErrNotReady     = errors.New("that ability is not ready")
	ErrUnknown      = errors.New("unknown command")
)
