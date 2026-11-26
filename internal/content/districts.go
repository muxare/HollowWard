package content

// BuildingDef is a named building in a district. A ruined one becomes a hazard
// room in the hollow; a well-kept one becomes a safe room.
type BuildingDef struct {
	Slug   string
	Name   string
	Hazard string // the hazard room it becomes below the ruin threshold
	Haven  string // the safe room it becomes at or above the haven threshold
}

// DistrictDef is the fixed part of a district.
type DistrictDef struct {
	ID          DistrictID
	Slug        string
	Name        string
	Plural      bool   // "the Docks are", not "the Docks is"
	FigureTitle string // the key figure's office; their name is reshuffled each tenure
	Buildings   []BuildingDef
	Jobs        []JobID // one resident per entry
}

// Districts is indexed by DistrictID.
var Districts = [NumDistricts]DistrictDef{
	Market: {
		ID: Market, Slug: "market", Name: "Market", FigureTitle: "Grain Factor",
		Buildings: []BuildingDef{
			{"granary", "Granary", "The Empty Granary", "The Full Granary"},
			{"counting-house", "Counting House", "The Counting House, Unbalanced", "The Counting House"},
			{"stalls", "Stalls", "The Rotten Stalls", "The Covered Stalls"},
		},
		Jobs: []JobID{Merchant, Merchant, Smith},
	},
	Watch: {
		ID: Watch, Slug: "watch", Name: "Watch", FigureTitle: "Captain",
		Buildings: []BuildingDef{
			{"barracks", "Barracks", "The Barracks, Bunks Stripped", "The Barracks"},
			{"gatehouse", "Gatehouse", "The Gate That Will Not Close", "The Gatehouse"},
		},
		Jobs: []JobID{Watchman, Watchman, Smith},
	},
	Docks: {
		ID: Docks, Slug: "docks", Name: "Docks", Plural: true, FigureTitle: "Harbourmaster",
		Buildings: []BuildingDef{
			{"warehouse", "Warehouse", "The Flooded Warehouse", "The Dry Warehouse"},
			{"crane", "Crane", "The Crane, Still Turning", "The Crane Yard"},
			{"customs-house", "Customs House", "The Customs House, Sealed", "The Customs House"},
		},
		Jobs: []JobID{Dockhand, Dockhand, Dockhand, Labourer},
	},
	Temple: {
		ID: Temple, Slug: "temple", Name: "Temple", FigureTitle: "Prior",
		Buildings: []BuildingDef{
			{"sanctum", "Sanctum", "The Sanctum, Candles Out", "The Lit Sanctum"},
			{"almshouse", "Almshouse", "The Almshouse, Beds Empty", "The Almshouse"},
		},
		Jobs: []JobID{Priest, Priest, Labourer},
	},
	Mills: {
		ID: Mills, Slug: "mills", Name: "Mills", Plural: true, FigureTitle: "Mill Reeve",
		Buildings: []BuildingDef{
			{"east-wheel", "East Wheel", "The East Wheel, Grinding Nothing", "The East Wheel"},
			{"grain-store", "Grain Store", "The Grain Store, Gone to Dust", "The Grain Store"},
			{"bakehouse", "Bakehouse", "The Cold Bakehouse", "The Warm Bakehouse"},
		},
		Jobs: []JobID{Miller, Miller, Miller, Smith},
	},
	Tenements: {
		ID: Tenements, Slug: "tenements", Name: "Tenements", Plural: true, FigureTitle: "Rentmaster",
		Buildings: []BuildingDef{
			{"well", "Well", "The Well, Rope Cut", "The Sweet Well"},
			{"long-stair", "Long Stair", "The Long Stair, Burned", "The Long Stair"},
		},
		Jobs: []JobID{Labourer, Labourer, Merchant},
	},
}

// FigureNames are the family names a key figure can carry. A new tenure
// shuffles them over the districts.
var FigureNames = []string{
	"Harrow", "Vane", "Marsh", "Coldwell", "Thrace", "Wick",
	"Ashby", "Morrow", "Pell", "Strand", "Gault", "Fenwick",
}

// StewardNames are the three stewards.
var StewardNames = [NumStewards]string{"Steward Ilse", "Steward Odo", "Steward Brann"}

// NumStewards is how many stewards the warden has. Fewer than districts, so
// some district always slips.
const NumStewards = 3

// Verb picks the verb form that agrees with the district's name: the Market
// turns, the Docks turn.
func (d DistrictID) Verb(singular, plural string) string {
	if d.Valid() && Districts[d].Plural {
		return plural
	}
	return singular
}
