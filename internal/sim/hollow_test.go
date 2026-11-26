package sim

import (
	"strings"
	"testing"

	"hollowward/internal/content"
)

// district builds a district in a given state.
func district(id content.DistrictID, needs [3]int, conds ...int) District {
	d := District{ID: id, Needs: needs, Figure: "Harrow"}
	for i := range content.Districts[id].Buildings {
		c := 50
		if i < len(conds) {
			c = conds[i]
		}
		d.Buildings = append(d.Buildings, Building{Condition: c})
	}
	return d
}

// checkHollow holds a generated hollow to every promise the generator makes.
// It returns a description of the first broken promise, or "".
func checkHollow(d *District, h *Hollow) string {
	t := content.T
	fail := func(format string, args ...any) string {
		return strings.TrimSpace(strings.Join([]string{d.ID.String(), sprintf(format, args...)}, ": "))
	}
	if n := len(h.Rooms); n < 6 || n > 10 {
		return fail("%d rooms, want 6 to 10", n)
	}
	if h.District != d.ID || h.Grievance != d.Grievance() || h.Neglect != Neglect(d).Neglect {
		return fail("header does not match the district")
	}

	// The graph: one entry, one heart, every edge leads deeper, every room is
	// reachable and every path ends at the heart.
	hearts, entries := 0, 0
	indeg := make([]int, len(h.Rooms))
	for i := range h.Rooms {
		r := &h.Rooms[i]
		if r.ID != i {
			return fail("room %d has ID %d", i, r.ID)
		}
		switch r.Kind {
		case RoomHeart:
			hearts++
			if i != h.Heart || len(r.Exits) != 0 {
				return fail("heart misplaced or has exits")
			}
		case RoomEntry:
			entries++
			if i != h.Entry || r.Depth != 0 {
				return fail("entry misplaced")
			}
		}
		if r.Kind != RoomHeart && len(r.Exits) == 0 {
			return fail("room %q is a dead end", r.Name)
		}
		for _, x := range r.Exits {
			if x < 0 || x >= len(h.Rooms) {
				return fail("room %q has an exit to nowhere", r.Name)
			}
			if h.Rooms[x].Depth <= r.Depth {
				return fail("edge %q -> %q does not lead deeper", r.Name, h.Rooms[x].Name)
			}
			indeg[x]++
		}
	}
	if hearts != 1 || entries != 1 {
		return fail("%d hearts and %d entries", hearts, entries)
	}
	for i, n := range indeg {
		if n == 0 && i != h.Entry {
			return fail("room %q cannot be reached", h.Rooms[i].Name)
		}
	}

	// Causes: every room, enemy and boss points at real state with the
	// recorded value, and the cause is of the right kind.
	real := func(c Cause) bool {
		v, ok := c.Resolve(d)
		return ok && v == c.ValueAtGeneration
	}
	perFamily := [content.NumFamilies]int{}
	special := map[int]RoomKind{}
	bosses := 0
	for i := range h.Rooms {
		r := &h.Rooms[i]
		if !real(r.Cause) {
			return fail("room %q: cause %+v does not point at real state", r.Name, r.Cause)
		}
		switch r.Kind {
		case RoomEntry, RoomFight, RoomHeart:
			if r.Cause.Source != FromDistrict || r.Building != -1 || r.Hazard != 0 {
				return fail("room %q: a layout room must be caused by the district", r.Name)
			}
		case RoomHazard:
			if r.Cause.Source != FromBuilding || r.Cause.ValueAtGeneration >= t.RuinThreshold || r.Hazard <= 0 || len(r.Enemies) != 0 {
				return fail("hazard %q: bad cause or contents", r.Name)
			}
			special[r.Building] = r.Kind
		case RoomHaven:
			if r.Cause.Source != FromBuilding || r.Cause.ValueAtGeneration < t.HavenThreshold || len(r.Enemies) != 0 {
				return fail("haven %q: bad cause or contents", r.Name)
			}
			special[r.Building] = r.Kind
		default:
			return fail("room %q: unknown kind %q", r.Name, r.Kind)
		}
		limit := t.MaxEnemiesPerRoom
		if r.Kind == RoomHeart {
			limit = 1 + t.BossEscortMax
		}
		if len(r.Enemies) > limit {
			return fail("room %q holds %d enemies", r.Name, len(r.Enemies))
		}
		for j := range r.Enemies {
			e := &r.Enemies[j]
			if !real(e.Cause) {
				return fail("%s in %q: cause %+v does not point at real state", e.Name, r.Name, e.Cause)
			}
			if e.HP <= 0 || e.HP != e.MaxHP || e.Attack <= 0 || e.Speed <= 0 || e.Guard < 0 || e.Name == "" {
				return fail("%s: bad stats %+v", e.Name, *e)
			}
			if e.Boss {
				bosses++
				if r.Kind != RoomHeart || e.Cause.Source != FromFigure || e.Name != content.BossName(d.ID, d.Grievance()) {
					return fail("boss %q misplaced or miscaused", e.Name)
				}
				continue
			}
			def := content.Enemies[e.Kind]
			need := content.NeedOf(def.Family)
			if e.Cause.Source != FromNeed || e.Cause != needCause(d, need) {
				return fail("%s: caused by %+v, want the %s need", e.Name, e.Cause, need)
			}
			if d.Needs[need] >= t.EnemyNeedBelow {
				return fail("%s exists but %s is %d", e.Name, need, d.Needs[need])
			}
			if r.Kind != RoomHeart {
				perFamily[def.Family]++
			} else if need != d.Grievance() {
				return fail("escort %s is not of the grievance's family", e.Name)
			}
		}
	}
	if bosses != 1 {
		return fail("%d bosses", bosses)
	}

	// Needs pick the enemies: each need below the line breeds its family, in
	// the promised number; a need at or above it breeds nothing.
	for n := content.Need(0); n < content.NumNeeds; n++ {
		want := 0
		if sev := t.EnemyNeedBelow - d.Needs[n]; sev > 0 {
			want = t.EnemyCountBase + sev/t.EnemyCountPer
		}
		if got := perFamily[content.FamilyOf(n)]; got != want {
			return fail("%s at %d bred %d %s enemies, want %d", n, d.Needs[n], got, content.FamilyOf(n), want)
		}
	}
	// Buildings pick the special rooms.
	for b := range d.Buildings {
		want := RoomKind("")
		switch c := d.Buildings[b].Condition; {
		case c < t.RuinThreshold:
			want = RoomHazard
		case c >= t.HavenThreshold:
			want = RoomHaven
		}
		if special[b] != want {
			return fail("building %d at %d became %q, want %q", b, d.Buildings[b].Condition, special[b], want)
		}
	}
	if len(h.Rooms) != len(content.Layouts[d.ID].Rooms)+len(special) {
		return fail("room count does not match the layout plus special rooms")
	}
	return ""
}

