// Command simcli plays and generates Hollow Ward headlessly from a seed and a
// policy.
//
//	simcli generate -seed N [-days D] [-district NAME] [-needs S,O,U] [-conds A,B,C]
//
//	simcli play -seed N [-policy NAME] [-nights K] [-q]
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "simcli:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: simcli generate|play [flags]")
	}
	switch args[0] {
	case "generate":
		return generate(args[1:], out)
	case "play":
		return play(args[1:], out)
	}
	return fmt.Errorf("unknown command %q; want generate or play", args[0])
}
