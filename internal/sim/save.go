package sim

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Game is a tenure in play: the state and the event log that produced it.
type Game struct {
	State State   `json:"state"`
	Log   []Event `json:"log"`
}

// NewGame starts a tenure from a seed.
func NewGame(seed uint64) (*Game, error) {
	g := &Game{}
	if _, err := g.Do(NewTenure(seed)); err != nil {
		return nil, err
	}
	return g, nil
}

// Do applies a command and keeps the log. A rejected command changes nothing.
func (g *Game) Do(c Command) ([]Event, error) {
	next, events, err := Apply(g.State, c)
	if err != nil {
		return nil, err
	}
	if c.Kind == CmdNewTenure {
		g.Log = nil // a new warden, a new chronicle
	}
	g.State, g.Log = next, append(g.Log, events...)
	return events, nil
}

// Tonight returns the events since the last dusk began: the after-action
// report is built from these.
func (g *Game) Tonight() []Event {
	for i := len(g.Log) - 1; i >= 0; i-- {
		if g.Log[i].Kind == EvDayEnded {
			return g.Log[i:]
		}
	}
	return g.Log
}

// save is the file format: a version, then the game.
type save struct {
	Version int `json:"version"`
	Game
}

// ErrSaveVersion means a save was written by a different version of the game.
var ErrSaveVersion = errors.New("save is from a different version")

// EncodeSave turns a game into bytes. The presentation layer writes them to
// disk; this package never touches a file.
func EncodeSave(g *Game) ([]byte, error) {
	return json.Marshal(save{Version: SaveVersion, Game: *g})
}

// DecodeSave turns bytes back into a game. The loaded game plays out the
// same future as the one that was saved.
func DecodeSave(data []byte) (*Game, error) {
	var s save
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("save is damaged: %w", err)
	}
	if s.Version != SaveVersion {
		return nil, fmt.Errorf("%w: file has %d, this game writes %d", ErrSaveVersion, s.Version, SaveVersion)
	}
	return &s.Game, nil
}

// EncodeLedger turns a Ledger into bytes.
func EncodeLedger(l *Ledger) ([]byte, error) {
	out := *l
	out.Version = SaveVersion
	return json.Marshal(out)
}

// DecodeLedger turns bytes back into a Ledger.
func DecodeLedger(data []byte) (*Ledger, error) {
	var l Ledger
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("ledger is damaged: %w", err)
	}
	if l.Version != SaveVersion {
		return nil, fmt.Errorf("%w: ledger has %d, this game writes %d", ErrSaveVersion, l.Version, SaveVersion)
	}
	return &l, nil
}