func families(h *Hollow) map[content.Family]int {
	out := map[content.Family]int{}
	for i := range h.Rooms {
		for _, e := range h.Rooms[i].Enemies {
			if !e.Boss {
				out[content.Enemies[e.Kind].Family]++
			}
		}
	}
	return out
}

func kinds(h *Hollow) map[RoomKind]int {
	out := map[RoomKind]int{}
	for i := range h.Rooms {
		out[h.Rooms[i].Kind]++
	}
	return out
}

func boss(h *Hollow) *Enemy {
	for i := range h.Rooms[h.Heart].Enemies {
		if h.Rooms[h.Heart].Enemies[i].Boss {
			return &h.Rooms[h.Heart].Enemies[i]
		}
	}
	return nil
}

// Fixtures that prove the mapping from city state to hollow.
func TestGeneratorLegibility(t *testing.T) {
	type fam = map[content.Family]bool
	tests := []struct {
		name     string
		d        District
		families fam
		hazards  int
		havens   int
		boss     string
	}{
		{"starved market breeds Hunger only",
			district(content.Market, [3]int{10, 80, 80}), fam{content.Hunger: true}, 0, 0, "The Starved Grain Factor"},
		{"lawless watch breeds Riot only; the unpaid captain is the boss",
			district(content.Watch, [3]int{80, 10, 80}), fam{content.Riot: true}, 0, 0, "The Unpaid Captain"},
		{"crumbling mills breed Rot only",
			district(content.Mills, [3]int{80, 80, 10}), fam{content.Rot: true}, 0, 0, "The Crumbling Mill Reeve"},
		{"two low needs breed two families; the lower one is the grievance",
			district(content.Docks, [3]int{30, 20, 90}), fam{content.Hunger: true, content.Riot: true}, 0, 0, "The Unpaid Harbourmaster"},
		{"all three low breed all three",
			district(content.Tenements, [3]int{5, 5, 5}), fam{content.Hunger: true, content.Riot: true, content.Rot: true}, 0, 0, "The Starved Rentmaster"},
		{"needs at the line breed nothing, but the figure still waits",
			district(content.Temple, [3]int{50, 50, 50}), fam{}, 0, 0, "The Starved Prior"},
		{"one point under the line is enough",
			district(content.Temple, [3]int{80, 80, 49}), fam{content.Rot: true}, 0, 0, "The Crumbling Prior"},
		{"a ruined building becomes a hazard room",
			district(content.Market, [3]int{60, 60, 60}, 29, 50, 50), fam{}, 1, 0, "The Starved Grain Factor"},
		{"a kept building becomes a safe room",
			district(content.Market, [3]int{60, 60, 60}, 70, 50, 50), fam{}, 0, 1, "The Starved Grain Factor"},
		{"buildings between the lines become nothing",
			district(content.Market, [3]int{60, 60, 60}, 30, 69, 50), fam{}, 0, 0, "The Starved Grain Factor"},
		{"every building can become a room",
			district(content.Docks, [3]int{60, 60, 60}, 0, 100, 10), fam{}, 2, 1, "The Starved Harbourmaster"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := NewStream(1, 4)
			h := GenerateHollow(&tt.d, 1, &rng)
			if msg := checkHollow(&tt.d, &h); msg != "" {
				t.Fatal(msg)
			}
			got := families(&h)
			for f := content.Family(0); f < content.NumFamilies; f++ {
				if (got[f] > 0) != tt.families[f] {
					t.Errorf("%s enemies: %d, want present=%v", f, got[f], tt.families[f])
				}
			}
			k := kinds(&h)
			if k[RoomHazard] != tt.hazards || k[RoomHaven] != tt.havens {
				t.Errorf("%d hazards and %d havens, want %d and %d", k[RoomHazard], k[RoomHaven], tt.hazards, tt.havens)
			}
			if b := boss(&h); b == nil || b.Name != tt.boss {
				t.Errorf("boss %+v, want %s", b, tt.boss)
			}
		})
	}
}

