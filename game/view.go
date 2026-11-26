package game

import (
	"fmt"
	"strings"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

// This file turns state into words. It has no engine calls, so plain Go tests
// cover it.

// tenths prints a value kept in tenths of a point: 235 is "23.5".
func tenths(v int) string { return fmt.Sprintf("%d.%d", v/10, v%10) }

func costText(c content.Cost) string {
	var parts []string
	for r, v := range c {
		if v > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", v, content.Resource(r)))
		}
	}
	if len(parts) == 0 {
		return "free"
	}
	return strings.Join(parts, ", ")
}

func matsText(m content.Materials) string {
	var parts []string
	for f, v := range m {
		if v > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", v, content.Family(f).Material()))
		}
	}
	return strings.Join(parts, ", ")
}

// causeLine is one line of the Neglect breakdown: "27.3  Market Supply was 18".
func causeLine(c sim.NeglectCause) string {
	return fmt.Sprintf("%4.1f  %s", float64(c.Thirds)/3, capital(c.Describe()))
}

func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// riseText says why the Dreamer rose: one line of the dusk screen.
func riseText(r sim.Rise) string {
	switch r.Reason {
	case sim.RiseTurned:
		return fmt.Sprintf("+%s  the %s turned tonight (its Neglect, over ten)", tenths(r.Tenths), r.District)
	case sim.RiseOpen:
		return fmt.Sprintf("+%s  the %s %s still turned from an earlier night", tenths(r.Tenths), r.District, r.District.Verb("is", "are"))
	}
	return fmt.Sprintf("+%s  a quiet night: it stirs anyway", tenths(r.Tenths))
}

func citizenName(s *sim.State, id int) string {
	if id < 0 || id >= len(s.Citizens) {
		return "nobody"
	}
	return s.Citizens[id].Name
}

func moveText(s *sim.State, in sim.Intent) string {
	t := content.Moves[in.Move].Text
	if strings.Contains(t, "%s") {
		return fmt.Sprintf(t, citizenName(s, in.Target))
	}
	return t
}

// intentText is what an enemy has announced: its next action and the one
// after. Both are always on show.
func intentText(s *sim.State, e *sim.Enemy) string {
	if e.Stunned {
		return "Stunned: loses its next action. After that it " + moveText(s, e.Upcoming()) + "."
	}
	if e.Acted {
		return "Has acted. Next round it " + moveText(s, e.Next) + "."
	}
	return capital("It "+moveText(s, e.Intent)) + ". Then it " + moveText(s, e.Next) + "."
}

func traitsText(c *sim.Citizen) string {
	parts := []string{content.Traits[c.Traits[0]].Name, content.Traits[c.Traits[1]].Name}
	for _, sc := range c.Scars {
		parts = append(parts, "scar: "+content.Scars[sc].Name)
	}
	return strings.Join(parts, ", ")
}

func traitsTip(c *sim.Citizen) string {
	var b strings.Builder
	job := content.Jobs[c.Job]
	fmt.Fprintf(&b, "%s, %s of the %s\n", c.Name, job.Name, c.Home)
	fmt.Fprintf(&b, "%s: %s\n", content.Abilities[job.Ability].Name, content.Abilities[job.Ability].Text)
	for _, t := range c.Traits {
		fmt.Fprintf(&b, "%s: %s\n", content.Traits[t].Name, content.Traits[t].Text)
	}
	for _, sc := range c.Scars {
		fmt.Fprintf(&b, "%s (scar): %s\n", content.Scars[sc].Name, content.Scars[sc].Text)
	}
	fmt.Fprintf(&b, "Works at %d%% today.", c.WorkPct())
	return b.String()
}

func statsText(c *sim.Citizen) string {
	st := c.Stats()
	return fmt.Sprintf("HP %d/%d  Atk %d  Grd %d  Spd %d", c.HP, st.HP, st.Attack, st.Guard, st.Speed)
}

