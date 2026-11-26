package sim

import "hollowward/internal/content"

// EventKind names something that happened.
type EventKind string

// What can happen by day. Night events are in night.go.
const (
	EvTenureStarted EventKind = "tenure-started" // N: seed's low bits are not logged; ID: ""
	EvDayStarted    EventKind = "day-started"
	EvCardDrawn     EventKind = "card-drawn"     // ID: card slug; District: target
	EvCardResolved  EventKind = "card-resolved"  // ID: card slug; N: choice
	EvDecreeIssued  EventKind = "decree-issued"  // ID: decree slug; District: target or none
	EvStewardWorked EventKind = "steward-worked" // District; ID: need slug; N: need gained; V: steward
	EvRepaired      EventKind = "repaired"       // District; ID: building slug; N: condition gained
	EvProduced      EventKind = "produced"       // District; ID: resource; N: amount
	EvRuined        EventKind = "ruined"         // District; ID: building slug; N: condition now
	EvDayEnded      EventKind = "day-ended"
)

// Event is one entry in the event log. The log is the single source for the
// chronicle, the after-action report and the Ledger.
type Event struct {
	Kind     EventKind          `json:"kind"`
	Day      int                `json:"day"`
	District content.DistrictID `json:"district"`
	Citizen  int                `json:"citizen"`
	Who      string             `json:"who,omitempty"` // the citizen's name, so the log reads alone
	ID       string             `json:"id,omitempty"`
	N        int                `json:"n,omitempty"`
	V        int                `json:"v,omitempty"`
	Causes   []Cause            `json:"causes,omitempty"`
}

// log collects the events of one Apply.
type log struct {
	day    int
	events []Event
}

func (l *log) add(kind EventKind) *Event {
	l.events = append(l.events, Event{Kind: kind, Day: l.day, District: content.NoDistrict, Citizen: -1})
	return &l.events[len(l.events)-1]
}

func (e *Event) in(d content.DistrictID) *Event { e.District = d; return e }
func (e *Event) who(c *Citizen) *Event          { e.Citizen, e.Who = c.ID, c.Name; return e }
func (e *Event) because(c ...Cause) *Event      { e.Causes = append(e.Causes, c...); return e }
func (e *Event) id(id string) *Event            { e.ID = id; return e }
func (e *Event) n(n int) *Event                 { e.N = n; return e }
func (e *Event) v(v int) *Event                 { e.V = v; return e }
