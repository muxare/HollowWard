// Package sim holds the rules of Hollow Ward as a pure, deterministic core.
//
// State is plain structs that marshal to JSON. Every change goes through
// Apply(state, cmd) (State, []Event, error), and the event log it returns is
// the single source for the chronicle, the after-action report and the Ledger.
//
// The package imports no graphics.gd, no package game, no os, no time and no
// global math/rand; golangci-lint enforces this (see .golangci.yml). Randomness
// comes from math/rand/v2 PCG sources that live in the state, one stream per
// subsystem. The package holds no package-level mutable state.
package sim
