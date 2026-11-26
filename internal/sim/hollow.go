package sim

import (
	"fmt"

	"hollowward/internal/content"
)

// RoomKind says what a room of a hollow is.
type RoomKind string

// The room kinds.
const (
	RoomEntry  RoomKind = "entry"  // the way in; a fight room like any other
	RoomFight  RoomKind = "fight"  // holds whatever the district's needs bred
	RoomHazard RoomKind = "hazard" // a ruined building; hurts everyone who enters
	RoomHaven  RoomKind = "haven"  // a well-kept building; rest, healing, no enemies
	RoomHeart  RoomKind = "heart"  // holds the boss
)

// Hollow is a district remade as a dungeon. It is generated once, at the dusk
// the district turns, from that district's state and nothing else.
type Hollow struct {
	District  content.DistrictID `json:"district"`
	Night     int                `json:"night"`     // the night it turned
	Neglect   int                `json:"neglect"`   // the district's Neglect that dusk
	Grievance content.Need       `json:"grievance"` // the lowest need that dusk
	Rooms     []Room             `json:"rooms"`
	Entry     int                `json:"entry"`
	Heart     int                `json:"heart"`
}

// Room is one node of a hollow's graph. Exits lead deeper; every path ends at
// the heart.
type Room struct {
	ID       int      `json:"id"`
	Kind     RoomKind `json:"kind"`
	Name     string   `json:"name"`
	Depth    int      `json:"depth"`
	Exits    []int    `json:"exits"`
	Building int      `json:"building"` // for hazards and havens: which building; else -1
	Hazard   int      `json:"hazard"`   // for hazards: HP lost by each citizen who enters
	Enemies  []Enemy  `json:"enemies"`
	Cause    Cause    `json:"cause"` // why the room exists
	Visited  bool     `json:"visited"`
	Cleared  bool     `json:"cleared"`
}

// Enemy is one thing that lives in a hollow.
type Enemy struct {
	Kind      content.EnemyID `json:"kind"`
	Boss      bool            `json:"boss"`
	Name      string          `json:"name"`
	HP        int             `json:"hp"`
	MaxHP     int             `json:"max_hp"`
	Attack    int             `json:"attack"`
	Guard     int             `json:"guard"`
	Speed     int             `json:"speed"`
	Cause     Cause           `json:"cause"`     // why it exists
	NextMove  int             `json:"next_move"` // position in its repertoire
	Intent    Intent          `json:"intent"`    // what it does this round
	Next      Intent          `json:"next"`      // what it does the round after
	Acted     bool            `json:"acted"`     // has taken this round's action
	Stunned   bool            `json:"stunned"`   // loses its next action
	Hooked    bool            `json:"hooked"`    // acts last next round
	AttackMod int             `json:"attack_mod"`
}

// Upcoming is the enemy's next action as the player should see it: this
// round's if it has yet to act, otherwise next round's.
func (e *Enemy) Upcoming() Intent {
	if e.Acted {
		return e.Next
	}
	return e.Intent
}

// Intent is an action an enemy has shown it will take. Enemies always show
// two: this round's and the next round's. So whenever a citizen has to choose,
// every enemy action that can happen before their next choice is on show.
type Intent struct {
	Move   content.MoveID `json:"move"`
	Target int            `json:"target"` // a citizen ID, or -1 for everyone or no one
}

// Slug identifies the enemy's kind for the Ledger.
func (e *Enemy) Slug(h *Hollow) string {
	if e.Boss {
		return content.BossSlug(h.District, h.Grievance)
	}
	return content.Enemies[e.Kind].Slug
}

func (h Hollow) clone() Hollow {
	n := h
	n.Rooms = make([]Room, len(h.Rooms))
	for i, r := range h.Rooms {
		r.Exits = append([]int(nil), r.Exits...)
		r.Enemies = append([]Enemy(nil), r.Enemies...)
		n.Rooms[i] = r
	}
	return n
}

// Group is a run of like enemies in a room, for display.
type Group struct {
	Name  string
	Count int
	Boss  bool
	Cause Cause
}

// String renders the group the way the UI and the report show it:
// "Gaunts ×4: Market Supply was 18".
func (g Group) String() string {
	if g.Count == 1 {
		return fmt.Sprintf("%s: %s", g.Name, g.Cause.Describe())
	}
	return fmt.Sprintf("%s ×%d: %s", g.Name, g.Count, g.Cause.Describe())
}

// Groups gathers a room's enemies by kind, in order of first appearance.
func (r *Room) Groups() []Group {
	var out []Group
	index := map[string]int{}
	for i := range r.Enemies {
		e := &r.Enemies[i]
		key := e.Name
		if at, ok := index[key]; ok {
			out[at].Count++
			continue
		}
		index[key] = len(out)
		out = append(out, Group{Name: e.Name, Count: 1, Boss: e.Boss, Cause: e.Cause})
	}
	for i := range out {
		if out[i].Count > 1 && !out[i].Boss {
			for id := range content.Enemies {
				if content.Enemies[id].Name == out[i].Name {
					out[i].Name = content.Enemies[id].Plural
				}
			}
		}
	}
	return out
}

// Alive counts the room's enemies still standing.
func (r *Room) Alive() int {
	n := 0
	for i := range r.Enemies {
		if r.Enemies[i].HP > 0 {
			n++
		}
	}
	return n
}
