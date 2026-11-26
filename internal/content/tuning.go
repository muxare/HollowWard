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
