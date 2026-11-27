package ai

import (
	"math/rand/v2"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
)

func parse(t *testing.T, text string) *board.Board {
	t.Helper()
	b, err := board.Parse("t", text)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNormalTakesMostPoints(t *testing.T) {
	// A food facility next to the two placed ones scores 3; col 4 or 5 score 1.
	b := parse(t, "OOOEOO")
	for _, col := range []int{0, 1} {
		if err := b.Place(hex.FromOffset(col, 0), board.FoodFacility); err != nil {
			t.Fatal(err)
		}
	}
	rng := rand.New(rand.NewPCG(1, 2))
	for range 20 {
		if h := Choose(b, board.Chip{Piece: board.FoodFacility}, 0, Normal, rng); h != hex.FromOffset(2, 0) {
			t.Fatalf("Choose = %v, want col 2", h)
		}
	}
}

func TestNormalCountsCastles(t *testing.T) {
	// Next to the castle the facility scores 1 + 5 castle points.
	b := parse(t, "OOOOC")
	rng := rand.New(rand.NewPCG(1, 2))
	if h := Choose(b, board.Chip{Piece: board.EnergyFacility}, 0, Normal, rng); h != hex.FromOffset(3, 0) {
		t.Fatalf("Choose = %v, want col 3", h)
	}
}

func TestNormalCountsMissions(t *testing.T) {
	// Player 0 holds "Most cathedrals" and has 1 cathedral, player 1 has 2.
	// A food facility scores 1 at col 4 and col 8, but col 4 reaches cathedral
	// K2 and ties for most cathedrals (+5).
	b := parse(t, "OKO.OKO.O")
	b.Players = 2
	b.SetDeck([]board.Mission{{Kind: board.MostCathedrals}, {Kind: board.CathedralBonus}, {Kind: board.CathedralBonus}})
	for _, m := range []struct{ col, p int }{{0, 0}, {2, 1}, {6, 1}} {
		if err := b.PlaceChip(hex.FromOffset(m.col, 0), board.Chip{Piece: board.FoodFacility}, m.p); err != nil {
			t.Fatal(err)
		}
	}
	rng := rand.New(rand.NewPCG(1, 2))
	for range 20 {
		if h := Choose(b, board.Chip{Piece: board.FoodFacility}, 0, Normal, rng); h != hex.FromOffset(4, 0) {
			t.Fatalf("Choose = %v, want col 4", h)
		}
	}
}

func TestRandomLevelsUseAllSpaces(t *testing.T) {
	b := parse(t, "OOOO")
	for _, l := range []Level{Easy, Hard} {
		rng := rand.New(rand.NewPCG(1, 2))
		seen := map[hex.Hex]bool{}
		for range 100 {
			seen[Choose(b, board.Chip{Piece: board.FoodFacility}, 0, l, rng)] = true
		}
		if len(seen) != 4 {
			t.Errorf("%v: chose %d of 4 spaces", l, len(seen))
		}
	}
}

func TestParseLevel(t *testing.T) {
	for _, l := range Levels {
		if got, err := ParseLevel(l.String()); err != nil || got != l {
			t.Errorf("ParseLevel(%q) = %v, %v", l, got, err)
		}
	}
	if _, err := ParseLevel("insane"); err == nil {
		t.Error("ParseLevel(insane): no error")
	}
}
