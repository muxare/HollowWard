package main

import (
	"flag"
	"fmt"
	"io"

	"hollowward/internal/policy"
	"hollowward/internal/sim"
)

// play runs a scripted policy and prints the event log as text.
func play(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	seed := fs.Uint64("seed", 1, "tenure seed")
	name := fs.String("policy", "repair-worst", "scripted player: random, repair-worst or farm")
	nights := fs.Int("nights", 1, "nights to play (0: the whole tenure)")
	quiet := fs.Bool("q", false, "print only the summary")
	if err := fs.Parse(args); err != nil {
		return err
	}
	p, err := policy.New(*name, *seed)
	if err != nil {
		return err
	}
	s, events, err := sim.Apply(sim.State{}, sim.NewTenure(*seed))
	if err != nil {
		return err
	}
	stop := func(s *sim.State) bool { return *nights > 0 && s.Day > *nights }
	s, more, err := policy.Play(s, p, stop)
	events = append(events, more...)
	if !*quiet {
		for _, e := range events {
			if line := e.Text(); line != "" {
				fmt.Fprintln(out, line)
			}
		}
		fmt.Fprintln(out)
	}
	if err != nil {
		return err
	}
	summary(out, &s, *name, *seed)
	return nil
}

func summary(out io.Writer, s *sim.State, name string, seed uint64) {
	turned := 0
	for _, h := range s.Hollows {
		if h != nil {
			turned++
		}
	}
	state := fmt.Sprintf("day %d", s.Day)
	if s.Phase == sim.PhaseOver {
		state = "over: " + string(s.Outcome)
	}
	fmt.Fprintf(out, "policy %s, seed %d: %s; %d of %d citizens alive; %d districts turned; coin %d, grain %d, timber %d; hungerbone %d, ash-masks %d, blackcap %d\n",
		name, seed, state, s.Living(), len(s.Citizens), turned,
		s.Resources[0], s.Resources[1], s.Resources[2], s.Materials[0], s.Materials[1], s.Materials[2])
}
