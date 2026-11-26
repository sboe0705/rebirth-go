// Package board contains the board data model and the map loader.
package board

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
	"unicode"

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

// kindChars maps map characters to space kinds. Lowercase 's' marks a
// settlement space with "+1" (harbour / wind collector).
var kindChars = map[rune]Kind{
	'O': Empty, 'F': Food, 'E': Energy, 'S': Settlement,
	'C': Castle, 'K': Cathedral, 'R': RoundTower,
}

// Tile is a board space. Water is not modelled: no entry means no space.
type Tile struct {
	Kind  Kind
	Bonus bool            // "+1" settlement space: harbour (Scotland) or wind collector (Ireland)
	Area  *SettlementArea // framed settlement this space belongs to; nil unless Kind == Settlement
}

// MaxSettlementSize is the largest settlement: the extra-large one in Ireland.
const MaxSettlementSize = 4

// SettlementArea is one framed settlement: 1 space = small, 2 = medium,
// 3 = large, 4 = extra-large (Ireland). It is scored once all spaces are occupied.
type SettlementArea struct {
	Spaces []hex.Hex
}

// Size returns the number of spaces (= number of frame lines on the board).
func (a *SettlementArea) Size() int { return len(a.Spaces) }

// Contains reports whether h belongs to this settlement.
func (a *SettlementArea) Contains(h hex.Hex) bool {
	for _, s := range a.Spaces {
		if s == h {
			return true
		}
	}
	return false
}

func (a *SettlementArea) connected() bool {
	seen := map[hex.Hex]bool{a.Spaces[0]: true}
	stack := []hex.Hex{a.Spaces[0]}
	for len(stack) > 0 {
		h := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, n := range h.Neighbors() {
			if !seen[n] && a.Contains(n) {
				seen[n] = true
				stack = append(stack, n)
			}
		}
	}
	return len(seen) == len(a.Spaces)
}

func parseChar(c rune) (*Tile, bool) {
	if k, ok := kindChars[c]; ok {
		return &Tile{Kind: k}, true
	}
	if c == 's' {
		return &Tile{Kind: Settlement, Bonus: true}, true
	}
	return nil, false
}

// Board is an arbitrarily shaped hex board. A missing entry means no tile.
type Board struct {
	Name        string
	Tiles       map[hex.Hex]*Tile
	Settlements []*SettlementArea // in order of first appearance in the map
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
//
// A line "---" starts the settlement section: a second grid of the same
// shape in which every settlement space carries a label (letter or digit).
// Spaces with the same label form one settlement (SCO p. 3, IRL p. 3).
// Other cells may hold '.', ' ' or the map character itself.
func Parse(name, text string) (*Board, error) {
	b := &Board{Name: name, Tiles: map[hex.Hex]*Tile{}}
	var grids [2][]string
	part := 0
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "#"):
		case strings.TrimSpace(line) == "---":
			if part == 1 {
				return nil, fmt.Errorf("%s: more than one settlement section", name)
			}
			part = 1
		default:
			grids[part] = append(grids[part], line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	for row, line := range grids[0] {
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
	}
	if err := b.parseSettlements(grids[0], grids[1]); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return b, nil
}

func (b *Board) parseSettlements(mapRows, rows []string) error {
	byLabel := map[rune]*SettlementArea{}
	var order []rune
	for row, line := range rows {
		var mapLine []rune
		if row < len(mapRows) {
			mapLine = []rune(mapRows[row])
		}
		for col, c := range []rune(line) {
			h := hex.FromOffset(col, row)
			t := b.Tiles[h]
			if t == nil || t.Kind != Settlement {
				if c != '.' && c != ' ' && (col >= len(mapLine) || mapLine[col] != c) {
					return fmt.Errorf("settlement label %q in row %d, column %d is not on a settlement space", c, row, col)
				}
				continue
			}
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return fmt.Errorf("invalid settlement label %q in row %d, column %d", c, row, col)
			}
			a := byLabel[c]
			if a == nil {
				a = &SettlementArea{}
				byLabel[c] = a
				order = append(order, c)
			}
			a.Spaces = append(a.Spaces, h)
			t.Area = a
		}
	}
	for h, t := range b.Tiles {
		if t.Kind == Settlement && t.Area == nil {
			off := h.Q + (h.R-(h.R&1))/2
			return fmt.Errorf("settlement space in row %d, column %d has no label", h.R, off)
		}
	}
	for _, c := range order {
		a := byLabel[c]
		if n := len(a.Spaces); n > MaxSettlementSize {
			return fmt.Errorf("settlement %q has %d spaces, max %d", c, n, MaxSettlementSize)
		}
		if !a.connected() {
			return fmt.Errorf("settlement %q is not contiguous", c)
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
