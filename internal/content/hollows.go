package content

// EnemyID names a kind of enemy. Each family has three, in rising weight.
type EnemyID int

// The enemies.
const (
	Gaunt       EnemyID = iota // Hunger, lesser
	Gnawer                     // Hunger, common
	HollowBelly                // Hunger, greater
	AshMask                    // Riot, lesser
	BrandBearer                // Riot, common
	ManyHands                  // Riot, greater
	Blackcap                   // Rot, lesser
	MouldHound                 // Rot, common
	Sump                       // Rot, greater
	NumEnemies
)

// EnemyDef is a kind of enemy. Its numbers are T.EnemyStats.
type EnemyDef struct {
	ID     EnemyID
	Slug   string
	Name   string
	Plural string
	Family Family
	Tier   int // 1 lesser, 2 common, 3 greater
	Text   string
}

// Stats are the enemy's base numbers before the need's severity is added.
func (e EnemyDef) Stats() Stats { return T.EnemyStats[e.ID] }

// Enemies is indexed by EnemyID.
var Enemies = [NumEnemies]EnemyDef{
	Gaunt:       {Gaunt, "gaunt", "Gaunt", "Gaunts", Hunger, 1, "Thin, quick, and never full."},
	Gnawer:      {Gnawer, "gnawer", "Gnawer", "Gnawers", Hunger, 2, "It eats what holds things up."},
	HollowBelly: {HollowBelly, "hollow-belly", "Hollow-Belly", "Hollow-Bellies", Hunger, 3, "A larder with legs. Empty."},
	AshMask:     {AshMask, "ash-mask", "Ash-Mask", "Ash-Masks", Riot, 1, "A neighbour's face, burned on."},
	BrandBearer: {BrandBearer, "brand-bearer", "Brand-Bearer", "Brand-Bearers", Riot, 2, "It carries the fire from door to door."},
	ManyHands:   {ManyHands, "many-hands", "Many-Hands", "Many-Hands", Riot, 3, "A crowd that has stopped being people."},
	Blackcap:    {Blackcap, "blackcap", "Blackcap", "Blackcaps", Rot, 1, "It grows where nobody sweeps."},
	MouldHound:  {MouldHound, "mould-hound", "Mould-Hound", "Mould-Hounds", Rot, 2, "Something's dog, once. Soft now."},
	Sump:        {Sump, "sump", "Sump", "Sumps", Rot, 3, "The cellar, standing up."},
}

// EnemyOf is the enemy of a family at a tier (1 to 3).
func EnemyOf(f Family, tier int) EnemyID { return EnemyID(int(f)*3 + tier - 1) }

// Twist is what a grievance does to a key figure: the adjective it earns
// them and the abilities it gives the boss they become.
type Twist struct {
	Need      Need
	Adjective string
	Text      string
	Moves     [2]MoveID // the boss's two special moves; it also strikes
}

// Twists is indexed by the need that became the grievance.
var Twists = [NumNeeds]Twist{
	Supply: {Supply, "Starved", "Went without so long that the wanting is all that is left.", [2]MoveID{MoveDevour, MoveWither}},
	Order:  {Order, "Unpaid", "Kept the peace for nothing, and has come to collect.", [2]MoveID{MoveRally, MoveBreakRanks}},
	Upkeep: {Upkeep, "Crumbling", "Held the roof up alone. The roof has come down.", [2]MoveID{MoveCollapse, MoveSpores}},
}

// BossName is what a district's key figure is called once twisted.
func BossName(d DistrictID, grievance Need) string {
	return "The " + Twists[grievance].Adjective + " " + Districts[d].FigureTitle
}

// BossSlug identifies a boss variant: one per district, twisted three ways.
func BossSlug(d DistrictID, grievance Need) string {
	return Districts[d].Slug + "-boss/" + grievance.Slug()
}

// MoveID is something an enemy can do on its turn. Enemies show the move
// they will make before they make it.
type MoveID int

// The moves.
const (
	MoveStrike     MoveID = iota // hit one citizen
	MoveLunge                    // Gaunt: hit the citizen with the least HP
	MoveGnaw                     // Gnawer: hit and lower the target's Guard
	MoveGorge                    // Hollow-Belly: heavy hit
	MoveShove                    // Ash-Mask: hit and delay the target
	MoveBrand                    // Brand-Bearer: hit one, burn them for a round
	MoveSurge                    // Many-Hands: hit everyone lightly
	MoveSpore                    // Blackcap: lower everyone's Attack a little
	MoveMaul                     // Mould-Hound: hit the citizen with the most HP
	MoveEngulf                   // Sump: heavy hit, then it must gather itself
	MoveGather                   // do nothing this turn
	MoveDevour                   // boss, Starved: heavy hit, heals itself
	MoveWither                   // boss, Starved: lower everyone's Attack
	MoveRally                    // boss, Unpaid: every enemy hits harder
	MoveBreakRanks               // boss, Unpaid: hit everyone
	MoveCollapse                 // boss, Crumbling: hit everyone, ignores Guard
	MoveSpores                   // boss, Crumbling: lower everyone's Guard
	NumMoves
)

// MoveDef describes a move. Its numbers live in tuning.go.
type MoveDef struct {
	ID   MoveID
	Name string
	Text string // shown as the enemy's intent
}

