// Package content holds the districts, enemies, abilities, events and text of
// Hollow Ward as typed Go data.
//
// This file, tuning.go, is the home of every number in the game. Balance work
// (milestone M6) changes this file and nothing else.
package content

// Tuning is every number the rules read. All arithmetic is integer.
type Tuning struct {
	// --- Tenure ---

	Nights int // nights a warden must survive to win

	// --- Needs and buildings ---

	NeedMax          int                             // needs and conditions run 0..NeedMax
	NeedBaseDecay    int                             // every need falls by this each day...
	NeedDecayAdj     [NumDistricts][NumNeeds]int     // ...adjusted by district type
	BuildingDecay    int                             // condition lost per building per day
	StartNeedMin     int                             // starting needs are drawn from
	StartNeedMax     int                             // [StartNeedMin, StartNeedMax]
	StartCondMin     int                             // starting building condition is drawn
	StartCondMax     int                             // from [StartCondMin, StartCondMax]
	StartResources   Cost                            // coin, grain, timber on day 1
	RuinThreshold    int                             // a building below this is ruined
	HavenThreshold   int                             // a building at or above this is a haven
	NeglectPerRuin   int                             // Neglect added per ruined building
	StatJitterHP     int                             // citizens' HP varies by up to this
	StatJitterSpeed  int                             // citizens' Speed varies by up to this
	StewardNeed      int                             // a steward adds this to one need...
	StewardBuilding  int                             // ...and this to one building
	Yield            [NumDistricts][NumResources]int // per resident per day at full output
	OutputFloorPct   int                             // output at mean need 0, in percent
	ExhaustedWorkPct int                             // output the day after fighting, in percent

	// --- Citizens ---

	JobStats  [NumJobs]Stats
	TraitMods [NumTraits]Mods
	ScarMods  [NumScars]Mods

	// --- Events ---

	Event EventMagnitudes

	// --- The hollow ---

	EnemyNeedBelow    int               // a need below this breeds its enemy family
	EnemyCountBase    int               // enemies per family: this...
	EnemyCountPer     int               // ...plus one per this many points of severity
	EnemyTier2At      int               // severity at which common enemies appear
	EnemyTier3At      int               // severity at which a greater enemy appears
	EnemyHPPer        int               // +1 HP per this many points of severity
	EnemyAttackPer    int               // +1 Attack per this many points of severity
	EnemyStats        [NumEnemies]Stats // base numbers per enemy
	MaxEnemiesPerRoom int
	BossStats         [NumDistricts]Stats // base numbers per district's boss
	BossHPPer         int                 // +1 boss HP per this many points the grievance is below NeedMax
	BossAttackPer     int                 // +1 boss Attack per this many points below NeedMax
	BossEscortPer     int                 // one escort per this many points of severity...
	BossEscortMax     int                 // ...up to this many
	HazardDamageBase  int                 // a hazard room hurts everyone by this...
	HazardDamagePer   int                 // ...plus 1 per this many points below the ruin threshold

	// --- Dusk and the expedition ---

	TurnThreshold    int // the most neglected district turns at this Neglect or more
	MaxParty         int // citizens in one expedition
	HavenHealPct     int // a safe room heals this share of max HP, in percent
	ClearRestoreNeed int // clearing the heart lifts each need to at least this
	HeartMaterials   int // materials of the grievance's family dropped by the heart
	HeartCoin        int // coin dropped by the heart
	DownedDeathPct   int // chance a Downed citizen dies at night's end, in percent
	LuckyDeathPct    int // the same for a Lucky citizen

	// --- Combat ---

	RoundCap        int // a fight that reaches this round drives the party out
	DefendGuard     int // Guard gained by defending
	AbilityCooldown int // rounds before a job ability can be used again
	Ability         AbilityNumbers
	Move            MoveNumbers

	// --- Decrees ---

	DecreeCost     [NumDecrees]Cost
	DecreeCityNeed int // Ration Grain, Curfew, Corvée: added to one need everywhere
	DecreeShoreUp  int // Shore Up: added to every building in one district
	DecreeRelief   int // Relief: added to all three needs in one district
}

// EventMagnitudes are the sizes event cards deal in. Cards name a size, never
// a number, so the whole deck can be retuned here.
type EventMagnitudes struct {
	XS, S, S6, M, L, XL, XXL   int // effect sizes
	CostLow, CostMid, CostHigh int // what a paid choice costs
}

// AbilityNumbers are the sizes of the job abilities.
type AbilityNumbers struct {
	HoldGuard    int // Guard the watchman gains while holding
	MendHeal     int // HP a priest restores
	GrindBonus   int // added to the miller's blow, which ignores Guard
	HaggleAmount int // Attack a merchant talks off an enemy
	HaggleFloor  int // lowest an enemy's Attack modifier can be haggled to
	HeavePenalty int // taken off the labourer's blow against each enemy
}

// MoveNumbers are the sizes of the enemy moves.
type MoveNumbers struct {
	GnawPenalty       int // Gnaw hits for this much less...
	GnawGuard         int // ...and takes this much Guard
	GorgeBonus        int
	ShovePenalty      int
	BurnDamage        int // Brand: lost at the start of the target's next turn
	SurgePenalty      int
	SporeAttack       int // Attack everyone loses to Spore
	MaulBonus         int
	EngulfBonus       int
	DevourBonus       int
	DevourHeal        int
	WitherAttack      int // Attack everyone loses to Wither
	RallyAttack       int // Attack every enemy gains from Rally
	BreakRanksPenalty int
	CollapseDamage    int // to everyone, ignoring Guard
	SporesGuard       int // Guard everyone loses to Spores
	ModFloor          int // lowest a citizen's Attack or Guard modifier can fall
}

