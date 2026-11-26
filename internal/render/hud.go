package render

import (
	"bytes"
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
	"github.com/sboe0705/rebirth-go/internal/layout"
)

var (
	regular = mustFont(goregular.TTF)
	bold    = mustFont(gobold.TTF)

	panelBg   = color.RGBA{0x0c, 0x18, 0x26, 0xd0} // premultiplied alpha
	panelEdge = color.RGBA{0x5a, 0x6a, 0x7a, 0xff}
	textCol   = color.RGBA{0xff, 0xff, 0xff, 0xff}
	dimCol    = color.RGBA{0xc8, 0xd2, 0xdc, 0xff}
)

func mustFont(ttf []byte) *text.GoTextFaceSource {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		panic(err)
	}
	return s
}

// Panel is what a player's corner panel shows.
type Panel struct {
	Player int
	Name   string
	Active bool // player to move
	Points int
	Label  string      // next to the points, e.g. "points"
	Detail []string    // small lines under the points
	Hand   *board.Chip // tile in hand, nil if not shown
}

// LogEntry is one move in the move log.
type LogEntry struct {
	Player int
	Chip   board.Chip
	Head   string   // e.g. "You +9"
	Lines  []string // reasons, wrapped to the log width
}

// DrawPanel draws player panel p into r with font size f.
func DrawPanel(dst *ebiten.Image, r layout.Rect, f float64, p Panel) {
	x, y, w, h := float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
	vector.FillRect(dst, x, y, w, h, panelBg, true)
	if p.Active {
		vector.StrokeRect(dst, x, y, w, h, float32(0.25*f), PlayerColors[p.Player], true)
	} else {
		vector.StrokeRect(dst, x, y, w, h, 1, panelEdge, true)
	}
	pad := 0.7 * f
	tx := r.X + pad
	if p.Active {
		drawTriangle(dst, tx, r.Y+pad+0.65*f, 0.4*f, PlayerColors[p.Player])
		tx += 1.0 * f
	}
	drawSwatch(dst, tx+0.5*f, r.Y+pad+0.65*f, 0.55*f, p.Player)
	drawText(dst, p.Name, bold, f, tx+1.3*f, r.Y+pad, textCol)

	py := r.Y + pad + 1.6*f
	pts := strconv.Itoa(p.Points)
	drawText(dst, pts, bold, 2*f, r.X+pad, py, textCol)
	pw := text.Advance(pts, &text.GoTextFace{Source: bold, Size: 2 * f})
	drawText(dst, p.Label, regular, f, r.X+pad+pw+0.4*f, py+0.8*f, dimCol)

	ly := py + 2.5*f
	for _, d := range p.Detail {
		drawText(dst, d, regular, 0.9*f, r.X+pad, ly, dimCol)
		ly += 1.25 * f
	}
	if p.Hand != nil {
		// the tile in hand at the top right, as on the board
		hr := 1.7 * f
		l := layout.Layout{Size: hr / pieceScale, OriginX: r.X + r.W - pad - hr, OriginY: r.Y + pad + hr}
		drawPiece(dst, l, hex.Hex{}, *p.Hand, p.Player)
	}
}

// DrawLog draws the move log (newest first) into r; entries that don't fit
// are left out, older ones are paler.
func DrawLog(dst *ebiten.Image, r layout.Rect, f float64, entries []LogEntry) {
	lh := 1.3 * f
	y := r.Y
	if len(entries) == 0 || r.H < 2*lh {
		return
	}
	drawText(dst, "Last moves", bold, f, r.X, y, dimCol)
	y += 1.6 * f
	for i, e := range entries {
		lines := []string{}
		for _, s := range e.Lines {
			lines = append(lines, wrap(s, regular, 0.9*f, r.W-1.6*f)...)
		}
		if y+lh*float64(1+len(lines)) > r.Y+r.H {
			return
		}
		a := 1 - 0.07*float32(i)
		l := layout.Layout{Size: 0.55 * f / pieceScale, OriginX: r.X + 0.55*f, OriginY: y + 0.65*f}
		drawPiece(dst, l, hex.Hex{}, e.Chip, e.Player)
		drawText(dst, e.Head, bold, f, r.X+1.6*f, y, fade(textCol, a))
		y += lh
		for _, s := range lines {
			drawText(dst, s, regular, 0.9*f, r.X+1.6*f, y, fade(dimCol, a))
			y += 1.1 * f
		}
		y += 0.35 * f
	}
}

// DrawBanner draws s on a dark background, horizontally centered at cx
// with its middle at cy, wrapped to at most maxW wide; alpha 0..1 fades it.
func DrawBanner(dst *ebiten.Image, s string, f, cx, cy, maxW float64, alpha float32) {
	pad, lh := 0.8*f, 1.3*f
	lines := wrap(s, bold, f, maxW-2*pad)
	face := &text.GoTextFace{Source: bold, Size: f}
	w := 0.0
	for _, l := range lines {
		w = max(w, text.Advance(l, face))
	}
	h := lh * float64(len(lines))
	top := cy - h/2
	vector.FillRect(dst, float32(cx-w/2-pad), float32(top-pad), float32(w+2*pad), float32(h+2*pad), fade(panelBg, alpha), true)
	for i, l := range lines {
		drawText(dst, l, bold, f, cx-text.Advance(l, face)/2, top+float64(i)*lh, fade(textCol, alpha))
	}
}

func drawText(dst *ebiten.Image, s string, src *text.GoTextFaceSource, size, x, y float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, &text.GoTextFace{Source: src, Size: size}, op)
}

// wrap splits s into lines no wider than width in font src at size f.
func wrap(s string, src *text.GoTextFaceSource, f, width float64) []string {
	face := &text.GoTextFace{Source: src, Size: f}
	var out []string
	line := ""
	for _, w := range strings.Fields(s) {
		try := w
		if line != "" {
			try = line + " " + w
		}
		if line != "" && text.Advance(try, face) > width {
			out, line = append(out, line), w
			continue
		}
		line = try
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

// drawSwatch draws a small hexagon in the colour of player p.
func drawSwatch(dst *ebiten.Image, x, y, r float64, p int) {
	drawHex(dst, layout.Layout{Size: r, OriginX: x, OriginY: y}, hex.Hex{}, 1, PlayerColors[p], borderCol)
}

// drawTriangle draws a right-pointing triangle centered at (x, y).
func drawTriangle(dst *ebiten.Image, x, y, r float64, c color.Color) {
	var p vector.Path
	p.MoveTo(float32(x-r*0.7), float32(y-r))
	p.LineTo(float32(x+r), float32(y))
	p.LineTo(float32(x-r*0.7), float32(y+r))
	p.Close()
	fill(dst, &p, c)
}

// fade scales a premultiplied colour by alpha.
func fade(c color.RGBA, a float32) color.RGBA {
	return color.RGBA{uint8(float32(c.R) * a), uint8(float32(c.G) * a), uint8(float32(c.B) * a), uint8(float32(c.A) * a)}
}
