package content

// UpgradeID names something the city can build from what the hollows drop.
type UpgradeID int

// The upgrades.
const (
	Ward        UpgradeID = iota // one building decays half as fast
	BoneCharms                   // every citizen is hardier
	AshHoods                     // every citizen takes less
	BoneBlades                   // every citizen hits harder
	GranarySeal                  // Supply decays more slowly everywhere
	CurfewBells                  // Order decays more slowly everywhere
	BlackMortar                  // Upkeep decays more slowly everywhere
	NumUpgrades
)

// Materials is an amount of each family's material.
type Materials [NumFamilies]int

// Upgrade is bought in the city with materials. Costs and sizes are in
// tuning.go. Several take a single material, so starving one district on
// purpose can be the right play.
type Upgrade struct {
	ID       UpgradeID
	Slug     string
	Name     string
	Text     string
	Building bool // bought for one building, and can be bought again for another
	Gear     Mods // added to every citizen
	Ease     Need // the need whose decay it slows, or NoNeed
}

// Cost is what the upgrade costs.
func (u Upgrade) Cost() Materials { return T.UpgradeCost[u.ID] }

// Upgrades is indexed by UpgradeID.
var Upgrades = [NumUpgrades]Upgrade{
	Ward: {ID: Ward, Slug: "ward", Name: "Blackcap Ward", Building: true, Ease: NoNeed,
		Text: "Painted on the lintel of one building. It decays half as fast."},
	BoneCharms: {ID: BoneCharms, Slug: "bone-charms", Name: "Bone Charms", Gear: T.GearCharms, Ease: NoNeed,
		Text: "One for every citizen. They are hardier for it."},
	AshHoods: {ID: AshHoods, Slug: "ash-hoods", Name: "Ash Hoods", Gear: T.GearHoods, Ease: NoNeed,
		Text: "Masks taken from the rioters, re-sewn. Blows land softer."},
	BoneBlades: {ID: BoneBlades, Slug: "bone-blades", Name: "Bone-and-Ash Blades", Gear: T.GearBlades, Ease: NoNeed,
		Text: "Edges that remember being hungry and angry."},
	GranarySeal: {ID: GranarySeal, Slug: "granary-seal", Name: "Granary Seals", Ease: Supply,
		Text: "Hungerbone ground into the door-wax. Supply drains more slowly, everywhere."},
	CurfewBells: {ID: CurfewBells, Slug: "curfew-bells", Name: "Curfew Bells", Ease: Order,
		Text: "Bells with ash in the bronze. Order drains more slowly, everywhere."},
	BlackMortar: {ID: BlackMortar, Slug: "black-mortar", Name: "Black Mortar", Ease: Upkeep,
		Text: "Blackcap in the lime. Upkeep drains more slowly, everywhere."},
}
