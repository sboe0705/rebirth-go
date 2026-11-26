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
	board  *board.Board
	layout layout.Layout
	w, h   int // current screen size; the layout is refitted when it changes
	supply *board.Supply
	hand   board.Chip // tile placed by the next left click
	over   bool       // supply empty, game ended
	hover  hex.Hex
	msg    string
}

// draw takes the next tile from the supply into the hand. Tiles that fit
// nowhere on the board are discarded (the rules don't cover this case).
func (g *game) draw() {
	for {
		c, ok := g.supply.Draw()
		if !ok {
			g.over, g.hand = true, board.Chip{}
			return
		}
		if g.board.CanPlaceAnywhere(c.Piece) {
			g.hand = c
			return
		}
		g.msg = fmt.Sprintf("drawn %v fits nowhere, discarded", c)
	}
}

func (g *game) Update() error {
	x, y := ebiten.CursorPosition()
	g.hover = g.layout.PixelToHex(float64(x), float64(y))
	if g.over || !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || g.board.Tiles[g.hover] == nil {
		return nil
	}
	if err := g.board.PlaceChip(g.hover, g.hand); err != nil {
		g.msg = fmt.Sprintf("cannot place %v: %v", g.hand, err)
		return nil
	}
	g.msg = ""
	g.draw()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(render.Background)
	render.DrawBoard(screen, g.board, g.layout)
	if !g.over && g.board.Tiles[g.hover] != nil {
		render.DrawHover(screen, g.layout, g.hover, g.hand.Piece, g.board.CanPlace(g.hover, g.hand.Piece) == nil)
	}
	sc := g.board.Score()
	head := fmt.Sprintf("Next tile: %v  (left click to place; %d left in supply)", g.hand, g.supply.Len())
	pts := "Points"
	if g.over {
		head, pts = "Game over: all tiles placed", "Final score"
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"%s\n%s: %d  (placed %d, castles %d, open settlements %d)\n%s\n%s",
		head, pts, sc.Total(), sc.Placed, sc.Castles, sc.OpenSettlements, lastPlacement(g.board.LastGains()), g.msg))
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
	seed := flag.Uint64("seed", 0, "seed for the random map and the tile order (0 = new random seed)")
	flag.Parse()

	if *seed == 0 {
		*seed = rand.Uint64()
	}
	var b *board.Board
	var err error
	if *mapName == "random" {
		b, err = board.Random(*seed, board.DefaultRandom)
	} else {
		b, err = board.Load(*mapName)
	}
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	title := "rebirth-go – " + b.Name
	if *mapName != "random" {
		title += fmt.Sprintf(" (seed %d)", *seed)
	}
	ebiten.SetWindowTitle(title)
	g := &game{board: b, supply: board.NewSupply(*seed)}
	g.draw()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