// Moves is indexed by MoveID.
var Moves = [NumMoves]MoveDef{
	MoveStrike:     {MoveStrike, "Strike", "will strike %s"},
	MoveLunge:      {MoveLunge, "Lunge", "will lunge at the weakest: %s"},
	MoveGnaw:       {MoveGnaw, "Gnaw", "will gnaw at %s's guard"},
	MoveGorge:      {MoveGorge, "Gorge", "will fall on %s with its whole weight"},
	MoveShove:      {MoveShove, "Shove", "will shove %s to the back"},
	MoveBrand:      {MoveBrand, "Brand", "will set %s alight"},
	MoveSurge:      {MoveSurge, "Surge", "will surge over everyone"},
	MoveSpore:      {MoveSpore, "Spore", "will burst and weaken everyone's blows"},
	MoveMaul:       {MoveMaul, "Maul", "will maul the strongest: %s"},
	MoveEngulf:     {MoveEngulf, "Engulf", "will engulf %s"},
	MoveGather:     {MoveGather, "Gather", "is gathering itself"},
	MoveDevour:     {MoveDevour, "Devour", "will devour %s and grow stronger"},
	MoveWither:     {MoveWither, "Wither", "will breathe out and weaken everyone's blows"},
	MoveRally:      {MoveRally, "Rally", "will rally the others to hit harder"},
	MoveBreakRanks: {MoveBreakRanks, "Break Ranks", "will charge through everyone"},
	MoveCollapse:   {MoveCollapse, "Collapse", "will bring the ceiling down on everyone"},
	MoveSpores:     {MoveSpores, "Spores", "will rot everyone's guard"},
}

// EnemyMoves is each enemy's repertoire. An enemy cycles through it.
var EnemyMoves = [NumEnemies][]MoveID{
	Gaunt:       {MoveLunge, MoveStrike},
	Gnawer:      {MoveStrike, MoveGnaw},
	HollowBelly: {MoveGorge, MoveStrike, MoveStrike},
	AshMask:     {MoveStrike, MoveShove},
	BrandBearer: {MoveBrand, MoveStrike},
	ManyHands:   {MoveSurge, MoveStrike},
	Blackcap:    {MoveStrike, MoveSpore, MoveStrike},
	MouldHound:  {MoveMaul, MoveStrike},
	Sump:        {MoveEngulf, MoveGather, MoveStrike},
}

// Layout is a district's hollow template: its core rooms and where the
// district's buildings attach. Room 0 is the entry; the last room is the
// heart. Every edge leads deeper, and every path ends at the heart.
type Layout struct {
	Rooms []string // names of the core rooms
	Edges [][2]int // from, to
	// Slots[i] is the core room that building i's special room follows: the
	// special room is spliced into every way out of that core room.
	Slots []int
}

// Layouts is indexed by DistrictID: one template per district.
var Layouts = [NumDistricts]Layout{
	Market: {
		Rooms: []string{"The Empty Square", "Butchers' Row", "Chandlers' Row", "The Weigh-House", "The Fish Slabs", "The Cloth Hall", "The Factor's Table"},
		Edges: [][2]int{{0, 1}, {0, 2}, {1, 3}, {2, 3}, {3, 4}, {3, 5}, {4, 6}, {5, 6}},
		Slots: []int{0, 3, 1},
	},
	Watch: {
		Rooms: []string{"The Outer Yard", "The Drill Square", "The Armoury", "The East Wall-Walk", "The West Wall-Walk", "The Captain's Post"},
		Edges: [][2]int{{0, 1}, {1, 2}, {2, 3}, {2, 4}, {3, 5}, {4, 5}},
		Slots: []int{1, 2},
	},
	Docks: {
		Rooms: []string{"The Mole", "Salt Pier", "Tar Pier", "Net Pier", "The Chandlery", "The Ropewalk", "The Harbourmaster's Light"},
		Edges: [][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 4}, {2, 4}, {3, 5}, {4, 6}, {5, 6}},
		Slots: []int{1, 3, 4},
	},
	Temple: {
		Rooms: []string{"The Porch", "The Nave", "The North Aisle", "The South Aisle", "The Choir", "The Prior's Chair"},
		Edges: [][2]int{{0, 1}, {1, 2}, {1, 3}, {2, 4}, {3, 4}, {4, 5}},
		Slots: []int{0, 4},
	},
	Mills: {
		Rooms: []string{"The Race", "The Sluice", "The West Wheel", "The Sack Loft", "The Chaff Floor", "The Stone Floor", "The Reeve's Tally"},
		Edges: [][2]int{{0, 1}, {1, 2}, {0, 3}, {3, 4}, {2, 5}, {4, 5}, {5, 6}},
		Slots: []int{2, 4, 5},
	},
	Tenements: {
		Rooms: []string{"The Yard", "The Front Stair", "The Back Stair", "The Garrets", "The Shared Landing", "The Cellars", "The Top Floor", "The Rentmaster's Door"},
		Edges: [][2]int{{0, 1}, {0, 2}, {1, 3}, {1, 4}, {2, 4}, {2, 5}, {3, 6}, {4, 6}, {5, 6}, {6, 7}},
		Slots: []int{0, 6},
	},
}
