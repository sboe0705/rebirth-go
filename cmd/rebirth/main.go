package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
	"github.com/sboe0705/rebirth-go/internal/layout"
	"github.com/sboe0705/rebirth-go/internal/render"
)

const screenW, screenH = 1024, 768

type game struct {
	board    *board.Board
	layout   layout.Layout
	selected board.Facility // facility placed by the next left click
	hover    hex.Hex
	msg      string
}

func (g *game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyF):
		g.selected = board.FoodFacility
	case inpututil.IsKeyJustPressed(ebiten.KeyE):
		g.selected = board.EnergyFacility
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		if g.selected == board.FoodFacility {
			g.selected = board.EnergyFacility
		} else {
			g.selected = board.FoodFacility
		}
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
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Selected: %v  (F / E or right click to switch, left click to place)\n%s", g.selected, g.msg))
}

func (g *game) Layout(_, _ int) (int, int) { return screenW, screenH }

func main() {
	mapName := flag.String("map", "example", "name of the embedded map (internal/board/maps/<name>.txt)")
	flag.Parse()

	b, err := board.Load(*mapName)
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("rebirth-go – " + b.Name)
	g := &game{board: b, layout: render.FitLayout(b, screenW, screenH, 20), selected: board.FoodFacility}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
