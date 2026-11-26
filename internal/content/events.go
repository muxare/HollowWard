package content

// Choice is one answer to an event card.
type Choice struct {
	Label   string
	Result  string // one line shown after choosing
	Cost    Cost
	Effects []Effect
}

// EventCard is a daily event. Each has at least one choice that costs nothing,
// so a card can always be answered.
type EventCard struct {
	Slug     string
	Title    string
	Text     string     // "{district}" is replaced by the target's name
	District DistrictID // the target; NoDistrict means a seeded pick
	Choices  []Choice
}

// e is short for the event magnitudes in tuning.go.
var e = T.Event

// EventCards is the deck. One card is drawn per day without replacement; the
// deck is reshuffled when it runs out.
var EventCards = []EventCard{
	{
		Slug: "grain-barge", Title: "The Barge", District: Docks,
		Text: "A grain barge has run aground off the mole. The dockhands are already wading out.",
		Choices: []Choice{
			{Label: "Post the watch on the cargo", Result: "The grain reaches the stores. The docks remember who took it.",
				Effects: []Effect{gain(Grain, e.XXL), needIn(Docks, Order, -e.L)}},
			{Label: "Let the quarter take what it can carry", Result: "By noon the barge is a rib cage. The docks eat well.",
				Effects: []Effect{needIn(Docks, Supply, e.XXL)}},
		},
	},
	{
		Slug: "rats", Title: "Rats", District: NoDistrict,
		Text: "Rats in the walls of the {district}. More than last year. They are not afraid.",
		Choices: []Choice{
			{Label: "Pay the ratcatchers", Result: "They work three nights. They do not talk about the third.",
				Cost: Cost{Coin: e.CostMid}, Effects: []Effect{need(ScopeTarget, Upkeep, e.XL)}},
			{Label: "Let them be", Result: "The walls go on moving.",
				Effects: []Effect{need(ScopeTarget, Upkeep, -e.L), need(ScopeTarget, Supply, -e.S)}},
		},
	},
	{
		Slug: "tithe", Title: "The Early Tithe", District: Temple,
		Text: "The prior asks for the winter tithe early. He does not say why.",
		Choices: []Choice{
			{Label: "Pay it", Result: "The bells ring on time for a week.",
				Cost: Cost{Coin: e.CostHigh}, Effects: []Effect{needIn(Temple, NoNeed, e.M)}},
			{Label: "Refuse", Result: "The sermon that evening is about locked doors.",
				Effects: []Effect{needIn(Temple, Order, -e.XL)}},
		},
	},
	{
		Slug: "deserters", Title: "Empty Cots", District: Watch,
		Text: "Three of the watch did not report at dawn. Their cots are made.",
		Choices: []Choice{
			{Label: "Raise the pay", Result: "The rest stay. They count the coin twice.",
				Cost: Cost{Coin: e.CostHigh}, Effects: []Effect{needIn(Watch, Order, e.XXL)}},
			{Label: "Say nothing", Result: "The others noticed that you said nothing.",
				Effects: []Effect{needIn(Watch, Order, -e.L)}},
		},
	},
	{
		Slug: "grey-flour", Title: "Grey Flour", District: Mills,
		Text: "The flour from the east wheel is grey. It bakes. It rises. Nobody wants to be first.",
		Choices: []Choice{
			{Label: "Burn it", Result: "It burns slowly and smells of rain.",
				Effects: []Effect{gain(Grain, -e.L), needIn(Mills, Upkeep, e.M)}},
			{Label: "Sell it", Result: "It sells. People are hungry enough.",
				Effects: []Effect{gain(Coin, e.M), need(ScopeAll, Supply, -e.XS)}},
		},
	},
	{
		Slug: "stair-fire", Title: "The Stair", District: Tenements,
		Text: "A cookfire took a stair in the tenements. Out by morning. The smell stays.",
		Choices: []Choice{
			{Label: "Send timber", Result: "New treads by nightfall. They creak in a different voice.",
				Cost: Cost{Timber: e.CostMid}, Effects: []Effect{buildingsIn(Tenements, e.XXL)}},
			{Label: "Let them patch it", Result: "They patch it with doors.",
				Effects: []Effect{buildingsIn(Tenements, -e.XL)}},
		},
	},
	{
		Slug: "no-flag", Title: "No Flag", District: Docks,
		Text: "A captain with no flag offers timber at half price, if nobody counts the crates.",
		Choices: []Choice{
			{Label: "Buy", Result: "Good timber. Some of it is carved, on the inside faces.",
				Cost: Cost{Coin: e.CostMid}, Effects: []Effect{gain(Timber, e.XL), needIn(Docks, Order, -e.M)}},
			{Label: "Turn him away", Result: "He sails on the evening tide without a light.",
				Effects: []Effect{needIn(Docks, Order, e.S)}},
		},
	},
	{
		Slug: "bread-price", Title: "The Price of Bread", District: Market,
		Text: "The grain factor has raised the price of bread twice since the new moon.",
		Choices: []Choice{
			{Label: "Fix the price", Result: "The factor smiles and writes something down.",
				Cost: Cost{Coin: e.CostMid}, Effects: []Effect{needIn(Market, Order, e.L), needIn(Market, Supply, e.S)}},
			{Label: "Let the market find its level", Result: "It finds it. The level is low.",
				Effects: []Effect{gain(Coin, e.M), needIn(Market, Supply, -e.L)}},
		},
	},
	{
		Slug: "pilgrims", Title: "Pilgrims", District: Temple,
		Text: "Pilgrims at the gate. They say they dreamed of this city. All of them, the same night.",
		Choices: []Choice{
			{Label: "Let them in", Result: "They sleep in the tenement yards, facing the ground.",
				Effects: []Effect{gain(Coin, e.S6), needIn(Tenements, Supply, -e.M), needIn(Temple, Order, e.S)}},
			{Label: "Bar the gate", Result: "They kneel outside the wall until the frost.",
				Effects: []Effect{needIn(Watch, Order, e.S), needIn(Temple, Order, -e.M)}},
		},
	},
	{
		Slug: "cistern", Title: "The Cistern", District: NoDistrict,
		Text: "The cistern under the {district} has cracked. The water that comes up is warm.",
		Choices: []Choice{
			{Label: "Seal it", Result: "Sealed. Something knocks, twice, and stops.",
				Cost: Cost{Timber: e.CostLow}, Effects: []Effect{need(ScopeTarget, Upkeep, e.L)}},
			{Label: "Draw from it anyway", Result: "It tastes of nothing at all.",
				Effects: []Effect{need(ScopeTarget, Supply, e.M), need(ScopeTarget, Upkeep, -e.XL)}},
		},
	},
	{
		Slug: "quiet-street", Title: "The Quiet Street", District: NoDistrict,
		Text: "A street in the {district} has gone quiet. Doors shut. Shutters nailed from the inside.",
		Choices: []Choice{
			{Label: "Knock on every door, with bread", Result: "Some doors open. You do not ask about the others.",
				Cost: Cost{Grain: e.CostLow}, Effects: []Effect{need(ScopeTarget, Supply, e.S6), need(ScopeTarget, Order, e.M)}},
			{Label: "Leave them to it", Result: "The quiet spreads one street over.",
				Effects: []Effect{need(ScopeTarget, Order, -e.M)}},
		},
	},
	{
		Slug: "good-week", Title: "A Good Week", District: Mills,
		Text: "A good week at the wheels. The reeve asks what to do with the surplus.",
		Choices: []Choice{
			{Label: "Store it", Result: "The stores are full. For now that is all it means.",
				Effects: []Effect{gain(Grain, e.XL)}},
			{Label: "Share it out", Result: "For one evening the city sounds like a city.",
				Effects: []Effect{need(ScopeAll, Supply, e.S)}},
		},
	},
}
