package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
	"github.com/sboe0705/rebirth-go/internal/layout"
	"github.com/sboe0705/rebirth-go/internal/render"
)

const screenW, screenH = 1024, 768 // initial window size

type game struct {
	board    *board.Board
	layout   layout.Layout
	w, h     int         // current screen size; the layout is refitted when it changes
	selected board.Piece // piece placed by the next left click
	hover    hex.Hex
	msg      string
}

func (g *game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyF):
		g.selected = board.FoodFacility
	case inpututil.IsKeyJustPressed(ebiten.KeyE):
		g.selected = board.EnergyFacility
	case inpututil.IsKeyJustPressed(ebiten.KeyS):
		g.selected = board.SettlementTile
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		g.selected = g.selected%board.SettlementTile + 1 // cycle food -> energy -> settlement
	}
	x, y := ebiten.CursorPosition()
	g.hover = g.layout.PixelToHex(float64(x), float64(y))
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.board.Tiles[g.hover] != nil {
		if err := g.board.Place(g.hover, g.selected); err != nil {
			g.msg = fmt.Sprintf("cannot place %v: %v", g.selected, err)
		} else {
			g.msg = ""
		}
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(render.Background)
	render.DrawBoard(screen, g.board, g.layout)
	if g.board.Tiles[g.hover] != nil {
		render.DrawHover(screen, g.layout, g.hover, g.selected, g.board.CanPlace(g.hover, g.selected) == nil)
	}
	sc := g.board.Score()
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"Selected: %v  (F / E / S or right click to switch, left click to place)\n"+
			"Points: %d  (placed %d, castles %d, open settlements %d)\n%s\n%s",
		g.selected, sc.Total(), sc.Placed, sc.Castles, sc.OpenSettlements, lastPlacement(g.board.LastGains()), g.msg))
}

// lastPlacement explains the points of the most recent placement,
// e.g. "Last placement: +9 = 1 ("+1" space) + 8 (medium settlement completed)".
func lastPlacement(gains []board.Gain) string {
	if gains == nil {
		return ""
	}
	sum, parts, notes := 0, []string{}, []string{}
	for _, g := range gains {
		if g.Points == 0 {
			notes = append(notes, g.Reason)
			continue
		}
		sum += g.Points
		parts = append(parts, fmt.Sprintf("%d (%s)", g.Points, g.Reason))
	}
	s := fmt.Sprintf("Last placement: +%d", sum)
	switch {
	case len(parts) == 1:
		s = fmt.Sprintf("Last placement: +%s", parts[0])
	case len(parts) > 1:
		s += " = " + strings.Join(parts, " + ")
	}
	if len(notes) > 0 {
		s += "; " + strings.Join(notes, "; ")
	}
	return s
}

// Layout uses the full window (resizable / maximizable) and refits the board.
func (g *game) Layout(w, h int) (int, int) {
	if w != g.w || h != g.h {
		g.w, g.h = w, h
		g.layout = render.FitLayout(g.board, w, h, 20)
	}
	return w, h
}

func main() {
	mapName := flag.String("map", "random", `"random" or name of an embedded map (internal/board/maps/<name>.txt)`)
	seed := flag.Uint64("seed", 0, "seed for the random map (0 = new random seed)")
	flag.Parse()

	var b *board.Board
	var err error
	if *mapName == "random" {
		if *seed == 0 {
			*seed = rand.Uint64()
		}
		b, err = board.Random(*seed, board.DefaultRandom)
	} else {
		b, err = board.Load(*mapName)
	}
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("rebirth-go – " + b.Name)
	g := &game{board: b, selected: board.FoodFacility}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
