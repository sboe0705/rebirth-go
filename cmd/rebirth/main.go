package main

import (
	"flag"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
	"github.com/sboe0705/rebirth-go/internal/layout"
	"github.com/sboe0705/rebirth-go/internal/render"
)

const (
	screenW, screenH = 1024, 768 // initial window size
	maxPlayers       = 4
	computerDelay    = 40  // ticks (1/60 s) before a computer player places its tile
	msgTicks         = 240 // how long a message is shown (the last second fading)
	logLen           = 8   // moves shown in the log
)

// player is the user (index 0) or a computer player placing at random.
type player struct {
	name     string
	computer bool
	supply   *board.Supply
	hand     board.Chip
	hasHand  bool
}

// move is a placement for the move log.
type move struct {
	player int
	chip   board.Chip
	gains  []board.Gain
}

type game struct {
	board   *board.Board
	layout  layout.Layout
	hud     layout.HUD
	w, h    int // current screen size; the layout is refitted when it changes
	seed    uint64
	rng     *rand.Rand // computer players' choices
	players []*player  // nil until the number of players is chosen
	turn    int        // index of the player to move
	wait    int        // ticks until the computer player to move places
	over    bool       // all supplies empty, game ended
	hover   hex.Hex
	moves   []move // newest first, at most logLen
	msg     string
	msgT    int // ticks the message is still shown
}

// say shows message s for a few seconds.
func (g *game) say(s string) { g.msg, g.msgT = s, msgTicks }

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
		g.say(fmt.Sprintf("%s: %v fits nowhere, discarded", p.name, p.hand))
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
		g.say(fmt.Sprintf("%s: drawn %v fits nowhere, discarded", p.name, c))
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
	g.moves = append([]move{{g.turn, p.hand, g.board.LastGains()}}, g.moves...)
	g.moves = g.moves[:min(len(g.moves), logLen)]
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
	if g.msgT > 0 {
		g.msgT--
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
			if err := g.place(free[g.rng.IntN(len(free))]); err != nil {
				return err // refill guarantees a free space
			}
		}
		return nil
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || g.board.Tiles[g.hover] == nil {
		return nil
	}
	if err := g.place(g.hover); err != nil {
		g.say(fmt.Sprintf("Cannot place %v: %v", p.hand, err))
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(render.Background)
	render.DrawBoard(screen, g.board, g.layout)
	f := g.hud.Font
	if g.players == nil {
		render.DrawBanner(screen, fmt.Sprintf("Press 1-%d to choose the number of players (you + computer players)", maxPlayers),
			1.2*f, float64(g.w)/2, float64(g.h)/2, float64(g.w)-2*g.hud.Gap, 1)
		return
	}
	p := g.players[g.turn]
	if !g.over && !p.computer && g.board.Tiles[g.hover] != nil {
		render.DrawHover(screen, g.layout, g.hover, p.hand.Piece, g.board.CanPlace(g.hover, p.hand.Piece) == nil)
	}
	for i := range g.players {
		render.DrawPanel(screen, g.hud.Panels[i], f, g.panel(i))
	}
	render.DrawLog(screen, g.hud.Log, f, g.logEntries())
	// messages at the bottom, centered between the two bottom panels
	l, r := g.hud.Panels[0], g.hud.Panels[3]
	cx, cy, mw := (l.X+l.W+r.X)/2, l.Y+l.H-2*f, r.X-l.X-l.W-2*g.hud.Gap
	switch {
	case g.over:
		render.DrawBanner(screen, strings.TrimSpace("Game over. "+g.winner()), 1.2*f, cx, cy, mw, 1)
	case g.msgT > 0:
		render.DrawBanner(screen, g.msg, f, cx, cy, mw, min(1, float32(g.msgT)/60))
	}
}

// panel collects what the corner panel of player i shows.
func (g *game) panel(i int) render.Panel {
	q, sc := g.players[i], g.board.Score(i)
	pn := render.Panel{
		Player: i, Name: q.name, Active: !g.over && i == g.turn,
		Points: sc.Total(), Label: "points",
		Detail: []string{
			fmt.Sprintf("placed %d · castles %d · open %d", sc.Placed, sc.Castles, sc.OpenSettlements),
			fmt.Sprintf("%d tiles in supply", q.supply.Len()),
		},
	}
	if g.over {
		pn.Label = "final score"
	}
	if q.hasHand && !g.over && (!q.computer || pn.Active) {
		h := q.hand
		pn.Hand = &h
	}
	return pn
}

// logEntries describes the recent moves, e.g. head "You +9" with lines
// "+1 harbor", "+8 medium settlement completed", "Computer 1: castle lost".
func (g *game) logEntries() []render.LogEntry {
	var out []render.LogEntry
	for _, m := range g.moves {
		own, lines, others := 0, []string{}, []int{}
		for _, x := range m.gains {
			if x.Player == m.player {
				own += x.Points
				lines = append(lines, reason(x))
			} else if !slices.Contains(others, x.Player) {
				others = append(others, x.Player)
			}
		}
		for _, q := range others {
			var rs []string
			for _, x := range m.gains {
				if x.Player == q {
					rs = append(rs, reason(x))
				}
			}
			lines = append(lines, g.players[q].name+": "+strings.Join(rs, ", "))
		}
		out = append(out, render.LogEntry{
			Player: m.player, Chip: m.chip, Lines: lines,
			Head: fmt.Sprintf("%s %+d", g.players[m.player].name, own),
		})
	}
	return out
}

// reason formats one gain, e.g. "+8 medium settlement completed".
func reason(x board.Gain) string {
	if x.Points == 0 {
		return x.Reason
	}
	return fmt.Sprintf("%+d %s", x.Points, x.Reason)
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

// Layout uses the full window (resizable / maximizable) and refits the board.
func (g *game) Layout(w, h int) (int, int) {
	if w != g.w || h != g.h {
		g.w, g.h = w, h
		g.hud = layout.NewHUD(float64(w), float64(h))
		g.layout = layout.Fit(slices.Collect(maps.Keys(g.board.Tiles)), g.hud.Board, g.hud.Panels[:])
	}
	return w, h
}

func main() {
	mapName := flag.String("map", "random", `"random", name of an embedded map (internal/board/maps/<name>.txt) or path to a map file (contains "/" or ends in ".txt")`)
	seed := flag.Uint64("seed", 0, "seed for the random map, the tile order and the computer moves (0 = new random seed)")
	players := flag.Int("players", 0, fmt.Sprintf("number of players 1-%d: you + computer players (0 = choose at start)", maxPlayers))
	lax := flag.Bool("lax", false, "also load maps that break the map rules; the problems are printed as warnings")
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
	} else if strings.ContainsRune(*mapName, filepath.Separator) || strings.Contains(*mapName, "/") || strings.HasSuffix(*mapName, ".txt") {
		b, err = board.LoadFile(*mapName)
	} else {
		b, err = board.Load(*mapName)
	}
	if err != nil {
		if b == nil || !*lax {
			log.Fatal(err)
		}
		log.Printf("warning: the map breaks the map rules:\n%v", err)
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
