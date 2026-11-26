// Package policy holds the scripted players (random, repair-worst, farm) that
// drive the sim for simcli, the balance harness and the smoke test.
//
// A policy reads a state and returns the next command; it never touches state
// directly and draws randomness only from a seeded source it is given.
package policy
