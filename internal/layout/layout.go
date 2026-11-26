// Package layout rechnet zwischen Hex-Koordinaten und Pixeln um (pointy-top).
package layout

import (
	"math"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Layout beschreibt Größe und Ursprung des Gitters in Pixeln.
type Layout struct {
	Size             float64 // Radius eines Sechsecks
	OriginX, OriginY float64 // Pixelposition von Hex{0,0}
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

// Corners liefert die 6 Eckpunkte (Winkel 30° + 60°·i).
func (l Layout) Corners(h hex.Hex) [6][2]float64 {
	cx, cy := l.HexToPixel(h)
	var c [6][2]float64
	for i := range c {
		a := math.Pi / 180 * float64(30+60*i)
		c[i] = [2]float64{cx + l.Size*math.Cos(a), cy + l.Size*math.Sin(a)}
	}
	return c
}
