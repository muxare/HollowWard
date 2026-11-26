package sim

import "hollowward/internal/content"

// Phase is where in the day-and-night loop the tenure stands.
type Phase string

// The phases.
const (
	PhaseNone  Phase = ""      // no tenure yet
	PhaseDay   Phase = "day"   // the warden runs the city
	PhaseDusk  Phase = "dusk"  // a district has turned (or not); choose who goes in
	PhaseNight Phase = "night" // a party is in a hollow
	PhaseOver  Phase = "over"  // the tenure has ended
)

// Outcome is how a tenure ended.
type Outcome string

// The outcomes.
const (
	Ongoing     Outcome = ""
	Won         Outcome = "won"         // survived every night
	LostLid     Outcome = "lost-lid"    // the Dreamer reached its limit
	LostNoParty Outcome = "lost-nobody" // no citizen is left to fight
)

// State is a whole tenure. It is plain data and marshals to JSON. Nothing
// changes it but Apply.
type State struct {
	Version int     `json:"version"`
	Seed    uint64  `json:"seed"`
	Day     int     `json:"day"` // 1-based; night N follows day N
	Phase   Phase   `json:"phase"`
	Outcome Outcome `json:"outcome"`

	Resources content.Cost                   `json:"resources"`
	Districts [content.NumDistricts]District `json:"districts"`
	Citizens  []Citizen                      `json:"citizens"`
	Stewards  [content.NumStewards]Post      `json:"stewards"`

	DecreeIssued bool       `json:"decree_issued"` // one decree per day
	Card         *DrawnCard `json:"card"`          // today's event
	Deck         []int      `json:"deck"`          // cards not yet drawn, top first

	RNG Streams `json:"rng"`
}

// SaveVersion is the save format this code writes.
const SaveVersion = 1

// District is the changing part of a district.
type District struct {
	ID        content.DistrictID    `json:"id"`
	Needs     [content.NumNeeds]int `json:"needs"`
	Buildings []Building            `json:"buildings"` // in the order of content.Districts[ID].Buildings
	Figure    string                `json:"figure"`    // the key figure's name this tenure
	Turned    bool                  `json:"turned"`
}

// Building is the changing part of a building.
type Building struct {
	Condition int `json:"condition"`
}

// Grievance is the district's lowest need: what its key figure resents. Ties
// go to the earlier need.
func (d *District) Grievance() content.Need {
	low := content.Supply
	for n := content.Need(1); n < content.NumNeeds; n++ {
		if d.Needs[n] < d.Needs[low] {
			low = n
		}
	}
	return low
}

// WorstBuilding is the index of the building in the worst condition. Ties go
// to the earlier building.
func (d *District) WorstBuilding() int {
	low := 0
	for i := range d.Buildings {
		if d.Buildings[i].Condition < d.Buildings[low].Condition {
			low = i
		}
	}
	return low
}

// Citizen is one of the city's people. The same people work by day and fight
// by night.
type Citizen struct {
	ID     int                `json:"id"` // index into State.Citizens
	Name   string             `json:"name"`
	Job    content.JobID      `json:"job"`
	Home   content.DistrictID `json:"home"` // where they live and work
	Traits [2]content.TraitID `json:"traits"`
	Scars  []content.ScarID   `json:"scars"`

	// Jitter is this citizen's personal variation on their job's stats.
	JitterHP    int `json:"jitter_hp"`
	JitterSpeed int `json:"jitter_speed"`

	HP        int  `json:"hp"`
	Alive     bool `json:"alive"`
	Downed    bool `json:"downed"`
	Exhausted bool `json:"exhausted"` // fought last night; works less today
}

// Stats are the citizen's combat numbers: job, plus jitter, traits and scars.
func (c *Citizen) Stats() content.Stats {
	st := content.Jobs[c.Job].Stats()
	st.HP += c.JitterHP
	st.Speed += c.JitterSpeed
	for _, t := range c.Traits {
		m := content.Traits[t].Mods()
		st.HP, st.Attack, st.Guard, st.Speed = st.HP+m.HP, st.Attack+m.Attack, st.Guard+m.Guard, st.Speed+m.Speed
	}
	for _, sc := range c.Scars {
		m := content.Scars[sc].Mods()
		st.HP, st.Attack, st.Guard, st.Speed = st.HP+m.HP, st.Attack+m.Attack, st.Guard+m.Guard, st.Speed+m.Speed
	}
	st.HP, st.Attack, st.Guard, st.Speed = max(st.HP, 1), max(st.Attack, 1), max(st.Guard, 0), max(st.Speed, 1)
	return st
}

// HasTrait reports whether the citizen has a trait for which pick is true.
func (c *Citizen) HasTrait(pick func(content.TraitDef) bool) bool {
	for _, t := range c.Traits {
		if pick(content.Traits[t]) {
			return true
		}
	}
	return false
}

// WorkPct is the citizen's output today, in percent of a full day.
func (c *Citizen) WorkPct() int {
	if !c.Alive {
		return 0
	}
	pct := 100
	for _, t := range c.Traits {
		pct += content.Traits[t].Mods().WorkPct
	}
	for _, sc := range c.Scars {
		pct += content.Scars[sc].Mods().WorkPct
	}
	if c.Exhausted && !c.HasTrait(func(t content.TraitDef) bool { return t.Tireless }) {
		pct = pct * content.T.ExhaustedWorkPct / 100
	}
	return max(pct, 0)
}

// Post is where a steward is sent today.
type Post struct {
	District content.DistrictID `json:"district"` // NoDistrict: idle
	Need     content.Need       `json:"need"`     // NoNeed: the lowest need
	Building int                `json:"building"` // -1: the worst building
}

// DrawnCard is today's event card.
type DrawnCard struct {
	Card     int                `json:"card"`     // index into content.EventCards
	District content.DistrictID `json:"district"` // the card's target
	Resolved bool               `json:"resolved"`
	Choice   int                `json:"choice"`
}

// Clone returns a deep copy. Apply works on a clone, so the state passed in
// is never changed.
func (s State) Clone() State {
	n := s
	for i := range n.Districts {
		n.Districts[i].Buildings = append([]Building(nil), s.Districts[i].Buildings...)
	}
	n.Citizens = make([]Citizen, len(s.Citizens))
	for i, c := range s.Citizens {
		c.Scars = append([]content.ScarID(nil), c.Scars...)
		n.Citizens[i] = c
	}
	if s.Card != nil {
		card := *s.Card
		n.Card = &card
	}
	n.Deck = append([]int(nil), s.Deck...)
	n.RNG = s.RNG.clone()
	return n
}

// Living counts the citizens still alive.
func (s *State) Living() int {
	n := 0
	for i := range s.Citizens {
		if s.Citizens[i].Alive {
			n++
		}
	}
	return n
}

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }
