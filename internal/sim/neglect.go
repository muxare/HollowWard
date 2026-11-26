package sim

import (
	"sort"

	"hollowward/internal/content"
)

// NeglectCause is one reason a district is neglected.
type NeglectCause struct {
	Cause
	// Thirds is this cause's share of the Neglect score, in thirds of a point,
	// so three needs can share "100 minus the mean" without rounding.
	Thirds int `json:"thirds"`
}

// NeglectReport is a district's Neglect and why.
type NeglectReport struct {
	Neglect int            `json:"neglect"`
	Causes  []NeglectCause `json:"causes"` // largest first
}

// Neglect scores a district: 100 minus the mean of its three needs, plus a
// fixed amount for each ruined building. It is not capped. The causes come
// back largest first; causes that add nothing are left out.
func Neglect(d *District) NeglectReport {
	t := content.T
	var rep NeglectReport
	thirds := 0
	for n := content.Need(0); n < content.NumNeeds; n++ {
		lack := t.NeedMax - d.Needs[n]
		thirds += lack
		if lack > 0 {
			rep.Causes = append(rep.Causes, NeglectCause{needCause(d, n), lack})
		}
	}
	for i := range d.Buildings {
		if d.Buildings[i].Condition < t.RuinThreshold {
			thirds += 3 * t.NeglectPerRuin
			rep.Causes = append(rep.Causes, NeglectCause{buildingCause(d, i), 3 * t.NeglectPerRuin})
		}
	}
	rep.Neglect = (thirds + 1) / 3 // to the nearest point
	sort.SliceStable(rep.Causes, func(i, j int) bool { return rep.Causes[i].Thirds > rep.Causes[j].Thirds })
	return rep
}
