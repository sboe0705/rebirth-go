package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/sboe0705/rebirth-go/internal/board"
)

// Icons are simple own shapes in unit coordinates (-1..1, y down).
type pt = [2]float32

var (
	boltIcon   = []pt{{0.15, -0.9}, {-0.45, 0.1}, {-0.05, 0.1}, {-0.2, 0.9}, {0.45, -0.15}, {0.05, -0.15}}
	houseIcon  = []pt{{-0.6, 0.6}, {-0.6, -0.05}, {0, -0.65}, {0.6, -0.05}, {0.6, 0.6}}
	castleIcon = []pt{{-0.7, 0.7}, {-0.7, -0.6}, {-0.45, -0.6}, {-0.45, -0.35}, {-0.15, -0.35}, {-0.15, -0.6},
		{0.15, -0.6}, {0.15, -0.35}, {0.45, -0.35}, {0.45, -0.6}, {0.7, -0.6}, {0.7, 0.7}}
	crossV    = []pt{{-0.15, -0.85}, {0.15, -0.85}, {0.15, 0.85}, {-0.15, 0.85}}
	crossH    = []pt{{-0.55, -0.45}, {0.55, -0.45}, {0.55, -0.15}, {-0.55, -0.15}}
	towerIcon = []pt{{-0.25, 0.85}, {-0.25, -0.3}, {0, -0.9}, {0.25, -0.3}, {0.25, 0.85}}
	plusV     = []pt{{-0.3, -1}, {0.3, -1}, {0.3, 1}, {-0.3, 1}}
	plusH     = []pt{{-1, -0.3}, {1, -0.3}, {1, 0.3}, {-1, 0.3}}
)

// drawKindIcon draws the symbol of a space kind at (x, y) with half-size s.
func drawKindIcon(dst *ebiten.Image, k board.Kind, x, y, s float32, c color.Color) {
	switch k {
	case board.Food:
		drawLeaf(dst, x, y, s, c)
	case board.Energy:
		fillPoly(dst, boltIcon, x, y, s, c)
	case board.Settlement:
		fillPoly(dst, houseIcon, x, y, s, c)
	case board.Castle:
		fillPoly(dst, castleIcon, x, y, s, c)
	case board.Cathedral:
		fillPoly(dst, crossV, x, y, s, c)
		fillPoly(dst, crossH, x, y, s, c)
	case board.RoundTower:
		fillPoly(dst, towerIcon, x, y, s, c)
	}
}

func drawPieceIcon(dst *ebiten.Image, f board.Piece, x, y, s float32, c color.Color) {
	switch f {
	case board.FoodFacility:
		drawLeaf(dst, x, y, s, c)
	case board.EnergyFacility:
		fillPoly(dst, boltIcon, x, y, s, c)
	case board.SettlementTile:
		fillPoly(dst, houseIcon, x, y, s, c)
	}
}

// drawPlus draws the "+1" marker as a plus sign.
func drawPlus(dst *ebiten.Image, x, y, s float32, c color.Color) {
	fillPoly(dst, plusV, x, y, s, c)
	fillPoly(dst, plusH, x, y, s, c)
}

func drawLeaf(dst *ebiten.Image, x, y, s float32, c color.Color) {
	var p vector.Path
	p.MoveTo(x-0.45*s, y+0.55*s)
	p.QuadTo(x+0.75*s, y+0.45*s, x+0.6*s, y-0.8*s)
	p.QuadTo(x-0.65*s, y-0.7*s, x-0.45*s, y+0.55*s)
	p.Close()
	fill(dst, &p, c)
	var stem vector.Path
	stem.MoveTo(x-0.4*s, y+0.5*s)
	stem.LineTo(x-0.75*s, y+0.85*s)
	so := &vector.StrokeOptions{Width: 0.15 * s, LineCap: vector.LineCapRound}
	do := &vector.DrawPathOptions{AntiAlias: true}
	do.ColorScale.ScaleWithColor(c)
	vector.StrokePath(dst, &stem, so, do)
}

func fillPoly(dst *ebiten.Image, pts []pt, x, y, s float32, c color.Color) {
	var p vector.Path
	for i, q := range pts {
		if i == 0 {
			p.MoveTo(x+q[0]*s, y+q[1]*s)
		} else {
			p.LineTo(x+q[0]*s, y+q[1]*s)
		}
	}
	p.Close()
	fill(dst, &p, c)
}

func fill(dst *ebiten.Image, p *vector.Path, c color.Color) {
	do := &vector.DrawPathOptions{AntiAlias: true}
	do.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, p, &vector.FillOptions{}, do)
}

// darker returns c with RGB scaled to 55 %.
func darker(c color.RGBA) color.RGBA {
	return color.RGBA{uint8(int(c.R) * 55 / 100), uint8(int(c.G) * 55 / 100), uint8(int(c.B) * 55 / 100), c.A}
}
