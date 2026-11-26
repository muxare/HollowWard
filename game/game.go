// Package game is the Godot presentation of Hollow Ward. It renders sim state
// and sends commands; it never mutates state. It is the only package that may
// import graphics.gd.
//
// At M0 it holds the toolchain spike only: a Go-defined node that draws a
// rectangle and a label.
package game

import (
	"fmt"
	"os"
	"strings"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Label"
	"graphics.gd/classdb/Node2D"
	"graphics.gd/classdb/OS"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/classdb/Viewport"
	"graphics.gd/startup"
	"graphics.gd/variant/Color"
	"graphics.gd/variant/Float"
	"graphics.gd/variant/Rect2"
	"graphics.gd/variant/Vector2"
)

// Main registers the Go-defined classes, builds the scene and runs the engine
// until it shuts down.
func Main() {
	classdb.Register[Spike]()
	startup.LoadingScene()
	SceneTree.Add(&Spike{opts: parseOptions(OS.GetCmdlineUserArgs())})
	startup.Scene()
}

// options are the user arguments, the ones after "--" on the engine's command
// line (gd run -- --autoplay).
type options struct {
	autoplay   bool   // run a scripted session, then quit.
	screenshot string // save the first rendered frames to this PNG.
	fail       bool   // autoplay reports a failure (proves the exit code path).
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
		}
	}
	return o
}

// Spike is the M0 proof that a node defined in Go draws in a Godot window.
type Spike struct {
	Node2D.Extension[Spike] `gd:"HollowWardSpike"`

	opts   options
	frames int
	drawn  bool
}

const spikeFrames = 10 // frames to render before autoplay or a screenshot ends the run.

func (s *Spike) Ready() {
	label := Label.New()
	label.SetText("Hollow Ward: toolchain spike (Go node, Godot window)")
	label.AsControl().SetPosition(Vector2.New(60, 50))
	s.AsNode().AddChild(label.AsNode())
}

func (s *Spike) Draw() {
	s.AsCanvasItem().DrawRect(
		Rect2.PositionSize{Position: Vector2.New(60, 100), Size: Vector2.New(280, 140)},
		Color.RGBA{R: 0.55, G: 0.12, B: 0.16, A: 1},
	)
	s.drawn = true
}

func (s *Spike) Process(delta Float.X) {
	s.frames++
	if s.frames < spikeFrames || (!s.opts.autoplay && s.opts.screenshot == "") {
		return
	}
	code := 0
	if s.opts.screenshot != "" {
		image := Viewport.Get(s.AsNode()).GetTexture().AsTexture2D().GetImage()
		if err := image.SavePng(s.opts.screenshot); err != nil {
			fmt.Fprintln(os.Stderr, "spike: screenshot:", err)
			code = 1
		}
	}
	if s.opts.autoplay {
		switch {
		case s.opts.fail:
			fmt.Fprintln(os.Stderr, "spike: autoplay asked to fail")
			code = 1
		case !s.drawn:
			fmt.Fprintln(os.Stderr, "spike: Draw was never called")
			code = 1
		default:
			fmt.Println("spike: autoplay ok,", s.frames, "frames")
		}
	}
	SceneTree.Get(s.AsNode()).MoreArgs().Quit(code)
}
