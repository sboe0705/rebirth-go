// Package board contains the board data model and the map loader.
package board

import (
	"bufio"
	"embed"
	"fmt"
	"strings"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Terrain is the terrain type of a tile.
type Terrain int

const (
	Grass Terrain = iota
	Forest
	Highland
	Lake
	Settlement
)

// terrainChars maps map characters to terrain types.
var terrainChars = map[rune]Terrain{
	'G': Grass, 'F': Forest, 'H': Highland, 'W': Lake, 'S': Settlement,
}

type Tile struct {
	Terrain Terrain
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
			t, ok := terrainChars[c]
			if !ok {
				return nil, fmt.Errorf("%s: unknown character %q in row %d, column %d", name, c, row, col)
			}
			b.Tiles[hex.FromOffset(col, row)] = &Tile{Terrain: t}
		}
		row++
	}
	return b, sc.Err()
}

// Radius creates a hexagon-shaped board with radius n (grass only).
func Radius(n int) *Board {
	b := &Board{Name: fmt.Sprintf("radius-%d", n), Tiles: map[hex.Hex]*Tile{}}
	for q := -n; q <= n; q++ {
		for r := -n; r <= n; r++ {
			if h := (hex.Hex{Q: q, R: r}); abs(h.S()) <= n {
				b.Tiles[h] = &Tile{Terrain: Grass}
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
