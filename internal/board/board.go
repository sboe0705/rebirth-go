// Package board contains the board data model and the map loader.
package board

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Kind is the type of a board space (Rebirth rules: SCO p. 2, IRL p. 2).
type Kind int

const (
	Empty      Kind = iota // no symbol: food or energy facility allowed
	Food                   // food symbol (leaves): food facility
	Energy                 // energy symbol (lightning bolt): energy facility
	Settlement             // house symbol: settlement tile
	Castle                 // never holds tiles; castle control by adjacency
	Cathedral              // never holds tiles; adjacency gives a mission card (Scotland)
	RoundTower             // never holds tiles; adjacency gives a bonus (Ireland)
)

// Buildable reports whether player tiles can be placed on this kind of space.
func (k Kind) Buildable() bool { return k <= Settlement }

// String returns the name of the space kind.
func (k Kind) String() string {
	switch k {
	case Empty:
		return "empty space"
	case Food:
		return "food space"
	case Energy:
		return "energy space"
	case Settlement:
		return "settlement space"
	case Castle:
		return "castle"
	case Cathedral:
		return "cathedral"
	case RoundTower:
		return "round tower"
	}
	return "unknown space"
}

// Bonus is a "+1" marker on a coastal space: placing a tile there scores 1.
type Bonus int

const (
	NoBonus       Bonus = iota
	Harbor              // on settlement spaces (Scotland)
	WindCollector       // on empty, food or energy spaces, never in settlements (Ireland)
)

func (x Bonus) String() string {
	switch x {
	case Harbor:
		return "harbor"
	case WindCollector:
		return "wind collector"
	}
	return "none"
}

// kindChars maps map characters to space kinds. Lowercase marks a "+1"
// space: 's' settlement with harbor, 'o' / 'f' / 'e' with wind collector.
var kindChars = map[rune]Kind{
	'O': Empty, 'F': Food, 'E': Energy, 'S': Settlement,
	'C': Castle, 'K': Cathedral, 'R': RoundTower,
}

// Tile is a board space. Water is not modelled: no entry means no space.
type Tile struct {
	Kind  Kind
	Bonus Bonus           // "+1" marker: harbor or wind collector
	Area  *SettlementArea // settlement this space belongs to; nil unless Kind == Settlement
	// Locked spaces are covered by a neutral tile in 2-player games (SCO p. 2), see Blocked.
	Locked bool

	Piece  Piece // placed tile, NoPiece if none
	Houses int   // settlement symbols on a placed settlement tile
	Owner  int   // player (0-based) who placed Piece
}

// MaxSettlementSize is the largest settlement: the extra-large one in Ireland.
const MaxSettlementSize = 4

// SettlementArea is one settlement: a contiguous group of settlement spaces.
// 1 space = small, 2 = medium, 3 = large, 4 = extra-large (Ireland).
// It is scored once all spaces are occupied. Separate settlements never touch.
type SettlementArea struct {
	Spaces []hex.Hex
}

// Size returns the number of spaces.
func (a *SettlementArea) Size() int { return len(a.Spaces) }

func parseChar(c rune) (*Tile, bool) {
	if k, ok := kindChars[c]; ok {
		return &Tile{Kind: k}, true
	}
	switch c {
	case 's':
		return &Tile{Kind: Settlement, Bonus: Harbor}, true
	case 'o', 'f', 'e':
		return &Tile{Kind: kindChars[c-'a'+'A'], Bonus: WindCollector}, true
	}
	return nil, false
}

// Board is an arbitrarily shaped hex board. A missing entry means no tile.
type Board struct {
	Name        string
	Tiles       map[hex.Hex]*Tile
	Settlements []*SettlementArea

	// Players is the number of players (turn order 0, 1, ...); 0 counts as 1.
	// It decides ties when a settlement is completed.
	Players int

	points    map[int]int     // points per player scored while placing tiles, see Score
	castles   map[hex.Hex]int // controlling player per castle space
	lastGains []Gain          // how the most recent placement scored

	deck       []Mission          // hidden mission cards, drawn from the front
	missions   map[int][]Mission  // mission cards held per player
	cathedrals map[hex.Hex][]int  // players with a cathedral per cathedral space, bottom first
	labels     map[hex.Hex]string // "C1", "K1", ... per castle / cathedral space
}

