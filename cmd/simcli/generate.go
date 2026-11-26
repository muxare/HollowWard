package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"hollowward/internal/content"
	"hollowward/internal/sim"
)

// generate prints a hollow as text. By default it starts a tenure from the
// seed, lets the city go untended for a few days and remakes the most
// neglected district. -needs and -conds set a district's state by hand.
func generate(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	seed := fs.Uint64("seed", 1, "tenure seed")
	days := fs.Int("days", 4, "days the city goes untended before the hollow is made")
	name := fs.String("district", "", "district to remake (default: the most neglected)")
	needs := fs.String("needs", "", "set Supply,Order,Upkeep by hand, e.g. 10,80,80")
	conds := fs.String("conds", "", "set building conditions by hand, e.g. 20,80,50")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, _, err := sim.Apply(sim.State{}, sim.NewTenure(*seed))
	if err != nil {
		return err
	}
	for i := 0; i < *days; i++ {
		// Answer the card with its free choice and let the day pass.
		for c, ch := range content.EventCards[s.Card.Card].Choices {
			if ch.Cost.IsZero() {
				if s, _, err = sim.Apply(s, sim.ResolveEvent(c)); err != nil {
					return err
				}
				break
			}
		}
		if s, _, err = sim.Apply(s, sim.EndDay()); err != nil {
			return err
		}
		if s.Phase != sim.PhaseDay {
			break // dusk has fallen; stop here
		}
	}

	id := content.NoDistrict
	if *name != "" {
		for d := content.DistrictID(0); d < content.NumDistricts; d++ {
			if strings.EqualFold(d.String(), *name) || d.Slug() == strings.ToLower(*name) {
				id = d
			}
		}
		if id == content.NoDistrict {
			return fmt.Errorf("no district called %q", *name)
		}
	} else {
		worst := -1
		for d := range s.Districts {
			if n := sim.Neglect(&s.Districts[d]).Neglect; n > worst {
				worst, id = n, content.DistrictID(d)
			}
		}
	}
	d := &s.Districts[id]
	if *needs != "" {
		v, err := ints(*needs, len(d.Needs))
		if err != nil {
			return fmt.Errorf("-needs: %w", err)
		}
		copy(d.Needs[:], v)
	}
	if *conds != "" {
		v, err := ints(*conds, len(d.Buildings))
		if err != nil {
			return fmt.Errorf("-conds: %w", err)
		}
		for i := range d.Buildings {
			d.Buildings[i].Condition = v[i]
		}
	}

	h := sim.GenerateHollow(d, s.Day, &s.RNG.Gen)
	printHollow(out, d, &h)
	return nil
}

func ints(list string, n int) ([]int, error) {
	parts := strings.Split(list, ",")
	if len(parts) != n {
		return nil, fmt.Errorf("want %d numbers, got %d", n, len(parts))
	}
	out := make([]int, n)
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || v < 0 || v > content.T.NeedMax {
			return nil, fmt.Errorf("%q is not a number from 0 to %d", p, content.T.NeedMax)
		}
		out[i] = v
	}
	return out, nil
}

func printHollow(out io.Writer, d *sim.District, h *sim.Hollow) {
	def := content.Districts[d.ID]
	rep := sim.Neglect(d)
	fmt.Fprintf(out, "THE HOLLOW OF THE %s, night %d\n", strings.ToUpper(def.Name), h.Night)
	fmt.Fprintf(out, "Supply %d, Order %d, Upkeep %d.", d.Needs[content.Supply], d.Needs[content.Order], d.Needs[content.Upkeep])
	for i, b := range d.Buildings {
		fmt.Fprintf(out, " %s %d.", def.Buildings[i].Name, b.Condition)
	}
	fmt.Fprintf(out, "\nNeglect %d, because:\n", rep.Neglect)
	for _, c := range rep.Causes {
		fmt.Fprintf(out, "  %4.1f  %s\n", float64(c.Thirds)/3, c.Describe())
	}
	fmt.Fprintln(out)

	// Rooms in order of depth, then ID.
	order := make([]int, len(h.Rooms))
	for i := range order {
		order[i] = i
	}
	for i := 1; i < len(order); i++ {
		for j := i; j > 0; j-- {
			a, b := &h.Rooms[order[j-1]], &h.Rooms[order[j]]
			if a.Depth < b.Depth || (a.Depth == b.Depth && a.ID < b.ID) {
				break
			}
			order[j-1], order[j] = order[j], order[j-1]
		}
	}
	for _, i := range order {
		r := &h.Rooms[i]
		exits := "the end"
		if len(r.Exits) > 0 {
			var to []string
			for _, x := range r.Exits {
				to = append(to, fmt.Sprintf("[%d]", x))
			}
			exits = "leads to " + strings.Join(to, " ")
		}
		fmt.Fprintf(out, "[%d] %s%s (%s, depth %d), %s\n", r.ID, strings.Repeat("  ", r.Depth), r.Name, r.Kind, r.Depth, exits)
		pad := strings.Repeat("  ", r.Depth) + "      "
		fmt.Fprintf(out, "%swhy: %s\n", pad, r.Cause.Describe())
		switch r.Kind {
		case sim.RoomHazard:
			fmt.Fprintf(out, "%shazard: everyone who enters loses %d HP\n", pad, r.Hazard)
		case sim.RoomHaven:
			fmt.Fprintf(out, "%ssafe room: rest and healing, no enemies\n", pad)
		}
		for _, g := range r.Groups() {
			fmt.Fprintf(out, "%s%s\n", pad, g)
		}
		if len(r.Enemies) == 0 && (r.Kind == sim.RoomEntry || r.Kind == sim.RoomFight) {
			fmt.Fprintf(out, "%sempty: no need here was low enough to breed anything\n", pad)
		}
	}
}
