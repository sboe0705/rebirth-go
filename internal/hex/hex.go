// Package hex implementiert axiale Hex-Koordinaten (pointy-top, q nach rechts, r nach rechts unten).
// Referenz: Red Blob Games, "Hexagonal Grids".
package hex

// Hex ist eine axiale Koordinate. s = -q - r ergibt sich implizit.
type Hex struct{ Q, R int }

// S liefert die dritte Cube-Komponente.
func (h Hex) S() int { return -h.Q - h.R }

// Directions sind die 6 Nachbarrichtungen.
var Directions = [6]Hex{{+1, 0}, {+1, -1}, {0, -1}, {-1, 0}, {-1, +1}, {0, +1}}

func (h Hex) Add(o Hex) Hex { return Hex{h.Q + o.Q, h.R + o.R} }

// Neighbor liefert den Nachbarn in Richtung dir (0..5).
func (h Hex) Neighbor(dir int) Hex { return h.Add(Directions[dir]) }

// Neighbors liefert alle 6 Nachbarkoordinaten (unabhängig davon, ob dort ein Feld existiert).
func (h Hex) Neighbors() [6]Hex {
	var n [6]Hex
	for i, d := range Directions {
		n[i] = h.Add(d)
	}
	return n
}

// Distance liefert die Hex-Distanz.
func Distance(a, b Hex) int {
	dq, dr := a.Q-b.Q, a.R-b.R
	return (abs(dq) + abs(dr) + abs(dq+dr)) / 2
}

// Round rundet fraktionale axiale Koordinaten per Cube-Rounding.
func Round(fq, fr float64) Hex {
	fs := -fq - fr
	q, r, s := roundf(fq), roundf(fr), roundf(fs)
	dq, dr, ds := absf(q-fq), absf(r-fr), absf(s-fs)
	switch {
	case dq > dr && dq > ds:
		q = -r - s
	case dr > ds:
		r = -q - s
	}
	return Hex{int(q), int(r)}
}

// FromOffset wandelt "odd-r"-Offsetkoordinaten (Zeile, Spalte; ungerade Zeilen
// nach rechts eingerückt) in axiale Koordinaten um.
func FromOffset(col, row int) Hex {
	return Hex{col - (row-(row&1))/2, row}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func roundf(x float64) float64 {
	if x < 0 {
		return -float64(int(-x + 0.5))
	}
	return float64(int(x + 0.5))
}
