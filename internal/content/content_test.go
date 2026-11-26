package content

import "testing"

func TestDistrictsShape(t *testing.T) {
	total := 0
	for id, d := range Districts {
		if d.ID != DistrictID(id) {
			t.Errorf("%s: ID %d at index %d", d.Name, d.ID, id)
		}
		if n := len(d.Buildings); n < 2 || n > 3 {
			t.Errorf("%s: %d buildings, want 2 or 3", d.Name, n)
		}
		if n := len(d.Jobs); n < 2 || n > 4 {
			t.Errorf("%s: %d residents, want 2 to 4", d.Name, n)
		}
		if d.FigureTitle == "" || d.Slug == "" {
			t.Errorf("%s: missing figure title or slug", d.Name)
		}
		seen := map[string]bool{}
		for _, b := range d.Buildings {
			if b.Slug == "" || b.Hazard == "" || b.Haven == "" || seen[b.Slug] {
				t.Errorf("%s: bad or repeated building %+v", d.Name, b)
			}
			seen[b.Slug] = true
		}
		total += len(d.Jobs)
	}
	if total < 18 || total > 22 {
		t.Errorf("%d citizens in all, want about 20", total)
	}
	if len(CitizenNames) < total {
		t.Errorf("%d names for %d citizens", len(CitizenNames), total)
	}
	if len(FigureNames) < int(NumDistricts) {
		t.Errorf("%d figure names for %d districts", len(FigureNames), NumDistricts)
	}
}

func checkEffects(t *testing.T, where string, effects []Effect, hasTarget bool) {
	t.Helper()
	for _, e := range effects {
		switch e.Scope {
		case ScopeFixed:
			if !e.District.Valid() {
				t.Errorf("%s: fixed effect without a district", where)
			}
		case ScopeTarget:
			if !hasTarget {
				t.Errorf("%s: target effect but nothing is targeted", where)
			}
		case ScopeCity:
			if e.Kind != EffResource {
				t.Errorf("%s: only resources are city-wide", where)
			}
		}
		if e.Kind == EffNeed && e.Need != NoNeed && !e.Need.Valid() {
			t.Errorf("%s: bad need", where)
		}
		if e.Amount == 0 {
			t.Errorf("%s: effect of size 0", where)
		}
	}
}

func TestEventCards(t *testing.T) {
	if len(EventCards) != 12 {
		t.Fatalf("%d event cards, the slice needs 12", len(EventCards))
	}
	slugs := map[string]bool{}
	for _, c := range EventCards {
		if c.Slug == "" || slugs[c.Slug] {
			t.Errorf("card %q: missing or repeated slug", c.Title)
		}
		slugs[c.Slug] = true
		if len(c.Choices) < 2 {
			t.Errorf("%s: a card needs a choice", c.Slug)
		}
		free := false
		for _, ch := range c.Choices {
			free = free || ch.Cost.IsZero()
			for _, v := range ch.Cost {
				if v < 0 {
					t.Errorf("%s: negative cost", c.Slug)
				}
			}
			if ch.Label == "" || ch.Result == "" {
				t.Errorf("%s: choice without label or result", c.Slug)
			}
			checkEffects(t, c.Slug, ch.Effects, true)
		}
		if !free {
			t.Errorf("%s: no free choice, so the card could block the day", c.Slug)
		}
	}
}

func TestDecrees(t *testing.T) {
	for id, d := range Decrees {
		if d.ID != DecreeID(id) || d.Slug == "" || d.Name == "" {
			t.Errorf("decree %d: bad ID or names", id)
		}
		if d.Cost.IsZero() {
			t.Errorf("%s: a decree must cost something", d.Name)
		}
		checkEffects(t, d.Slug, d.Effects, d.Targeted)
	}
}

func TestCitizenTables(t *testing.T) {
	if NumTraits != 12 {
		t.Errorf("%d traits, want about 12", NumTraits)
	}
	for id, j := range Jobs {
		if j.ID != JobID(id) || j.Slug == "" {
			t.Errorf("job %d: bad ID or slug", id)
		}
		if st := j.Stats(); st.HP <= 0 || st.Attack <= 0 || st.Speed <= 0 {
			t.Errorf("%s: stats %+v", j.Name, st)
		}
		if Abilities[j.Ability].Name == "" {
			t.Errorf("%s: no ability", j.Name)
		}
	}
	for id, tr := range Traits {
		if tr.ID != TraitID(id) || tr.Name == "" || tr.Text == "" {
			t.Errorf("trait %d: bad ID or text", id)
		}
		if opp, ok := Opposite(tr.ID); ok {
			if back, ok2 := Opposite(opp); !ok2 || back != tr.ID {
				t.Errorf("%s: opposite is not mutual", tr.Name)
			}
		}
	}
	for id, sc := range Scars {
		m := sc.Mods()
		if sc.ID != ScarID(id) || (m == Mods{}) {
			t.Errorf("scar %d: every scar must cost something", id)
		}
	}
}
