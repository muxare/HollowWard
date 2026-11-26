package content

// JobID is a citizen's trade. It sets their base stats and their one ability.
type JobID int

// The jobs.
const (
	Watchman JobID = iota
	Smith
	Priest
	Dockhand
	Miller
	Merchant
	Labourer
	NumJobs
)

// AbilityID is a job ability used in combat.
type AbilityID int

// The job abilities.
const (
	Hold   AbilityID = iota // watchman: guard an ally
	Hammer                  // smith: hit and stun
	Mend                    // priest: heal an ally
	Hook                    // dockhand: pull an enemy to the back of the order
	Grind                   // miller: one heavy blow
	Haggle                  // merchant: talk an enemy's Attack down
	Heave                   // labourer: hit every enemy for a little
	NumAbilities
)

// JobDef is the fixed part of a job.
type JobDef struct {
	ID      JobID
	Slug    string
	Name    string
	Ability AbilityID
}

// Stats are a job's base numbers: see T.JobStats.
func (j JobDef) Stats() Stats { return T.JobStats[j.ID] }

// Stats are the four combat numbers.
type Stats struct {
	HP     int
	Attack int
	Guard  int
	Speed  int
}

// Jobs is indexed by JobID.
var Jobs = [NumJobs]JobDef{
	Watchman: {Watchman, "watchman", "Watchman", Hold},
	Smith:    {Smith, "smith", "Smith", Hammer},
	Priest:   {Priest, "priest", "Priest", Mend},
	Dockhand: {Dockhand, "dockhand", "Dockhand", Hook},
	Miller:   {Miller, "miller", "Miller", Grind},
	Merchant: {Merchant, "merchant", "Merchant", Haggle},
	Labourer: {Labourer, "labourer", "Labourer", Heave},
}

// AbilityDef describes a job ability. Its numbers live in tuning.go.
type AbilityDef struct {
	ID   AbilityID
	Slug string
	Name string
	Text string
}

// Abilities is indexed by AbilityID.
var Abilities = [NumAbilities]AbilityDef{
	Hold:   {Hold, "hold", "Hold", "Guard an ally: blows meant for them land on you this round, and you take less."},
	Hammer: {Hammer, "hammer", "Hammer", "Strike one enemy and stun it: it loses its next action."},
	Mend:   {Mend, "mend", "Mend", "Heal one ally."},
	Hook:   {Hook, "hook", "Hook", "Strike one enemy and drag it down the order: it acts last next round."},
	Grind:  {Grind, "grind", "Grind", "One heavy blow that ignores Guard."},
	Haggle: {Haggle, "haggle", "Haggle", "Talk one enemy's Attack down for the rest of the fight."},
	Heave:  {Heave, "heave", "Heave", "Strike every enemy for a little."},
}

// TraitID is one of a citizen's two traits.
type TraitID int

// The traits.
const (
	Sturdy TraitID = iota
	Frail
	Quick
	Plodding
	Strong
	Wary
	Diligent
	Shirker
	Tireless
	Lucky
	Scavenger
	Steady
	NumTraits
)

// Mods are the numbers a trait or a scar adds to a citizen.
type Mods struct {
	HP      int // added to max HP
	Attack  int
	Guard   int
	Speed   int
	WorkPct int // added to work output, in percent
}

// TraitDef is a trait. Its numbers are T.TraitMods; the flags are rules.
type TraitDef struct {
	ID        TraitID
	Slug      string
	Name      string
	Text      string
	Tireless  bool // not slowed by exhaustion
	Lucky     bool // less likely to die when Downed
	Scavenger bool // more material from loot
	Steady    bool // cannot be stunned
}

// Mods are the trait's numbers.
func (t TraitDef) Mods() Mods { return T.TraitMods[t.ID] }

