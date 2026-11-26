package content

// Need is one of the three things every district needs.
type Need int

// The needs, in fixed order. NoNeed means "none" or "pick for me".
const (
	Supply Need = iota
	Order
	Upkeep
	NumNeeds

	NoNeed Need = -1
)

var needNames = [NumNeeds]string{"Supply", "Order", "Upkeep"}
var needSlugs = [NumNeeds]string{"supply", "order", "upkeep"}

func (n Need) Valid() bool { return n >= 0 && n < NumNeeds }

func (n Need) String() string {
	if !n.Valid() {
		return "none"
	}
	return needNames[n]
}

// Slug is the lower-case name used in cause and event IDs.
func (n Need) Slug() string {
	if !n.Valid() {
		return ""
	}
	return needSlugs[n]
}

// DistrictID names one of the six districts.
type DistrictID int

// The districts, in fixed order. NoDistrict means "none".
const (
	Market DistrictID = iota
	Watch
	Docks
	Temple
	Mills
	Tenements
	NumDistricts

	NoDistrict DistrictID = -1
)

func (d DistrictID) Valid() bool { return d >= 0 && d < NumDistricts }

func (d DistrictID) String() string {
	if !d.Valid() {
		return "nowhere"
	}
	return Districts[d].Name
}

// Slug is the lower-case name used in cause and event IDs.
func (d DistrictID) Slug() string {
	if !d.Valid() {
		return ""
	}
	return Districts[d].Slug
}

// Resource is one of the three city resources.
type Resource int

// The resources, in fixed order.
const (
	Coin Resource = iota
	Grain
	Timber
	NumResources
)

var resourceNames = [NumResources]string{"coin", "grain", "timber"}

func (r Resource) String() string {
	if r < 0 || r >= NumResources {
		return "nothing"
	}
	return resourceNames[r]
}

// Family is an enemy family. Each grows from one need and drops one material.
type Family int

// The families, in the order of the needs that breed them.
const (
	Hunger Family = iota // from Supply; drops hungerbone
	Riot                 // from Order; drops ash-masks
	Rot                  // from Upkeep; drops blackcap
	NumFamilies
)

var familyNames = [NumFamilies]string{"Hunger", "Riot", "Rot"}
var familySlugs = [NumFamilies]string{"hunger", "riot", "rot"}
var materialNames = [NumFamilies]string{"hungerbone", "ash-masks", "blackcap"}

func (f Family) Valid() bool { return f >= 0 && f < NumFamilies }

func (f Family) String() string {
	if !f.Valid() {
		return "none"
	}
	return familyNames[f]
}

func (f Family) Slug() string {
	if !f.Valid() {
		return ""
	}
	return familySlugs[f]
}

// Material is the name of what the family drops.
func (f Family) Material() string {
	if !f.Valid() {
		return ""
	}
	return materialNames[f]
}

// FamilyOf is the enemy family a low need breeds.
func FamilyOf(n Need) Family { return Family(n) }

// NeedOf is the need whose lack breeds the family.
func NeedOf(f Family) Need { return Need(f) }

// Cost is an amount of each resource.
type Cost [NumResources]int

// IsZero reports whether the cost is free.
func (c Cost) IsZero() bool { return c == Cost{} }
