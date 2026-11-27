package layout

import (
	"math"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Rect is an axis-aligned rectangle in pixels.
type Rect struct{ X, Y, W, H float64 }

// Overlaps reports whether r and o share any area.
func (r Rect) Overlaps(o Rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}

// HUD places the interface in a window: one panel per player in the
// corners (player 0 bottom-left, then clockwise), the move log in the right
// column between the two right panels, and the board left of that column.
type HUD struct {
	Font   float64 // font size in pixels
	Gap    float64 // distance between elements and to the window edge
	Panels [4]Rect // index = player
	Log    Rect
	Board  Rect // area for the board; it still has to avoid the left panels
	// Missions is the box with the user's mission cards at the bottom of the
	// right column (empty until WithMissions); the log ends above it.
	Missions Rect
}

// WithMissions returns u with a missions box for n card lines (at least 1)
// at the bottom of the log column; the log gets the rest. The box never
// takes more than half the column.
func (u HUD) WithMissions(n int) HUD {
	f := u.Font
	h := math.Min((3.0+1.2*float64(max(n, 1)))*f, u.Log.H/2)
	u.Missions = Rect{u.Log.X, u.Log.Y + u.Log.H - h, u.Log.W, h}
	u.Log.H = math.Max(0, u.Log.H-h-u.Gap)
	return u
}

// MissionRow returns the rectangle of card line i in the missions box.
func (u HUD) MissionRow(i int) Rect {
	f := u.Font
	return Rect{u.Missions.X, u.Missions.Y + 2.3*f + 1.2*f*float64(i), u.Missions.W, 1.2 * f}
}

// NewHUD computes the HUD for a w×h window; sizes scale with the height.
func NewHUD(w, h float64) HUD {
	f := math.Max(12, math.Min(20, h/45))
	g := 0.8 * f
	pw, ph := 16*f, 7.4*f
	right := w - g - pw
	return HUD{
		Font:     f,
		Gap:      g,
		Panels: [4]Rect{
			{g, h - g - ph, pw, ph},
			{g, g, pw, ph},
			{right, g, pw, ph},
			{right, h - g - ph, pw, ph},
		},
		Log:   Rect{right, 2*g + ph, pw, math.Max(0, h-4*g-2*ph)},
		Board: Rect{g, g, math.Max(0, right-2*g), math.Max(0, h-2*g)},
	}
}

// Fit returns the largest layout that shows all hexes centered in area
// without any hex touching one of the avoid rectangles.
func Fit(hexes []hex.Hex, area Rect, avoid []Rect) Layout {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	unit := Layout{Size: 1}
	for _, h := range hexes {
		for _, c := range unit.Corners(h) {
			minX, maxX = math.Min(minX, c[0]), math.Max(maxX, c[0])
			minY, maxY = math.Min(minY, c[1]), math.Max(maxY, c[1])
		}
	}
	if len(hexes) == 0 {
		return Layout{Size: 1}
	}
	at := func(size float64) Layout {
		return Layout{
			Size:    size,
			OriginX: area.X + area.W/2 - size*(minX+maxX)/2,
			OriginY: area.Y + area.H/2 - size*(minY+maxY)/2,
		}
	}
	free := func(l Layout) bool {
		hw := l.Size * math.Sqrt(3) / 2
		for _, h := range hexes {
			x, y := l.HexToPixel(h)
			b := Rect{x - hw, y - l.Size, 2 * hw, 2 * l.Size}
			for _, a := range avoid {
				if b.Overlaps(a) {
					return false
				}
			}
		}
		return true
	}
	hi := math.Min(area.W/(maxX-minX), area.H/(maxY-minY))
	if free(at(hi)) {
		return at(hi)
	}
	lo := 0.0 // binary search for the largest free size
	for range 30 {
		if mid := (lo + hi) / 2; free(at(mid)) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return at(lo)
}