func memberStatus(m *sim.Member) string {
	var parts []string
	if m.Defending {
		parts = append(parts, "braced")
	}
	if m.Holding >= 0 {
		parts = append(parts, "holding")
	}
	if m.Burning > 0 {
		parts = append(parts, fmt.Sprintf("burning %d", m.Burning))
	}
	if m.Delayed {
		parts = append(parts, "shoved back")
	}
	if m.AttackMod != 0 {
		parts = append(parts, fmt.Sprintf("Atk %+d", m.AttackMod))
	}
	if m.GuardMod != 0 {
		parts = append(parts, fmt.Sprintf("Grd %+d", m.GuardMod))
	}
	if m.Cooldown > 0 {
		parts = append(parts, "ability resting")
	}
	return strings.Join(parts, ", ")
}

func enemyStatus(e *sim.Enemy) string {
	var parts []string
	if e.Hooked {
		parts = append(parts, "hooked: acts last next round")
	}
	if e.AttackMod != 0 {
		parts = append(parts, fmt.Sprintf("Atk %+d", e.AttackMod))
	}
	return strings.Join(parts, ", ")
}

// reportLines is the after-action report: the night's events as text, each
// generated thing with its cause.
func reportLines(events []sim.Event) []string {
	var out []string
	for _, e := range events {
		switch e.Kind {
		case sim.EvRound, sim.EvCitizenActed, sim.EvEnemyActed, sim.EvDayStarted, sim.EvCardDrawn, sim.EvFightStarted:
			continue // the blow-by-blow stays in the fight; the report is what it added up to
		}
		if line := strings.TrimSpace(e.Text()); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// roomTip is the hover text for a room on the map: what it is and why.
func roomTip(r *sim.Room) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\nWhy: %s", r.Name, r.Cause.Describe())
	switch r.Kind {
	case sim.RoomHazard:
		fmt.Fprintf(&b, "\nHazard: everyone who enters loses %d HP.", r.Hazard)
	case sim.RoomHaven:
		b.WriteString("\nSafe room: rest and healing, no enemies.")
	case sim.RoomHeart:
		b.WriteString("\nThe heart: clear it and the district is a district again.")
	}
	if r.Visited {
		for _, g := range r.Groups() {
			b.WriteString("\n" + g.String())
		}
	}
	if r.Cleared {
		b.WriteString("\nCleared.")
	}
	return b.String()
}

// actionID names the UI action for a command: the id of the button that
// sends it. Autoplay presses buttons by these ids.
func actionID(c sim.Command) string {
	switch c.Kind {
	case sim.CmdAssignSteward:
		return fmt.Sprintf("assign:%d:%s:%d:%d", c.Steward, c.District.Slug(), c.Need, c.Building)
	case sim.CmdRecallSteward:
		return fmt.Sprintf("recall:%d", c.Steward)
	case sim.CmdIssueDecree:
		if content.Decrees[c.Decree].Targeted {
			return fmt.Sprintf("decree:%s:%s", content.Decrees[c.Decree].Slug, c.District.Slug())
		}
		return "decree:" + content.Decrees[c.Decree].Slug
	case sim.CmdResolveEvent:
		return fmt.Sprintf("choice:%d", c.Choice)
	case sim.CmdBuyUpgrade:
		if content.Upgrades[c.Upgrade].Building {
			return fmt.Sprintf("upgrade:%s:%s:%d", content.Upgrades[c.Upgrade].Slug, c.District.Slug(), c.Building)
		}
		return "upgrade:" + content.Upgrades[c.Upgrade].Slug
	case sim.CmdAdvance:
		return fmt.Sprintf("advance:%d", c.Room)
	case sim.CmdAct:
		return fmt.Sprintf("act:%s:%d", c.Action, c.Target)
	}
	return string(c.Kind) // end-day, stay-home, go-home, send-party
}

// times says how often: "once", "twice", "5 times".
func times(n int) string {
	switch n {
	case 1:
		return "once"
	case 2:
		return "twice"
	}
	return fmt.Sprintf("%d times", n)
}
