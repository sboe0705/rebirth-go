package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
	"github.com/sboe0705/rebirth-go/internal/layout"
	"github.com/sboe0705/rebirth-go/internal/render"
)

const (
	screenW, screenH = 1024, 768 // initial window size
	maxPlayers       = 4
	computerDelay    = 40 // ticks (1/60 s) before a computer player places its tile
)

// player is the user (index 0) or a computer player placing at random.
type player struct {
	name     string
	computer bool
	supply   *board.Supply
	hand     board.Chip
	hasHand  bool
}

type game struct {
	board   *board.Board
	layout  layout.Layout
	w, h    int // current screen size; the layout is refitted when it changes
	seed    uint64
	rng     *rand.Rand // computer players' choices
	players []*player  // nil until the number of players is chosen
	turn    int        // index of the player to move
	wait    int        // ticks until the computer player to move places
	over    bool       // all supplies empty, game ended
	hover   hex.Hex
	last    string // explanation of the most recent placement
	msg     string
}

// start sets up n players: the user and n-1 computer players, each with an
// own shuffled supply, and gives the first turn to the user.
func (g *game) start(n int) {
	g.board.Players = n
	g.rng = rand.New(rand.NewPCG(g.seed, 0x9e3779b97f4a7c15))
	for i := range n {
		p := &player{name: "You", supply: board.NewSupplyFor(g.seed+uint64(i)*0x632be59bd9b4e019, n)}
		if i > 0 {
			p.name, p.computer = fmt.Sprintf("Computer %d", i), true
		}
		g.players = append(g.players, p)
	}
	g.turn = n - 1
	g.nextTurn()
}

// refill makes sure p holds a tile that fits on the board. Tiles that fit
// nowhere are discarded (the rules don't cover this case).
func (g *game) refill(p *player) {
	if p.hasHand && g.board.CanPlaceAnywhere(p.hand.Piece) {
		return
	}
	if p.hasHand {
		g.msg = fmt.Sprintf("%s: %v fits nowhere, discarded", p.name, p.hand)
	}
	for {
		c, ok := p.supply.Draw()
		if !ok {
			p.hasHand, p.hand = false, board.Chip{}
			return
		}
		if g.board.CanPlaceAnywhere(c.Piece) {
			p.hand, p.hasHand = c, true
			return
		}
		g.msg = fmt.Sprintf("%s: drawn %v fits nowhere, discarded", p.name, c)
	}
}

// nextTurn passes the turn clockwise to the next player who still has a
// tile; the game is over when nobody has one.
func (g *game) nextTurn() {
	n := len(g.players)
	for i := 1; i <= n; i++ {
		t := (g.turn + i) % n
		if g.refill(g.players[t]); g.players[t].hasHand {
			g.turn, g.wait = t, computerDelay
			return
		}
	}
	g.over = true
}

// place puts the current player's tile at h, draws a new one and ends the turn.
func (g *game) place(h hex.Hex) error {
	p := g.players[g.turn]
	if err := g.board.PlaceChip(h, p.hand, g.turn); err != nil {
		return err
	}
	g.last = g.lastPlacement(g.turn, g.board.LastGains())
	p.hasHand = false
	g.refill(p)
	g.nextTurn()
	return nil
}

