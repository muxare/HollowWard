// Command balance runs many tenures per scripted policy in parallel and prints
// the win rate, mean nights survived and how often each district turned.
//
//	balance [-n 1000] [-seed 1] [-policy all] [-workers N] [-check]
//
// With -check it exits non-zero unless the balance targets are met.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"hollowward/internal/content"
	"hollowward/internal/policy"
	"hollowward/internal/sim"
)

func main() {
	ok, err := run(os.Args[1:], os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "balance:", err)
		os.Exit(2)
	}
	if !ok {
		os.Exit(1)
	}
}

// result is one tenure, summed up.
type result struct {
	outcome sim.Outcome
	nights  int
	dreamer int
	dead    int
	turned  [content.NumDistricts]int
	cleared int
	mats    content.Materials
	err     error
}

// tally is every tenure of one policy.
type tally struct {
	name    string
	n       int
	wins    int
	lid     int
	nobody  int
	nights  int
	dreamer int
	dead    int
	turned  [content.NumDistricts]int
	cleared int
	mats    content.Materials
}

func (t *tally) winRate() float64 { return 100 * float64(t.wins) / float64(max(t.n, 1)) }

// tenure plays one tenure. It owns its state: nothing is shared with any
// other goroutine.
func tenure(name string, seed uint64) result {
	var r result
	p, err := policy.New(name, seed)
	if err != nil {
		r.err = err
		return r
	}
	s, events, err := sim.Apply(sim.State{}, sim.NewTenure(seed))
	if err != nil {
		r.err = err
		return r
	}
	s, more, err := policy.Play(s, p, nil)
	if err != nil {
		r.err = err
		return r
	}
	for _, e := range append(events, more...) {
		switch e.Kind {
		case sim.EvTurned:
			r.turned[e.District]++
		case sim.EvHeartCleared:
			r.cleared++
		}
	}
	r.outcome, r.nights, r.dreamer = s.Outcome, s.Nights, s.Dreamer
	r.dead = len(s.Citizens) - s.Living()
	r.mats = s.Materials
	return r
}

func run(args []string, out io.Writer) (bool, error) {
	fs := flag.NewFlagSet("balance", flag.ContinueOnError)
	n := fs.Int("n", 1000, "tenures per policy")
	seed := fs.Uint64("seed", 1, "first seed; tenure i uses seed+i")
	which := fs.String("policy", "all", "policy to run, or all")
	workers := fs.Int("workers", runtime.NumCPU(), "tenures to run at once")
	check := fs.Bool("check", false, "exit non-zero unless the balance targets are met")
	if err := fs.Parse(args); err != nil {
		return false, err
	}
	names := policy.Names()
	if *which != "all" {
		if _, err := policy.New(*which, 1); err != nil {
			return false, err
		}
		names = []string{*which}
	}

	tallies := map[string]*tally{}
	for _, name := range names {
		t := &tally{name: name, n: *n}
		tallies[name] = t
		results := make([]result, *n)
		var wg sync.WaitGroup
		jobs := make(chan int)
		for w := 0; w < max(*workers, 1); w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					results[i] = tenure(name, *seed+uint64(i)) // each index is written by one goroutine only
				}
			}()
		}
		for i := 0; i < *n; i++ {
			jobs <- i
		}
		close(jobs)
		wg.Wait()
		for _, r := range results {
			if r.err != nil {
				return false, r.err
			}
			switch r.outcome {
			case sim.Won:
				t.wins++
			case sim.LostLid:
				t.lid++
			case sim.LostNoParty:
				t.nobody++
			}
			t.nights += r.nights
			t.dreamer += r.dreamer
			t.dead += r.dead
			t.cleared += r.cleared
			for d := range r.turned {
				t.turned[d] += r.turned[d]
			}
			for f := range r.mats {
				t.mats[f] += r.mats[f]
			}
		}
	}

	fmt.Fprintf(out, "%d tenures per policy, seeds %d to %d, %d nights each\n\n", *n, *seed, *seed+uint64(*n)-1, content.T.Nights)
	fmt.Fprintf(out, "%-13s %6s %7s %8s %8s %6s %8s\n", "policy", "wins", "nights", "lid", "nobody", "dead", "dreamer")
	for _, name := range names {
		t := tallies[name]
		fn := float64(t.n)
		fmt.Fprintf(out, "%-13s %5.1f%% %7.2f %7.1f%% %7.1f%% %6.2f %8.1f\n", t.name, t.winRate(), float64(t.nights)/fn,
			100*float64(t.lid)/fn, 100*float64(t.nobody)/fn, float64(t.dead)/fn, float64(t.dreamer)/fn/10)
	}
	fmt.Fprintf(out, "\nwins: share of tenures that survived every night. nights: mean nights survived.\nlid, nobody: share lost to the Dreamer, or to having no citizen left. dead: mean citizens lost. dreamer: mean final meter.\n")

	fmt.Fprintf(out, "\nHow often each district turned, per tenure:\n%-13s", "policy")
	for d := content.DistrictID(0); d < content.NumDistricts; d++ {
		fmt.Fprintf(out, " %9s", d)
	}
	fmt.Fprintf(out, " %9s %8s\n", "all", "cleared")
	for _, name := range names {
		t := tallies[name]
		fn, all := float64(t.n), 0
		fmt.Fprintf(out, "%-13s", t.name)
		for d := range t.turned {
			fmt.Fprintf(out, " %9.2f", float64(t.turned[d])/fn)
			all += t.turned[d]
		}
		fmt.Fprintf(out, " %9.2f %8.2f\n", float64(all)/fn, float64(t.cleared)/fn)
	}
	fmt.Fprintf(out, "\nMaterials in hand at the end, per tenure:\n%-13s %10s %10s %10s\n", "policy", content.Hunger.Material(), content.Riot.Material(), content.Rot.Material())
	for _, name := range names {
		t := tallies[name]
		fmt.Fprintf(out, "%-13s %10.1f %10.1f %10.1f\n", t.name, float64(t.mats[0])/float64(t.n), float64(t.mats[1])/float64(t.n), float64(t.mats[2])/float64(t.n))
	}

	ok := true
	if rnd, rep, farm := tallies["random"], tallies["repair-worst"], tallies["farm"]; rnd != nil && rep != nil && farm != nil {
		fmt.Fprintf(out, "\nTargets:\n")
		for _, line := range targets(rnd.winRate(), rep.winRate(), farm.winRate()) {
			fmt.Fprintf(out, "  %s\n", line.text)
			ok = ok && line.met
		}
	} else if *check {
		return false, fmt.Errorf("-check needs every policy; run with -policy all")
	}
	if !*check {
		return true, nil
	}
	return ok, nil
}

type target struct {
	text string
	met  bool
}

// targets holds the three win rates to the spec.
func targets(random, repair, farm float64) []target {
	mark := func(ok bool) string {
		if ok {
			return "met"
		}
		return strings.ToUpper("not met")
	}
	t1 := random < 10
	t2 := repair >= 30 && repair <= 50
	t3 := farm >= 0.8*repair
	return []target{
		{fmt.Sprintf("random wins under 10%%: %.1f%% (%s)", random, mark(t1)), t1},
		{fmt.Sprintf("repair-worst wins 30 to 50%%: %.1f%% (%s)", repair, mark(t2)), t2},
		{fmt.Sprintf("farm wins at least 80%% of repair-worst (%.1f%%): %.1f%% (%s)", 0.8*repair, farm, mark(t3)), t3},
	}
}