// The lower the need, the more enemies and the stronger they are.
func TestLowerNeedMeansMoreAndStronger(t *testing.T) {
	prevCount, prevPower := 0, 0
	for _, supply := range []int{49, 40, 30, 18, 10, 0} {
		d := district(content.Market, [3]int{supply, 90, 90})
		rng := NewStream(3, 4)
		h := GenerateHollow(&d, 1, &rng)
		count, power := 0, 0
		for i := range h.Rooms {
			for _, e := range h.Rooms[i].Enemies {
				if !e.Boss {
					count++
					power += e.HP + 3*e.Attack
				}
			}
		}
		if count < prevCount || power <= prevPower {
			t.Errorf("Supply %d: %d enemies (power %d); the need above it had %d (power %d)", supply, count, power, prevCount, prevPower)
		}
		prevCount, prevPower = count, power
	}
	// And the boss grows with the grievance.
	mild, dire := district(content.Watch, [3]int{90, 45, 90}), district(content.Watch, [3]int{90, 5, 90})
	r1, r2 := NewStream(3, 4), NewStream(3, 4)
	hm, hd := GenerateHollow(&mild, 1, &r1), GenerateHollow(&dire, 1, &r2)
	if boss(&hd).HP <= boss(&hm).HP || boss(&hd).Attack < boss(&hm).Attack {
		t.Errorf("the boss did not grow with the grievance: %+v vs %+v", boss(&hm), boss(&hd))
	}
}

// The spec's own example: "Gaunts ×4: Market Supply was 18."
func TestGroupsReadLikeTheSpec(t *testing.T) {
	d := district(content.Market, [3]int{18, 80, 80})
	r := Room{}
	for range 4 {
		r.Enemies = append(r.Enemies, newEnemy(content.Gaunt, 32, needCause(&d, content.Supply)))
	}
	g := r.Groups()
	if len(g) != 1 || g[0].String() != "Gaunts ×4: Market Supply was 18" {
		t.Errorf("groups %v", g)
	}
	rng := NewStream(1, 4)
	h := GenerateHollow(&d, 1, &rng)
	total := 0
	for i := range h.Rooms {
		for _, grp := range h.Rooms[i].Groups() {
			if !grp.Boss {
				total += grp.Count
				if !strings.HasSuffix(grp.String(), "Market Supply was 18") {
					t.Errorf("group %q does not name its cause", grp)
				}
			}
		}
	}
	if total != 4+1 { // four in the rooms, one escort at the heart
		t.Errorf("Supply 18 bred %d Hunger enemies, want 4 and an escort", total)
	}
	b := boss(&h)
	if got := (Group{Name: b.Name, Count: 1, Boss: true, Cause: b.Cause}).String(); got != "The Starved Grain Factor: the Grain Factor's grievance: their lowest need was 18" {
		t.Errorf("boss line %q", got)
	}
}

