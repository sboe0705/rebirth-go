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

// kindChars maps map characters to space kinds. Lowercase letters of
// buildable kinds additionally mark a "+1" space (harbour / wind collector).
var kindChars = map[rune]Kind{
	'O': Empty, 'F': Food, 'E': Energy, 'S': Settlement,
	'C': Castle, 'K': Cathedral, 'R': RoundTower,
}

// Tile is a board space. Water is not modelled: no entry means no space.
type Tile struct {
	Kind  Kind
	Bonus bool // "+1" space: harbour (Scotland) or wind collector (Ireland)
}

func parseChar(c rune) (*Tile, bool) {
	if k, ok := kindChars[c]; ok {
		return &Tile{Kind: k}, true
	}
	if k, ok := kindChars[unicode.ToUpper(c)]; ok && unicode.IsLower(c) && k.Buildable() {
		return &Tile{Kind: k, Bonus: true}, true
	}
	return nil, false
}

// Board is an arbitrarily shaped hex board. A missing entry means no tile.
type Board struct {
	Name  string
	Tiles map[hex.Hex]*Tile
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
	return b, sc.Err()
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
