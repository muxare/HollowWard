// Command hollowward is the Godot build of Hollow Ward.
//
// gd builds the package beside the graphics/ folder, so the entry point lives
// at the module root. It only hands over to package game, which is the one
// package allowed to import graphics.gd.
package main

import "hollowward/game"

func main() { game.Main() }