func (g *game) Update() error {
	if g.players == nil {
		for n := 1; n <= maxPlayers; n++ {
			if inpututil.IsKeyJustPressed(ebiten.Key1 + ebiten.Key(n-1)) {
				g.start(n)
			}
		}
		return nil
	}
	x, y := ebiten.CursorPosition()
	g.hover = g.layout.PixelToHex(float64(x), float64(y))
	if g.over {
		return nil
	}
	p := g.players[g.turn]
	if p.computer {
		if g.wait--; g.wait <= 0 {
			free := g.board.FreeSpaces(p.hand.Piece)
			g.msg = ""
			if err := g.place(free[g.rng.IntN(len(free))]); err != nil {
				return err // refill guarantees a free space
			}
		}
		return nil
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || g.board.Tiles[g.hover] == nil {
		return nil
	}
	g.msg = ""
	if err := g.place(g.hover); err != nil {
		g.msg = fmt.Sprintf("cannot place %v: %v", p.hand, err)
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(render.Background)
	render.DrawBoard(screen, g.board, g.layout)
	if g.players == nil {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(
			"Choose the number of players: press 1-%d (you + computer players)", maxPlayers))
		return
	}
	p := g.players[g.turn]
	if !g.over && !p.computer && g.board.Tiles[g.hover] != nil {
		render.DrawHover(screen, g.layout, g.hover, p.hand.Piece, g.board.CanPlace(g.hover, p.hand.Piece) == nil)
	}
	head := fmt.Sprintf("%s is placing %v ...", p.name, p.hand)
	if !p.computer {
		head = fmt.Sprintf("Your tile: %v  (left click to place; %d left in supply)", p.hand, p.supply.Len())
	}
	pts := "Points"
	if g.over {
		head, pts = "Game over: all tiles placed. "+g.winner(), "Final score"
	}
	lines := []string{head}
	for i, q := range g.players {
		sc := g.board.Score(i)
		lines = append(lines, fmt.Sprintf("%-10s (%s): %s %3d  (placed %d, castles %d, open settlements %d; %d in supply)",
			q.name, render.PlayerColorNames[i], pts, sc.Total(), sc.Placed, sc.Castles, sc.OpenSettlements, q.supply.Len()))
	}
	ebitenutil.DebugPrint(screen, strings.Join(append(lines, g.last, g.msg), "\n"))
}

// winner names the player(s) with the most points. Tie-break by named
// castles is not modelled, so ties are shared.
func (g *game) winner() string {
	best, names := -1, []string{}
	for i, p := range g.players {
		switch t := g.board.Score(i).Total(); {
		case t > best:
			best, names = t, []string{p.name}
		case t == best:
			names = append(names, p.name)
		}
	}
	if len(g.players) == 1 {
		return ""
	}
	if len(names) > 1 {
		return "Tie: " + strings.Join(names, ", ")
	}
	if names[0] == "You" {
		return "You win!"
	}
	return names[0] + " wins."
}

// lastPlacement explains the points of the most recent placement by player
// p, e.g. "Last placement (You): +9 = 1 (harbor) + 8 (medium settlement completed)",
// followed by what other players got from it.
func (g *game) lastPlacement(p int, gains []board.Gain) string {
	if gains == nil {
		return ""
	}
	order := []int{p}
	for _, x := range gains {
		if !slices.Contains(order, x.Player) {
			order = append(order, x.Player)
		}
	}
	var out []string
	for _, q := range order {
		var mine []board.Gain
		for _, x := range gains {
			if x.Player == q {
				mine = append(mine, x)
			}
		}
		out = append(out, g.players[q].name+" "+explain(mine))
	}
	return "Last placement: " + strings.Join(out, "; ")
}

// explain sums gains, e.g. "+9 = 1 (harbor) + 8 (medium settlement completed)".
func explain(gains []board.Gain) string {
	sum, parts, notes := 0, []string{}, []string{}
	for _, g := range gains {
		if g.Points == 0 {
			notes = append(notes, g.Reason)
			continue
		}
		sum += g.Points
		parts = append(parts, fmt.Sprintf("%d (%s)", g.Points, g.Reason))
	}
	s := fmt.Sprintf("+%d", sum)
	switch {
	case len(parts) == 1:
		s = "+" + parts[0]
	case len(parts) > 1:
		s += " = " + strings.Join(parts, " + ")
	}
	if len(notes) > 0 {
		s += ", " + strings.Join(notes, ", ")
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
	seed := flag.Uint64("seed", 0, "seed for the random map, the tile order and the computer moves (0 = new random seed)")
	players := flag.Int("players", 0, fmt.Sprintf("number of players 1-%d: you + computer players (0 = choose at start)", maxPlayers))
	flag.Parse()

	if *players < 0 || *players > maxPlayers {
		log.Fatalf("-players must be between 0 and %d", maxPlayers)
	}
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
	g := &game{board: b, seed: *seed}
	if *players > 0 {
		g.start(*players)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
