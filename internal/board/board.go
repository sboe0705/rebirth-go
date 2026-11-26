// Package board contains the board data model and the map loader.
package board

import (
	"bufio"
	"embed"
	"fmt"
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

	Piece  Piece // placed tile, NoPiece if none
	Houses int   // settlement symbols on a placed settlement tile
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

	points    int    // points scored while placing tiles, see Score
	lastGains []Gain // how the most recent placement scored
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

// Parse reads a map in text format: each line is a hex row
// (odd-r offset: odd rows are shifted right by half a tile),
// each character a column. '.' or ' ' = no tile, '#' at line start = comment.
// Adjacent settlement spaces form one settlement.
func Parse(name, text string) (*Board, error) {
	b := &Board{Name: name, Tiles: map[hex.Hex]*Tile{}}
	sc := bufio.NewScanner(strings.NewReader(text))
	row := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
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
	if err := b.groupSettlements(); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if err := b.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return b, nil
}

// validate checks the map rules: castles, cathedrals and round towers are
// never adjacent to each other; harbors and wind collectors are at the coast,
// harbors only on settlement spaces (at most one per settlement), wind
// collectors never on settlement spaces.
func (b *Board) validate() error {
	for h, t := range b.Tiles {
		if t.Bonus != NoBonus && !b.Coast(h) {
			return fmt.Errorf("%v at %v is not at the coast", t.Bonus, h)
		}
		if (t.Bonus == Harbor) != (t.Bonus != NoBonus && t.Kind == Settlement) {
			return fmt.Errorf("%v at %v not allowed on this space", t.Bonus, h)
		}
		if t.Kind.Buildable() {
			continue
		}
		for _, n := range h.Neighbors() {
			if nt := b.Tiles[n]; nt != nil && !nt.Kind.Buildable() {
				return fmt.Errorf("special spaces %v and %v are adjacent", h, n)
			}
		}
	}
	for _, a := range b.Settlements {
		bonus := 0
		for _, h := range a.Spaces {
			if b.Tiles[h].Bonus == Harbor {
				bonus++
			}
		}
		if bonus > 1 {
			return fmt.Errorf("settlement at %v has %d harbors, max 1", a.Spaces[0], bonus)
		}
	}
	return nil
}

// Coast reports whether h is a space next to water (fewer than 6 neighbors).
func (b *Board) Coast(h hex.Hex) bool { return b.Tiles[h] != nil && landNeighbors(b, h) < 6 }

// groupSettlements assigns each connected group of settlement spaces to one
// SettlementArea (flood fill).
func (b *Board) groupSettlements() error {
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
		if a.Size() > MaxSettlementSize {
			return fmt.Errorf("settlement at %v has %d spaces, max %d", start, a.Size(), MaxSettlementSize)
		}
		b.Settlements = append(b.Settlements, a)
	}
	return nil
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
