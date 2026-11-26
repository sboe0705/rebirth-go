// Package render draws a board with Ebitengine. No game logic.
package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
	"github.com/sboe0705/rebirth-go/internal/layout"
)

var (
	Background = color.RGBA{0x2b, 0x5d, 0x8a, 0xff} // sea
	borderCol  = color.RGBA{0x20, 0x20, 0x20, 0xff}
	bonusCol   = color.RGBA{0xff, 0xff, 0xff, 0xff} // "+1" marker
	frameCol   = color.RGBA{0x5a, 0x2e, 0x12, 0xff} // settlement frame lines
	kindCol    = map[board.Kind]color.RGBA{
		board.Empty:      {0xc8, 0xd6, 0x9a, 0xff},
		board.Food:       {0x6f, 0xb0, 0x4a, 0xff},
		board.Energy:     {0xe8, 0xc5, 0x3a, 0xff},
		board.Settlement: {0xc0, 0x84, 0x5a, 0xff},
		board.Castle:     {0x70, 0x70, 0x78, 0xff},
		board.Cathedral:  {0x8a, 0x5a, 0xa8, 0xff},
		board.RoundTower: {0xd8, 0xa0, 0x30, 0xff},
	}
)

// FitLayout computes a layout that fits the board centered into w×h pixels.
func FitLayout(b *board.Board, w, h int, margin float64) layout.Layout {
	unit := layout.Layout{Size: 1}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for hx := range b.Tiles {
		for _, c := range unit.Corners(hx) {
			minX, maxX = math.Min(minX, c[0]), math.Max(maxX, c[0])
			minY, maxY = math.Min(minY, c[1]), math.Max(maxY, c[1])
		}
	}
	size := math.Min((float64(w)-2*margin)/(maxX-minX), (float64(h)-2*margin)/(maxY-minY))
	return layout.Layout{
		Size:    size,
		OriginX: float64(w)/2 - size*(minX+maxX)/2,
		OriginY: float64(h)/2 - size*(minY+maxY)/2,
	}
}

// DrawBoard draws all tiles as filled hexagons with a border.
func DrawBoard(dst *ebiten.Image, b *board.Board, l layout.Layout) {
	for h, t := range b.Tiles {
		drawHex(dst, l, h, kindCol[t.Kind])
		if t.Bonus {
			x, y := l.HexToPixel(h)
			vector.DrawFilledCircle(dst, float32(x), float32(y), float32(l.Size/4), bonusCol, true)
		}
		if t.Area != nil {
			drawFrame(dst, l, h, t.Area)
		}
	}
}

// drawFrame draws the settlement frame on the outer edges of h: one line per
// settlement space (single frame = small, double = medium, ...).
func drawFrame(dst *ebiten.Image, l layout.Layout, h hex.Hex, a *board.SettlementArea) {
	w := float32(math.Max(1, l.Size/20))
	for dir := 0; dir < 6; dir++ {
		if a.Contains(h.Neighbor(dir)) {
			continue
		}
		for i := 0; i < a.Size(); i++ {
			p, q := l.Edge(h, dir, 0.88-0.1*float64(i))
			vector.StrokeLine(dst, float32(p[0]), float32(p[1]), float32(q[0]), float32(q[1]), w, frameCol, true)
		}
	}
}

func drawHex(dst *ebiten.Image, l layout.Layout, h hex.Hex, fill color.Color) {
	var p vector.Path
	for i, c := range l.Corners(h) {
		if i == 0 {
			p.MoveTo(float32(c[0]), float32(c[1]))
		} else {
			p.LineTo(float32(c[0]), float32(c[1]))
		}
	}
	p.Close()

	fo := &vector.FillOptions{}
	do := &vector.DrawPathOptions{AntiAlias: true}
	do.ColorScale.ScaleWithColor(fill)
	vector.FillPath(dst, &p, fo, do)

	so := &vector.StrokeOptions{Width: float32(math.Max(1, l.Size/15)), LineJoin: vector.LineJoinRound}
	do = &vector.DrawPathOptions{AntiAlias: true}
	do.ColorScale.ScaleWithColor(borderCol)
	vector.StrokePath(dst, &p, so, do)
}