// Traits is indexed by TraitID. The numbers are small on purpose: traits
// colour a citizen, they do not make one.
var Traits = [NumTraits]TraitDef{
	Sturdy:    {ID: Sturdy, Slug: "sturdy", Name: "Sturdy", Text: "Hard to put down."},
	Frail:     {ID: Frail, Slug: "frail", Name: "Frail", Text: "Easy to put down."},
	Quick:     {ID: Quick, Slug: "quick", Name: "Quick", Text: "Acts early."},
	Plodding:  {ID: Plodding, Slug: "plodding", Name: "Plodding", Text: "Acts late."},
	Strong:    {ID: Strong, Slug: "strong", Name: "Strong", Text: "Hits harder."},
	Wary:      {ID: Wary, Slug: "wary", Name: "Wary", Text: "Takes less from every blow."},
	Diligent:  {ID: Diligent, Slug: "diligent", Name: "Diligent", Text: "Works harder by day."},
	Shirker:   {ID: Shirker, Slug: "shirker", Name: "Shirker", Text: "Works less by day."},
	Tireless:  {ID: Tireless, Slug: "tireless", Name: "Tireless", Text: "Works a full day after a night in the hollow.", Tireless: true},
	Lucky:     {ID: Lucky, Slug: "lucky", Name: "Lucky", Text: "Less likely to die when Downed.", Lucky: true},
	Scavenger: {ID: Scavenger, Slug: "scavenger", Name: "Scavenger", Text: "Brings home more material.", Scavenger: true},
	Steady:    {ID: Steady, Slug: "steady", Name: "Steady", Text: "Cannot be stunned.", Steady: true},
}

// ScarID is a permanent mark left on a citizen who was Downed and lived.
type ScarID int

// The scars.
const (
	Limp      ScarID = iota // slower in the hollow
	Shakes                  // weaker blows
	BadLung                 // less HP
	HollowEye               // works less
	NightFear               // works less
	NumScars
)

// ScarDef is a scar. Its numbers are T.ScarMods.
type ScarDef struct {
	ID   ScarID
	Slug string
	Name string
	Text string
}

// Mods are the scar's numbers.
func (s ScarDef) Mods() Mods { return T.ScarMods[s.ID] }

// Scars is indexed by ScarID. Each costs work output or a combat penalty.
var Scars = [NumScars]ScarDef{
	Limp:      {ID: Limp, Slug: "limp", Name: "Limp", Text: "Slower in the hollow."},
	Shakes:    {ID: Shakes, Slug: "shakes", Name: "The Shakes", Text: "Weaker blows."},
	BadLung:   {ID: BadLung, Slug: "bad-lung", Name: "Bad Lung", Text: "Goes down sooner."},
	HollowEye: {ID: HollowEye, Slug: "hollow-eye", Name: "Hollow-Eyed", Text: "Works less by day."},
	NightFear: {ID: NightFear, Slug: "night-fear", Name: "Night Fear", Text: "Works less by day."},
}

// CitizenNames are given names. A new tenure shuffles them.
var CitizenNames = []string{
	"Aldous", "Berit", "Corvin", "Dagny", "Edda", "Fenn", "Greta", "Hale",
	"Ilsa", "Jorun", "Kester", "Liesl", "Marek", "Nessa", "Osric", "Petra",
	"Quill", "Runa", "Sten", "Tove", "Ulf", "Vesna", "Wynn", "Ysolde",
	"Anselm", "Brida", "Cael", "Dorrit", "Emlyn", "Faro", "Gisla", "Hob",
	"Ingrid", "Jory", "Kelda", "Lorn", "Mab", "Nils", "Orla", "Piet",
}

// Opposite returns the trait that contradicts t, so nobody is both Sturdy and
// Frail. It reports false when t has no opposite.
func Opposite(t TraitID) (TraitID, bool) {
	switch t {
	case Sturdy:
		return Frail, true
	case Frail:
		return Sturdy, true
	case Quick:
		return Plodding, true
	case Plodding:
		return Quick, true
	case Diligent:
		return Shirker, true
	case Shirker:
		return Diligent, true
	}
	return 0, false
}
