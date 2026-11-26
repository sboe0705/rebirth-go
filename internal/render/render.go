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
	terrainCol = map[board.Terrain]color.RGBA{
		board.Grass:      {0x8c, 0xc0, 0x5a, 0xff},
		board.Forest:     {0x3d, 0x7a, 0x3a, 0xff},
		board.Highland:   {0x9a, 0x86, 0x6b, 0xff},
		board.Lake:       {0x5d, 0x9c, 0xd6, 0xff},
		board.Settlement: {0xd9, 0xb3, 0x6c, 0xff},
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
		drawHex(dst, l, h, terrainCol[t.Terrain])
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
