// Package board enthält das Spielfeld-Datenmodell und den Karten-Loader.
package board

import (
	"bufio"
	"embed"
	"fmt"
	"strings"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Terrain ist der Geländetyp eines Feldes.
type Terrain int

const (
	Grass Terrain = iota
	Forest
	Highland
	Lake
	Settlement
)

// terrainChars bildet Kartenzeichen auf Geländetypen ab.
var terrainChars = map[rune]Terrain{
	'G': Grass, 'F': Forest, 'H': Highland, 'W': Lake, 'S': Settlement,
}

type Tile struct {
	Terrain Terrain
}

// Board ist ein beliebig geformtes Hex-Spielfeld. Fehlender Eintrag = kein Feld.
type Board struct {
	Name  string
	Tiles map[hex.Hex]*Tile
}

//go:embed maps/*.txt
var mapFS embed.FS

// Load lädt eine eingebettete Karte, z. B. Load("example").
func Load(name string) (*Board, error) {
	data, err := mapFS.ReadFile("maps/" + name + ".txt")
	if err != nil {
		return nil, err
	}
	return Parse(name, string(data))
}

// Parse liest eine Karte im Textformat: jede Zeile ist eine Hex-Reihe
// (odd-r-Offset: ungerade Zeilen sind um ein halbes Feld nach rechts versetzt),
// jedes Zeichen eine Spalte. '.' oder ' ' = kein Feld, '#' am Zeilenanfang = Kommentar.
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
				return nil, fmt.Errorf("%s: unbekanntes Zeichen %q in Zeile %d, Spalte %d", name, c, row, col)
			}
			b.Tiles[hex.FromOffset(col, row)] = &Tile{Terrain: t}
		}
		row++
	}
	return b, sc.Err()
}

// Radius erzeugt ein sechseckiges Feld mit Radius n (nur Grasland).
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
