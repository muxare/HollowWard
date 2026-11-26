package sim

import (
	"fmt"

	"hollowward/internal/content"
)

// CauseSource is the kind of city state a Cause points at.
type CauseSource string

// The four sources of everything in a hollow.
const (
	FromNeed     CauseSource = "need"
	FromBuilding CauseSource = "building"
	FromFigure   CauseSource = "figure"
	FromDistrict CauseSource = "district"
)

// Cause says why a thing exists: which piece of city state produced it and
// what that state's value was at the time. Every generated room, enemy and
// boss carries one, and the Neglect meter lists them.
//
// ID is "<district>/<thing>": "market/supply", "market/granary",
// "market/figure", or just "market" for the district itself.
type Cause struct {
	Source            CauseSource `json:"source"`
	ID                string      `json:"id"`
	ValueAtGeneration int         `json:"value"`
}

func needCause(d *District, n content.Need) Cause {
	return Cause{FromNeed, d.ID.Slug() + "/" + n.Slug(), d.Needs[n]}
}

func buildingCause(d *District, b int) Cause {
	return Cause{FromBuilding, d.ID.Slug() + "/" + content.Districts[d.ID].Buildings[b].Slug, d.Buildings[b].Condition}
}

// Resolve looks a cause up in a district's state. It returns the value that
// state holds now and whether the cause points at something real there.
func (c Cause) Resolve(d *District) (value int, ok bool) {
	slug := d.ID.Slug()
	switch c.Source {
	case FromDistrict:
		return int(d.ID), c.ID == slug
	case FromFigure:
		return d.Needs[d.Grievance()], c.ID == slug+"/figure"
	case FromNeed:
		for n := content.Need(0); n < content.NumNeeds; n++ {
			if c.ID == slug+"/"+n.Slug() {
				return d.Needs[n], true
			}
		}
	case FromBuilding:
		for i, def := range content.Districts[d.ID].Buildings {
			if c.ID == slug+"/"+def.Slug && i < len(d.Buildings) {
				return d.Buildings[i].Condition, true
			}
		}
	}
	return 0, false
}

// Describe renders the cause for a player: "Market Supply was 18".
func (c Cause) Describe() string {
	d, thing := splitCauseID(c.ID)
	if !d.Valid() {
		return c.ID
	}
	def := content.Districts[d]
	switch c.Source {
	case FromDistrict:
		return "it is the " + def.Name
	case FromFigure:
		return fmt.Sprintf("the %s's grievance, at %d", def.FigureTitle, c.ValueAtGeneration)
	case FromNeed:
		for n := content.Need(0); n < content.NumNeeds; n++ {
			if n.Slug() == thing {
				return fmt.Sprintf("%s %s was %d", def.Name, n, c.ValueAtGeneration)
			}
		}
	case FromBuilding:
		for _, b := range def.Buildings {
			if b.Slug == thing {
				return fmt.Sprintf("the %s was at %d", b.Name, c.ValueAtGeneration)
			}
		}
	}
	return c.ID
}

func splitCauseID(id string) (content.DistrictID, string) {
	slug, thing := id, ""
	for i := 0; i < len(id); i++ {
		if id[i] == '/' {
			slug, thing = id[:i], id[i+1:]
			break
		}
	}
	for d := content.DistrictID(0); d < content.NumDistricts; d++ {
		if d.Slug() == slug {
			return d, thing
		}
	}
	return content.NoDistrict, thing
}