//go:embed maps/*.txt
var mapFS embed.FS

// Load loads an embedded map, e.g. Load("example").
func Load(name string) (*Board, error) {
	data, err := mapFS.ReadFile("maps/" + name + ".txt")
	if err != nil {
		return nil, err
	}
	return Parse(name, string(data))
}

// LoadFile loads a map from a file outside the program, e.g. a map you keep
// privately; it is named after the file without ".txt".
func LoadFile(path string) (*Board, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(strings.TrimSuffix(filepath.Base(path), ".txt"), string(data))
}

// Parse reads a map in text format: each line is a hex row
// (odd-r offset: odd rows are shifted right by half a tile),
// each character a column. '.' or ' ' = no tile, '#' at line start = comment.
// Adjacent settlement spaces form one settlement.
// An optional second grid after a line "[2 players]" has the same rows and
// columns; 'x' there marks a space locked in 2-player games.
// If the map only breaks the map rules (see validate), Parse returns the
// board together with the error, so it can still be shown.
func Parse(name, text string) (*Board, error) {
	b := &Board{Name: name, Tiles: map[hex.Hex]*Tile{}}
	sc := bufio.NewScanner(strings.NewReader(text))
	row, locks := 0, false
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.TrimSpace(line) == twoPlayers {
			if locks {
				return nil, fmt.Errorf("%s: %s appears twice", name, twoPlayers)
			}
			row, locks = 0, true
			continue
		}
		if locks {
			if err := b.parseLocks(line, row); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			row++
			continue
		}
		for col, c := range []rune(line) {
			if c == '.' || c == ' ' {
				continue
			}
			t, ok := parseChar(c)
			if !ok {
				return nil, fmt.Errorf("%s: unknown character %q in row %d, column %d", name, c, row, col)
			}
			b.Tiles[hex.FromOffset(col, row)] = t
		}
		row++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	b.groupSettlements()
	b.labelSpecials()
	if err := b.validate(); err != nil {
		return b, fmt.Errorf("%s: %w", name, err) // the board is usable, see Parse
	}
	return b, nil
}

// twoPlayers starts the grid of spaces locked in 2-player games.
const twoPlayers = "[2 players]"

// parseLocks reads one row of the "[2 players]" grid.
func (b *Board) parseLocks(line string, row int) error {
	for col, c := range []rune(line) {
		switch c {
		case '.', ' ':
		case 'x':
			t := b.Tiles[hex.FromOffset(col, row)]
			if t == nil {
				return fmt.Errorf("locked space at row %d, col %d is not on the map", row, col)
			}
			t.Locked = true
		default:
			return fmt.Errorf("unknown character %q in row %d, column %d of %s (only x and .)", c, row, col, twoPlayers)
		}
	}
	return nil
}

// Blocked reports whether space h is out of play: locked and a 2-player
// game. A blocked space counts as occupied by a neutral tile: nobody can
// place there, it belongs to no player, and a blocked castle can't be
// controlled.
func (b *Board) Blocked(h hex.Hex) bool {
	t := b.Tiles[h]
	return t != nil && t.Locked && b.Players == 2
}

// validate checks the map rules and reports every violation: settlements
// have at most MaxSettlementSize spaces; castles, cathedrals and round towers
// are never adjacent to each other, except a castle next to a cathedral;
// harbors and wind collectors are at the
// coast, harbors only on settlement spaces (at most one per settlement),
// wind collectors never on settlement spaces.
func (b *Board) validate() error {
	var errs []error
	for _, h := range b.sortedHexes() {
		t := b.Tiles[h]
		if t.Bonus != NoBonus && !b.Coast(h) {
			errs = append(errs, fmt.Errorf("%v at %s is not at the coast", t.Bonus, Pos(h)))
		}
		if (t.Bonus == Harbor) != (t.Bonus != NoBonus && t.Kind == Settlement) {
			errs = append(errs, fmt.Errorf("%v at %s not allowed on this space", t.Bonus, Pos(h)))
		}
		if t.Kind.Buildable() {
			continue
		}
		for _, n := range h.Neighbors() {
			// report each pair once, from the earlier space
			if nt := b.Tiles[n]; nt != nil && !nt.Kind.Buildable() && !castleAndCathedral(t.Kind, nt.Kind) && less(h, n) {
				errs = append(errs, fmt.Errorf("%v at %s and %v at %s are adjacent", t.Kind, Pos(h), nt.Kind, Pos(n)))
			}
		}
	}
	for _, a := range b.Settlements {
		first := slices.MinFunc(a.Spaces, cmpHex)
		if a.Size() > MaxSettlementSize {
			errs = append(errs, fmt.Errorf("settlement at %s has %d spaces, max %d", Pos(first), a.Size(), MaxSettlementSize))
		}
		bonus := 0
		for _, h := range a.Spaces {
			if b.Tiles[h].Bonus == Harbor {
				bonus++
			}
		}
		if bonus > 1 {
			errs = append(errs, fmt.Errorf("settlement at %s has %d harbors, max 1", Pos(first), bonus))
		}
	}
	return errors.Join(errs...)
}