func TestGenerationIsDeterministic(t *testing.T) {
	d := district(content.Docks, [3]int{12, 33, 47}, 10, 80, 50)
	r1, r2, r3 := NewStream(9, 4), NewStream(9, 4), NewStream(10, 4)
	a, b, c := GenerateHollow(&d, 3, &r1), GenerateHollow(&d, 3, &r2), GenerateHollow(&d, 3, &r3)
	if encode(t, a) != encode(t, b) {
		t.Error("the same district and stream gave two hollows")
	}
	// Another stream moves enemies between rooms and changes nothing else.
	if encode(t, families(&a)) != encode(t, families(&c)) || len(a.Rooms) != len(c.Rooms) {
		t.Error("the stream changed what the hollow contains, not just where")
	}
}

// randomDistrict draws an arbitrary district state.
func randomDistrict(rng *Stream) District {
	id := content.DistrictID(rng.IntN(int(content.NumDistricts)))
	d := District{ID: id, Figure: "Vane"}
	for n := range d.Needs {
		d.Needs[n] = rng.IntN(content.T.NeedMax + 1)
	}
	for range content.Districts[id].Buildings {
		d.Buildings = append(d.Buildings, Building{Condition: rng.IntN(content.T.NeedMax + 1)})
	}
	return d
}

// 500 fixed seeds: every room, enemy and boss has a cause that points at real
// state with the recorded value.
func TestCauseCoverage(t *testing.T) {
	seen := map[RoomKind]bool{}
	for seed := uint64(1); seed <= 500; seed++ {
		src := NewStream(seed, 99)
		d := randomDistrict(&src)
		gen := NewStream(seed, 4)
		h := GenerateHollow(&d, 1, &gen)
		if msg := checkHollow(&d, &h); msg != "" {
			t.Fatalf("seed %d: %s\nneeds %v buildings %v", seed, msg, d.Needs, d.Buildings)
		}
		for k := range kinds(&h) {
			seen[k] = true
		}
	}
	for _, k := range []RoomKind{RoomEntry, RoomFight, RoomHazard, RoomHaven, RoomHeart} {
		if !seen[k] {
			t.Errorf("500 seeds never produced a %s room", k)
		}
	}
}

// Districts taken from real play, not just drawn at random.
func TestCauseCoverageInPlay(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		s := start(t, seed)
		for day := 0; day < 12; day++ {
			s = pass(t, s)
			for id := range s.Districts {
				d := &s.Districts[id]
				h := GenerateHollow(d, s.Day, &s.RNG.Gen)
				if msg := checkHollow(d, &h); msg != "" {
					t.Fatalf("seed %d day %d: %s", seed, day, msg)
				}
			}
		}
	}
}

func FuzzGenerateHollow(f *testing.F) {
	f.Add(uint8(0), 10, 80, 80, 50, 50, 50, uint64(1))
	f.Add(uint8(1), 80, 10, 80, 20, 90, 0, uint64(2))
	f.Add(uint8(2), 0, 0, 0, 0, 0, 0, uint64(3))
	f.Add(uint8(5), 100, 100, 100, 100, 100, 100, uint64(4))
	f.Add(uint8(3), 49, 50, 51, 29, 30, 70, uint64(5))
	f.Fuzz(func(t *testing.T, id uint8, supply, order, upkeep, c0, c1, c2 int, seed uint64) {
		fit := func(v int) int { // any int to a legal need or condition
			m := content.T.NeedMax + 1
			return ((v % m) + m) % m
		}
		d := district(content.DistrictID(int(id)%int(content.NumDistricts)),
			[3]int{fit(supply), fit(order), fit(upkeep)}, fit(c0), fit(c1), fit(c2))
		rng := NewStream(seed, 4)
		h := GenerateHollow(&d, 1, &rng)
		if msg := checkHollow(&d, &h); msg != "" {
			t.Fatalf("%s\nneeds %v buildings %v seed %d", msg, d.Needs, d.Buildings, seed)
		}
	})
}
