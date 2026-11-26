// Package game is the Godot presentation of Hollow Ward. It renders sim state
// and sends commands; it never mutates state. It is the only package that may
// import graphics.gd.
//
// The whole UI is rebuilt from the state after every command. Every button
// has an action id; pressing the button and pressing the id (which is what
// autoplay does) run the same code.
package game

import (
	"strconv"
	"strings"

	"graphics.gd/classdb"
	"graphics.gd/classdb/OS"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
)

// Main registers the Go-defined classes, builds the scene and runs the engine
// until it shuts down.
func Main() {
	classdb.Register[App]()
	startup.LoadingScene()
	SceneTree.Add(&App{opts: parseOptions(OS.GetCmdlineUserArgs())})
	startup.Scene()
}

// options are the user arguments, the ones after "--" on the engine's command
// line (gd run -- --autoplay).
type options struct {
	autoplay   bool   // play a scripted session through the UI, then quit
	fail       bool   // autoplay reports a failure (proves the exit code path)
	screenshot string // save the first screen to this PNG, then quit
	shots      string // autoplay saves a PNG of each screen into this folder
	seed       uint64 // seed for new tenures (0: from the clock)
}

func parseOptions(args []string) options {
	var o options
	for _, arg := range args {
		switch {
		case arg == "--autoplay":
			o.autoplay = true
		case arg == "--autoplay-fail":
			o.autoplay, o.fail = true, true
		case strings.HasPrefix(arg, "--screenshot="):
			o.screenshot = strings.TrimPrefix(arg, "--screenshot=")
		case strings.HasPrefix(arg, "--shots="):
			o.shots = strings.TrimPrefix(arg, "--shots=")
		case strings.HasPrefix(arg, "--seed="):
			if v, err := strconv.ParseUint(strings.TrimPrefix(arg, "--seed="), 10, 64); err == nil {
				o.seed = v
			}
		}
	}
	return o
}
