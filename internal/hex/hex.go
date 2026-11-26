// Package hex implements axial hex coordinates (pointy-top, q to the right, r to the lower right).
// Referenz: Red Blob Games, "Hexagonal Grids".
package hex

// Hex is an axial coordinate. s = -q - r is implicit.
type Hex struct{ Q, R int }

// S returns the third cube component.
func (h Hex) S() int { return -h.Q - h.R }

// Directions are the 6 neighbor directions.
var Directions = [6]Hex{{+1, 0}, {+1, -1}, {0, -1}, {-1, 0}, {-1, +1}, {0, +1}}

func (h Hex) Add(o Hex) Hex { return Hex{h.Q + o.Q, h.R + o.R} }

// Neighbor returns the neighbor in direction dir (0..5).
func (h Hex) Neighbor(dir int) Hex { return h.Add(Directions[dir]) }

// Neighbors returns all 6 neighbor coordinates (whether or not a tile exists there).
func (h Hex) Neighbors() [6]Hex {
	var n [6]Hex
	for i, d := range Directions {
		n[i] = h.Add(d)
	}
	return n
}

// Distance returns the hex distance.
func Distance(a, b Hex) int {
	dq, dr := a.Q-b.Q, a.R-b.R
	return (abs(dq) + abs(dr) + abs(dq+dr)) / 2
}

// Round rounds fractional axial coordinates using cube rounding.
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

// FromOffset converts "odd-r" offset coordinates (column, row; odd rows
// indented to the right) to axial coordinates.
func FromOffset(col, row int) Hex {
	return Hex{col - (row-(row&1))/2, row}
}

// ToOffset converts h to "odd-r" offset coordinates (column, row).
func (h Hex) ToOffset() (col, row int) {
	return h.Q + (h.R-(h.R&1))/2, h.R
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
