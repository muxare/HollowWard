package content

// EffectKind says what an Effect changes.
type EffectKind int

// The effect kinds.
const (
	EffNeed     EffectKind = iota // change a need (NoNeed means all three)
	EffBuilding                   // change every building's condition
	EffResource                   // change a city resource
)

// Scope says which districts an Effect reaches.
type Scope int

// The scopes.
const (
	ScopeCity   Scope = iota // not tied to a district (resources)
	ScopeTarget              // the decree's or card's target district
	ScopeAll                 // every district that has not turned
	ScopeFixed               // the district named in Effect.District
)

// Effect is one change a decree or an event choice makes.
type Effect struct {
	Kind     EffectKind
	Scope    Scope
	District DistrictID // for ScopeFixed
	Need     Need       // for EffNeed
	Resource Resource   // for EffResource
	Amount   int
}

func need(scope Scope, n Need, amount int) Effect {
	return Effect{Kind: EffNeed, Scope: scope, Need: n, Amount: amount}
}

func needIn(d DistrictID, n Need, amount int) Effect {
	return Effect{Kind: EffNeed, Scope: ScopeFixed, District: d, Need: n, Amount: amount}
}

func buildings(scope Scope, amount int) Effect {
	return Effect{Kind: EffBuilding, Scope: scope, Amount: amount}
}

func buildingsIn(d DistrictID, amount int) Effect {
	return Effect{Kind: EffBuilding, Scope: ScopeFixed, District: d, Amount: amount}
}

func gain(r Resource, amount int) Effect {
	return Effect{Kind: EffResource, Scope: ScopeCity, Resource: r, Amount: amount}
}

// DecreeID names a decree.
type DecreeID int

// The decrees. The warden may issue one per day.
const (
	RationGrain DecreeID = iota
	Curfew
	Corvee
	ShoreUp
	Relief
	NumDecrees
)

// Decree is a one-per-day order with a cost.
type Decree struct {
	ID       DecreeID
	Slug     string
	Name     string
	Text     string
	Cost     Cost
	Targeted bool // needs a target district
	Effects  []Effect
}

// Decrees is indexed by DecreeID. Costs and amounts are tuning: see tuning.go.
var Decrees = [NumDecrees]Decree{
	RationGrain: {
		ID: RationGrain, Slug: "ration-grain", Name: "Ration Grain",
		Text:    "Open the stores. Every district eats today.",
		Cost:    T.DecreeCost[RationGrain],
		Effects: []Effect{need(ScopeAll, Supply, T.DecreeCityNeed)},
	},
	Curfew: {
		ID: Curfew, Slug: "curfew", Name: "Curfew",
		Text:    "Pay the watch double and clear the streets at dusk.",
		Cost:    T.DecreeCost[Curfew],
		Effects: []Effect{need(ScopeAll, Order, T.DecreeCityNeed)},
	},
	Corvee: {
		ID: Corvee, Slug: "corvee", Name: "Corvée",
		Text:    "Every household sends a pair of hands to mend what is failing.",
		Cost:    T.DecreeCost[Corvee],
		Effects: []Effect{need(ScopeAll, Upkeep, T.DecreeCityNeed)},
	},
	ShoreUp: {
		ID: ShoreUp, Slug: "shore-up", Name: "Shore Up",
		Text:     "Timber and nails for every building in one district.",
		Cost:     T.DecreeCost[ShoreUp],
		Targeted: true,
		Effects:  []Effect{buildings(ScopeTarget, T.DecreeShoreUp)},
	},
	Relief: {
		ID: Relief, Slug: "relief", Name: "Relief",
		Text:     "Bread, wardens and work gangs for one district.",
		Cost:     T.DecreeCost[Relief],
		Targeted: true,
		Effects:  []Effect{need(ScopeTarget, NoNeed, T.DecreeRelief)},
	},
}