// T is the tuning in force. Treat it as a constant: nothing writes to it.
var T = Tuning{
	Nights: 20,

	NeedMax:       100,
	NeedBaseDecay: 12,
	NeedDecayAdj: [NumDistricts][NumNeeds]int{
		//           Supply Order Upkeep
		Market:    {3, 0, -1},
		Watch:     {0, 3, 0},
		Docks:     {0, 2, 2},
		Temple:    {-2, -2, 2},
		Mills:     {1, 0, 3},
		Tenements: {2, 2, 1},
	},
	BuildingDecay:  4,
	StartNeedMin:   55,
	StartNeedMax:   85,
	StartCondMin:   45,
	StartCondMax:   90,
	StartResources: Cost{Coin: 20, Grain: 20, Timber: 10},
	RuinThreshold:  30,
	HavenThreshold: 70,
	NeglectPerRuin: 10,

	StatJitterHP:    2,
	StatJitterSpeed: 1,

	StewardNeed:     30,
	StewardBuilding: 15,

	Yield: [NumDistricts][NumResources]int{
		//           Coin Grain Timber
		Market:    {3, 0, 0},
		Watch:     {1, 0, 0},
		Docks:     {1, 0, 2},
		Temple:    {1, 1, 0},
		Mills:     {0, 3, 0},
		Tenements: {2, 0, 0},
	},
	OutputFloorPct:   50,
	ExhaustedWorkPct: 50,

	JobStats: [NumJobs]Stats{
		//         HP Attack Guard Speed
		Watchman: {16, 4, 2, 4},
		Smith:    {15, 5, 1, 3},
		Priest:   {11, 2, 0, 5},
		Dockhand: {14, 4, 1, 5},
		Miller:   {14, 4, 1, 3},
		Merchant: {10, 3, 0, 6},
		Labourer: {13, 4, 1, 4},
	},
	TraitMods: [NumTraits]Mods{
		Sturdy:   {HP: 4},
		Frail:    {HP: -3},
		Quick:    {Speed: 2},
		Plodding: {Speed: -1},
		Strong:   {Attack: 1},
		Wary:     {Guard: 1},
		Diligent: {WorkPct: 25},
		Shirker:  {WorkPct: -25},
	},
	ScarMods: [NumScars]Mods{
		Limp:      {Speed: -1},
		Shakes:    {Attack: -1},
		BadLung:   {HP: -3},
		HollowEye: {WorkPct: -20},
		NightFear: {WorkPct: -20},
	},

	Event: EventMagnitudes{
		XS: 4, S: 5, S6: 6, M: 8, L: 10, XL: 12, XXL: 15,
		CostLow: 5, CostMid: 6, CostHigh: 8,
	},

	EnemyNeedBelow: 50,
	EnemyCountBase: 1,
	EnemyCountPer:  10,
	EnemyTier2At:   17,
	EnemyTier3At:   34,
	EnemyHPPer:     10,
	EnemyAttackPer: 25,
	EnemyStats: [NumEnemies]Stats{
		//            HP Attack Guard Speed
		Gaunt:       {6, 3, 0, 6},
		Gnawer:      {9, 3, 1, 4},
		HollowBelly: {16, 5, 1, 2},
		AshMask:     {7, 3, 0, 5},
		BrandBearer: {9, 4, 0, 4},
		ManyHands:   {15, 3, 1, 3},
		Blackcap:    {5, 2, 1, 3},
		MouldHound:  {10, 4, 0, 5},
		Sump:        {18, 5, 2, 1},
	},
	MaxEnemiesPerRoom: 4,
	BossStats: [NumDistricts]Stats{
		Market:    {26, 4, 1, 4},
		Watch:     {30, 5, 2, 4},
		Docks:     {28, 5, 1, 5},
		Temple:    {24, 4, 1, 5},
		Mills:     {30, 5, 1, 3},
		Tenements: {26, 4, 1, 4},
	},
	BossHPPer:        6,
	BossAttackPer:    34,
	BossEscortPer:    20,
	BossEscortMax:    2,
	HazardDamageBase: 2,
	HazardDamagePer:  10,

	TurnThreshold:    40,
	MaxParty:         4,
	HavenHealPct:     50,
	ClearRestoreNeed: 50,
	HeartMaterials:   2,
	HeartCoin:        8,
	DownedDeathPct:   25,
	LuckyDeathPct:    12,

	RoundCap:        12,
	DefendGuard:     2,
	AbilityCooldown: 2,
	Ability: AbilityNumbers{
		HoldGuard: 2, MendHeal: 6, GrindBonus: 2, HaggleAmount: 2, HaggleFloor: -4, HeavePenalty: 2,
	},
	Move: MoveNumbers{
		GnawPenalty: 1, GnawGuard: 1, GorgeBonus: 2, ShovePenalty: 1, BurnDamage: 2,
		SurgePenalty: 2, SporeAttack: 1, MaulBonus: 1, EngulfBonus: 3,
		DevourBonus: 2, DevourHeal: 4, WitherAttack: 1, RallyAttack: 1,
		BreakRanksPenalty: 1, CollapseDamage: 3, SporesGuard: 1, ModFloor: -3,
	},

	DecreeCost: [NumDecrees]Cost{
		RationGrain: {Grain: 12},
		Curfew:      {Coin: 12},
		Corvee:      {Timber: 10},
		ShoreUp:     {Coin: 4, Timber: 6},
		Relief:      {Coin: 8, Grain: 8},
	},
	DecreeCityNeed: 20,
	DecreeShoreUp:  35,
	DecreeRelief:   15,
}
