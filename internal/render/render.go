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
	Background       = color.RGBA{0x2b, 0x5d, 0x8a, 0xff} // sea
	borderCol        = color.RGBA{0x20, 0x20, 0x20, 0xff}
	bonusCol         = color.RGBA{0x40, 0x40, 0x40, 0xff} // "+1" marker
	settlementBorder = color.RGBA{0xff, 0xff, 0xff, 0xff}
	// kindCol are moderately saturated tones so the saturated player colours stand out.
	kindCol = map[board.Kind]color.RGBA{
		board.Empty:      {0xd6, 0xe0, 0xa8, 0xff},
		board.Food:       {0x96, 0xc8, 0x78, 0xff},
		board.Energy:     {0xec, 0xd5, 0x7a, 0xff},
		board.Settlement: {0xd4, 0xa6, 0x82, 0xff},
		board.Castle:     {0xa4, 0xa4, 0xae, 0xff},
		board.Cathedral:  {0xb8, 0x9a, 0xd0, 0xff},
		board.RoundTower: {0x8f, 0xc8, 0xc0, 0xff},
	}
	pieceCol = map[board.Piece]color.RGBA{
		board.FoodFacility:   {0x2e, 0x6b, 0x1f, 0xff},
		board.EnergyFacility: {0xd9, 0x82, 0x2b, 0xff},
		board.SettlementTile: {0xb0, 0x30, 0x2a, 0xff},
	}
	invalidCol  = color.RGBA{0xe0, 0x30, 0x30, 0xff}
	iconOnPiece = color.RGBA{0xff, 0xff, 0xff, 0xff}

	// PlayerColors are the player colours (index = player), shown as the
	// fill of placed tiles and as the icon of controlled castles.
	PlayerColors = []color.RGBA{
		{0x1d, 0x4e, 0xd8, 0xff}, // blue
		{0xdc, 0x26, 0x26, 0xff}, // red
		{0xf5, 0xb0, 0x00, 0xff}, // yellow
		{0x15, 0x80, 0x3d, 0xff}, // green
	}
	PlayerColorNames = []string{"blue", "red", "yellow", "green"}
)

// pieceScale is the size of a placed tile relative to its space.
const pieceScale = 0.7

// castleIconScale is the half-size of the castle icon relative to its space.
const castleIconScale = 0.6

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

// DrawBoard draws all tiles as filled hexagons with a border. Settlement
// spaces are drawn last so their edges stay on top.
func DrawBoard(dst *ebiten.Image, b *board.Board, l layout.Layout) {
	for _, settlements := range []bool{false, true} {
		for h, t := range b.Tiles {
			if (t.Kind == board.Settlement) != settlements {
				continue
			}
			border := borderCol
			if settlements {
				border = settlementBorder
			}
			drawHex(dst, l, h, 1, kindCol[t.Kind], border)
			x, y := l.HexToPixel(h)
			cx, cy, sz := float32(x), float32(y), float32(l.Size)
			switch {
			case t.Piece != board.NoPiece:
				// Placed tiles are filled with the owner's colour.
				pc := PlayerColors[t.Owner]
				drawHex(dst, l, h, pieceScale, pc, borderCol)
				ic := iconOnPiece
				if isLight(pc) {
					ic = borderCol // a white icon is hard to see on a light tile
				}
				if t.Piece == board.SettlementTile {
					drawHouses(dst, t.Houses, cx, cy, 0.35*sz, ic)
				} else {
					drawPieceIcon(dst, t.Piece, cx, cy, 0.35*sz, ic)
				}
			case t.Bonus != board.NoBonus:
				drawKindIcon(dst, t.Kind, cx, cy+0.15*sz, 0.4*sz, darker(kindCol[t.Kind]))
				drawPlus(dst, cx, cy-0.55*sz, 0.15*sz, bonusCol)
			default:
				is := 0.45 * sz
				if t.Kind == board.Castle {
					is = castleIconScale * sz
				}
				if p, ok := b.CastleOwner(h); ok && t.Kind == board.Castle {
					// Controlled castle: icon in the owner's colour with a black outline.
					drawKindIcon(dst, t.Kind, cx, cy, is, PlayerColors[p])
					outlinePoly(dst, castleIcon, cx, cy, is, borderCol, float32(lineWidth(l, 1)))
					break
				}
				drawKindIcon(dst, t.Kind, cx, cy, is, darker(kindCol[t.Kind]))
			}
		}
	}
}

// DrawHover marks the space under the cursor: a translucent preview of
// piece f if it can be placed there, otherwise a red outline.
func DrawHover(dst *ebiten.Image, l layout.Layout, h hex.Hex, f board.Piece, ok bool) {
	if !ok {
		drawHex(dst, l, h, 1, nil, invalidCol)
		return
	}
	c := pieceCol[f]
	c.R, c.G, c.B, c.A = c.R/2, c.G/2, c.B/2, 0x80 // premultiplied alpha
	drawHex(dst, l, h, pieceScale, c, borderCol)
}

// drawHex draws hexagon h scaled around its center; a nil fill draws only the border.
func drawHex(dst *ebiten.Image, l layout.Layout, h hex.Hex, scale float64, fill, border color.Color) {
	p := hexPath(l, h, scale)
	if fill != nil {
		do := &vector.DrawPathOptions{AntiAlias: true}
		do.ColorScale.ScaleWithColor(fill)
		vector.FillPath(dst, p, &vector.FillOptions{}, do)
	}
	strokePath(dst, l, p, border, 1)
}

// lineWidth is the pixel width of a line width times the normal width.
func lineWidth(l layout.Layout, width float64) float64 {
	return width * math.Max(1, l.Size/15)
}

func strokePath(dst *ebiten.Image, l layout.Layout, p *vector.Path, c color.Color, width float64) {
	so := &vector.StrokeOptions{Width: float32(lineWidth(l, width)), LineJoin: vector.LineJoinRound}
	do := &vector.DrawPathOptions{AntiAlias: true}
	do.ColorScale.ScaleWithColor(c)
	vector.StrokePath(dst, p, so, do)
}

func hexPath(l layout.Layout, h hex.Hex, scale float64) *vector.Path {
	cx, cy := l.HexToPixel(h)
	var p vector.Path
	for i, c := range l.Corners(h) {
		x, y := float32(cx+(c[0]-cx)*scale), float32(cy+(c[1]-cy)*scale)
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	return &p
}
