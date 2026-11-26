// Package layout converts between hex coordinates and pixels (pointy-top).
package layout

import (
	"math"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Layout describes size and origin of the grid in pixels.
type Layout struct {
	Size             float64 // radius of a hexagon
	OriginX, OriginY float64 // pixel position of Hex{0,0}
}

func (l Layout) HexToPixel(h hex.Hex) (x, y float64) {
	q, r := float64(h.Q), float64(h.R)
	x = l.Size * math.Sqrt(3) * (q + r/2)
	y = l.Size * 1.5 * r
	return x + l.OriginX, y + l.OriginY
}

func (l Layout) PixelToHex(x, y float64) hex.Hex {
	x, y = x-l.OriginX, y-l.OriginY
	q := (math.Sqrt(3)/3*x - y/3) / l.Size
	r := (2.0 / 3 * y) / l.Size
	return hex.Round(q, r)
}

// Corners returns the 6 corner points (angle 30° + 60°·i).
func (l Layout) Corners(h hex.Hex) [6][2]float64 {
	cx, cy := l.HexToPixel(h)
	var c [6][2]float64
	for i := range c {
		a := math.Pi / 180 * float64(30+60*i)
		c[i] = [2]float64{cx + l.Size*math.Cos(a), cy + l.Size*math.Sin(a)}
	}
	return c
}

// Edge returns the two end points of the hexagon edge facing neighbor
// direction dir (see hex.Directions), with the corners scaled towards the
// hexagon center by scale (1 = outline).
func (l Layout) Edge(h hex.Hex, dir int, scale float64) (a, b [2]float64) {
	cx, cy := l.HexToPixel(h)
	c := l.Corners(h)
	p := func(i int) [2]float64 {
		return [2]float64{cx + scale*(c[i][0]-cx), cy + scale*(c[i][1]-cy)}
	}
	return p((5 - dir + 6) % 6), p((6 - dir) % 6)
}
