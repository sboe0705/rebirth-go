package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/render"
)

const screenW, screenH = 1024, 768

type game struct {
	board *board.Board
}

func (g *game) Update() error { return nil }

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(render.Background)
	render.DrawBoard(screen, g.board, render.FitLayout(g.board, screenW, screenH, 20))
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
	if err := ebiten.RunGame(&game{board: b}); err != nil {
		log.Fatal(err)
	}
}
