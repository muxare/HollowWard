package game

import (
	"strings"
	"testing"

	"hollowward/internal/content"
	"hollowward/internal/policy"
	"hollowward/internal/sim"
)

func TestSmallTexts(t *testing.T) {
	if got := tenths(235); got != "23.5" {
		t.Errorf("tenths(235) = %q", got)
	}
	if got := tenths(1000); got != "100.0" {
		t.Errorf("tenths(1000) = %q", got)
	}
	if got := costText(content.Cost{content.Coin: 4, content.Timber: 6}); got != "4 coin, 6 timber" {
		t.Errorf("costText = %q", got)
	}
	if got := costText(content.Cost{}); got != "free" {
		t.Errorf("costText of nothing = %q", got)
	}
	if got := matsText(content.Materials{4, 4, 0}); got != "4 hungerbone, 4 ash-masks" {
		t.Errorf("matsText = %q", got)
	}
	for n, want := range map[int]string{1: "once", 2: "twice", 7: "7 times"} {
		if got := times(n); got != want {
			t.Errorf("times(%d) = %q", n, got)
		}
	}
	if got := riseText(sim.Rise{Reason: sim.RiseOpen, District: content.Docks, Tenths: 150}); got != "+15.0  the Docks are still turned from an earlier night" {
		t.Errorf("riseText = %q", got)
	}
	if got := riseText(sim.Rise{Reason: sim.RiseTurned, District: content.Market, Tenths: 61}); !strings.HasPrefix(got, "+6.1  the Market turned tonight") {
		t.Errorf("riseText = %q", got)
	}
}

func TestCauseLine(t *testing.T) {
	d := sim.District{ID: content.Market, Needs: [3]int{18, 80, 80}, Buildings: []sim.Building{{Condition: 10}, {Condition: 80}, {Condition: 80}}}
	rep := sim.Neglect(&d)
	if got := causeLine(rep.Causes[0]); got != "27.3  Market Supply was 18" {
		t.Errorf("largest cause %q", got)
	}
	if got := causeLine(rep.Causes[1]); got != "10.0  The Granary was at 10" {
		t.Errorf("second cause %q", got)
	}
}

// Every command a policy can send must have its own button id, or autoplay
// and the player could not both reach it.
func TestActionIDsAreDistinct(t *testing.T) {
	seen := map[string]string{}
	check := func(c sim.Command) {
		id := actionID(c)
		key := string(c.Kind) + "|" + id
		if id == "" {
			t.Errorf("%s has no action id", c.Kind)
		}
		if other, ok := seen[id]; ok && other != key {
			t.Errorf("action id %q is shared by %s and %s", id, other, key)
		}
		seen[id] = key
	}
	for st := 0; st < content.NumStewards; st++ {
		check(sim.RecallSteward(st))
		for d := content.DistrictID(0); d < content.NumDistricts; d++ {
			for n := content.NoNeed; n < content.NumNeeds; n++ {
				for b := -1; b < 3; b++ {
					c := sim.AssignSteward(st, d, n, b)
					if id := actionID(c); seen[id] != "" {
						t.Fatalf("assign id %q repeats", id)
					}
					check(c)
				}
			}
		}
	}
	for id := content.DecreeID(0); id < content.NumDecrees; id++ {
		check(sim.IssueDecree(id, content.Mills))
	}
	for id := content.UpgradeID(0); id < content.NumUpgrades; id++ {
		check(sim.BuyUpgrade(id, content.Mills, 1))
	}
	for i := 0; i < 3; i++ {
		check(sim.ResolveEvent(i))
		check(sim.Advance(i))
		check(sim.Act(sim.ActAttack, i))
		check(sim.Act(sim.ActAbility, i))
	}
	check(sim.Act(sim.ActDefend, -1))
	check(sim.EndDay())
	check(sim.StayHome())
	check(sim.GoHome())
	check(sim.SendParty(content.Docks, 1, 2))
}

// The fight screen's words: every living enemy says what it will do.
func TestIntentText(t *testing.T) {
	g, err := sim.NewGame(3)
	if err != nil {
		t.Fatal(err)
	}
	p := policy.RepairWorst{}
	for step := 0; step < 5000 && (g.State.Expedition == nil || g.State.Expedition.Fight == nil); step++ {
		if g.State.Phase == sim.PhaseOver {
			t.Fatal("the tenure ended without a fight")
		}
		if _, err := g.Do(p.Next(&g.State)); err != nil {
			t.Fatal(err)
		}
	}
	s := &g.State
	room := &s.Hollows[s.Expedition.District].Rooms[s.Expedition.Room]
	for i := range room.Enemies {
		e := &room.Enemies[i]
		if e.HP <= 0 {
			continue
		}
		got := intentText(s, e)
		if !strings.Contains(got, "it ") || strings.Contains(got, "%") || strings.Contains(got, "nobody") {
			t.Errorf("%s: intent %q", e.Name, got)
		}
		if !e.Acted && !strings.Contains(got, "Then it") {
			t.Errorf("%s has not acted, so both its actions must show: %q", e.Name, got)
		}
		if tipText := roomTip(room); !strings.Contains(tipText, "Why: ") {
			t.Errorf("room tooltip lacks a cause: %q", tipText)
		}
	}
	for _, m := range s.Expedition.Party {
		c := &s.Citizens[m.Citizen]
		if tipText := traitsTip(c); !strings.Contains(tipText, c.Name) || !strings.Contains(tipText, "Works at") {
			t.Errorf("citizen tooltip %q", tipText)
		}
	}
}

// The after-action report keeps the causes and drops the blow-by-blow.
func TestReportLines(t *testing.T) {
	g, err := sim.NewGame(3)
	if err != nil {
		t.Fatal(err)
	}
	p := policy.RepairWorst{}
	fought := false
	for step := 0; step < 20000 && g.State.Phase != sim.PhaseOver; step++ {
		events, err := g.Do(p.Next(&g.State))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range events {
			fought = fought || e.Kind == sim.EvFightStarted
		}
		if fought && g.State.Phase == sim.PhaseDay {
			break
		}
	}
	lines := reportLines(g.Tonight())
	text := strings.Join(lines, "\n")
	for _, want := range []string{"Neglect", "Waiting here:", " was ", "Dawn."} {
		if !strings.Contains(text, want) {
			t.Errorf("report lacks %q:\n%s", want, text)
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "Round ") || strings.Contains(line, "strikes for") {
			t.Errorf("the report carries blow-by-blow: %q", line)
		}
		if strings.Contains(line, "Waiting here:") && !strings.Contains(line, "(") {
			t.Errorf("no cause on %q", line)
		}
	}
}
