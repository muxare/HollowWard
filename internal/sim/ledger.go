package sim

import (
	"fmt"
	"sort"

	"hollowward/internal/content"
)

// LedgerKind is what a Ledger entry records.
type LedgerKind string

// The kinds.
const (
	LedgerEnemy  LedgerKind = "enemy"
	LedgerHazard LedgerKind = "hazard"
	LedgerBoss   LedgerKind = "boss"
)

// LedgerEntry is one thing a warden has met, with the cause that produced it
// the first time.
type LedgerEntry struct {
	Kind     LedgerKind         `json:"kind"`
	ID       string             `json:"id"`       // enemy slug, boss slug, or "<district>/<building>"
	District content.DistrictID `json:"district"` // where it was first met
	Cause    Cause              `json:"cause"`    // why it existed, the first time
	Seen     int                `json:"seen"`     // how many times it has been met
	Tenure   int                `json:"tenure"`   // the tenure it was first met in, from 1
}

// Ledger is the record that passes from warden to warden: which neglect
// breeds which horror. It is knowledge only. Nothing in the rules reads it:
// Apply does not take a Ledger, so it cannot add a stat.
type Ledger struct {
	Version   int           `json:"version"`
	Tenures   int           `json:"tenures"`   // tenures ended
	Won       int           `json:"won"`       // of which survived every night
	Entries   []LedgerEntry `json:"entries"`   // in the order first met
	Fragments []int         `json:"fragments"` // the old warden's notes found, sorted
}

// Knows reports whether the Ledger records a thing.
func (l *Ledger) Knows(kind LedgerKind, id string) bool {
	for i := range l.Entries {
		if l.Entries[i].Kind == kind && l.Entries[i].ID == id {
			return true
		}
	}
	return false
}

// KnowsFamily reports whether any enemy of a family is on record.
func (l *Ledger) KnowsFamily(f content.Family) bool {
	for _, e := range content.Enemies {
		if e.Family == f && l.Knows(LedgerEnemy, e.Slug) {
			return true
		}
	}
	return false
}

// HasAllFragments reports whether every one of the old warden's notes has
// been found, which unlocks the closing text.
func (l *Ledger) HasAllFragments() bool { return len(l.Fragments) >= len(content.Fragments) }

// Record returns the Ledger with a run of events written into it. It reads
// the event log and nothing else.
func (l Ledger) Record(events []Event) Ledger {
	n := l
	n.Version = SaveVersion
	n.Entries = append([]LedgerEntry(nil), l.Entries...)
	n.Fragments = append([]int(nil), l.Fragments...)
	meet := func(kind LedgerKind, id string, e Event) {
		for i := range n.Entries {
			if n.Entries[i].Kind == kind && n.Entries[i].ID == id {
				n.Entries[i].Seen++
				return
			}
		}
		entry := LedgerEntry{Kind: kind, ID: id, District: e.District, Seen: 1, Tenure: n.Tenures + 1}
		if len(e.Causes) > 0 {
			entry.Cause = e.Causes[0]
		}
		n.Entries = append(n.Entries, entry)
	}
	for _, e := range events {
		switch e.Kind {
		case EvMet:
			if e.V == 1 {
				meet(LedgerBoss, e.ID, e)
			} else {
				meet(LedgerEnemy, e.ID, e)
			}
		case EvHazard:
			meet(LedgerHazard, e.District.Slug()+"/"+e.ID, e)
		case EvFragment:
			found := false
			for _, f := range n.Fragments {
				found = found || f == e.N
			}
			if !found {
				n.Fragments = append(n.Fragments, e.N)
				sort.Ints(n.Fragments)
			}
		case EvTenureEnded:
			n.Tenures++
			if Outcome(e.ID) == Won {
				n.Won++
			}
		}
	}
	return n
}

// PreviewLine is one line of the dusk preview: a thing the hollow holds, named
// only if the Ledger already records it. The cause is always shown, because
// the cause is the city's own state, which the warden can see.
type PreviewLine struct {
	Known bool   `json:"known"`
	Text  string `json:"text"`
	Cause Cause  `json:"cause"`
}

// String renders the line with its cause.
func (p PreviewLine) String() string { return p.Text + ": " + p.Cause.Describe() }

// Preview says what a hollow holds, as far as the Ledger can tell. Things the
// Ledger does not record yet appear as unknowns, with their cause.
func Preview(l *Ledger, h *Hollow) []PreviewLine {
	var out []PreviewLine
	type tally struct {
		count int
		cause Cause
	}
	// Enemies, by kind, in a fixed order.
	kinds := map[content.EnemyID]*tally{}
	var boss *Enemy
	for i := range h.Rooms {
		for j := range h.Rooms[i].Enemies {
			e := &h.Rooms[i].Enemies[j]
			if e.HP <= 0 {
				continue
			}
			if e.Boss {
				boss = e
				continue
			}
			if kinds[e.Kind] == nil {
				kinds[e.Kind] = &tally{cause: e.Cause}
			}
			kinds[e.Kind].count++
		}
	}
	unknown := map[content.Family]*tally{}
	for id := content.EnemyID(0); id < content.NumEnemies; id++ {
		tl := kinds[id]
		if tl == nil {
			continue
		}
		def := content.Enemies[id]
		if l.Knows(LedgerEnemy, def.Slug) {
			name := def.Name
			if tl.count > 1 {
				name = fmt.Sprintf("%s ×%d", def.Plural, tl.count)
			}
			out = append(out, PreviewLine{true, name, tl.cause})
			continue
		}
		if unknown[def.Family] == nil {
			unknown[def.Family] = &tally{cause: tl.cause}
		}
		unknown[def.Family].count += tl.count
	}
	for f := content.Family(0); f < content.NumFamilies; f++ {
		if tl := unknown[f]; tl != nil {
			text := fmt.Sprintf("Something not in the Ledger ×%d", tl.count)
			if l.KnowsFamily(f) {
				text = fmt.Sprintf("More of the %s kind, not yet in the Ledger ×%d", f, tl.count)
			}
			out = append(out, PreviewLine{false, text, tl.cause})
		}
	}
	// Buildings.
	for i := range h.Rooms {
		r := &h.Rooms[i]
		switch r.Kind {
		case RoomHazard:
			if r.Cleared {
				continue
			}
			id := h.District.Slug() + "/" + content.Districts[h.District].Buildings[r.Building].Slug
			if l.Knows(LedgerHazard, id) {
				out = append(out, PreviewLine{true, fmt.Sprintf("Hazard: %s (-%d HP each)", r.Name, r.Hazard), r.Cause})
			} else {
				out = append(out, PreviewLine{false, "A hazard not in the Ledger", r.Cause})
			}
		case RoomHaven:
			out = append(out, PreviewLine{true, "Safe room: " + r.Name, r.Cause})
		}
	}
	// The boss.
	if boss != nil {
		if l.Knows(LedgerBoss, content.BossSlug(h.District, h.Grievance)) {
			out = append(out, PreviewLine{true, "At the heart: " + boss.Name, boss.Cause})
		} else {
			out = append(out, PreviewLine{false, "At the heart: someone not in the Ledger", boss.Cause})
		}
	}
	return out
}