// castleAndCathedral reports whether a and b are a castle and a cathedral,
// the only special spaces that may be adjacent.
func castleAndCathedral(a, b Kind) bool {
	return a == Castle && b == Cathedral || a == Cathedral && b == Castle
}

// Pos formats h as its row and column in a map file (both counted from 0,
// comment lines not counted).
func Pos(h hex.Hex) string {
	col, row := h.ToOffset()
	return fmt.Sprintf("row %d, col %d", row, col)
}

// cmpHex orders spaces like a map file: by row, then column.
func cmpHex(a, b hex.Hex) int {
	ac, ar := a.ToOffset()
	bc, br := b.ToOffset()
	if ar != br {
		return ar - br
	}
	return ac - bc
}

func less(a, b hex.Hex) bool { return cmpHex(a, b) < 0 }

func (b *Board) sortedHexes() []hex.Hex {
	hs := slices.Collect(maps.Keys(b.Tiles))
	slices.SortFunc(hs, cmpHex)
	return hs
}

// Clone returns a copy of b that can be played on without changing b,
// e.g. to try out a placement. Settlement areas and labels are shared; they
// never change. The clone has no mission deck: a trial placement must not
// reveal the hidden top card, so it draws none.
func (b *Board) Clone() *Board {
	c := *b
	c.Tiles = make(map[hex.Hex]*Tile, len(b.Tiles))
	for h, t := range b.Tiles {
		tc := *t
		c.Tiles[h] = &tc
	}
	c.points = maps.Clone(b.points)
	c.castles = maps.Clone(b.castles)
	c.lastGains = slices.Clone(b.lastGains)
	c.deck = nil
	c.missions = maps.Clone(b.missions)     // the slices are copied on write
	c.cathedrals = maps.Clone(b.cathedrals) // the slices are copied on write
	return &c
}

// Coast reports whether h is a space next to water (fewer than 6 neighbors).
func (b *Board) Coast(h hex.Hex) bool { return b.Tiles[h] != nil && landNeighbors(b, h) < 6 }

// groupSettlements assigns each connected group of settlement spaces to one
// SettlementArea (flood fill).
func (b *Board) groupSettlements() {
	for start, t := range b.Tiles {
		if t.Kind != Settlement || t.Area != nil {
			continue
		}
		a := &SettlementArea{}
		t.Area = a
		for stack := []hex.Hex{start}; len(stack) > 0; {
			h := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			a.Spaces = append(a.Spaces, h)
			for _, n := range h.Neighbors() {
				if nt := b.Tiles[n]; nt != nil && nt.Kind == Settlement && nt.Area == nil {
					nt.Area = a
					stack = append(stack, n)
				}
			}
		}
		b.Settlements = append(b.Settlements, a)
	}
}

// Radius creates a hexagon-shaped board with radius n (empty spaces only).
func Radius(n int) *Board {
	b := &Board{Name: fmt.Sprintf("radius-%d", n), Tiles: map[hex.Hex]*Tile{}}
	for q := -n; q <= n; q++ {
		for r := -n; r <= n; r++ {
			if h := (hex.Hex{Q: q, R: r}); abs(h.S()) <= n {
				b.Tiles[h] = &Tile{Kind: Empty}
			}
		}
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
