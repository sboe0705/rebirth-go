package board

import (
	"errors"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func TestParseLocks(t *testing.T) {
	b, err := Parse("t", "OSC\nFSEK\n# comment\n[2 players]\nx.x\n...x")
	if err != nil {
		t.Fatal(err)
	}
	locked := map[hex.Hex]bool{hex.FromOffset(0, 0): true, hex.FromOffset(2, 0): true, hex.FromOffset(3, 1): true}
	for h, tile := range b.Tiles {
		if tile.Locked != locked[h] {
			t.Errorf("%s: locked %v, want %v", Pos(h), tile.Locked, locked[h])
		}
	}
	for players, want := range map[int]bool{0: false, 1: false, 2: true, 3: false, 4: false} {
		b.Players = players
		if got := b.Blocked(hex.FromOffset(0, 0)); got != want {
			t.Errorf("%d players: blocked %v, want %v", players, got, want)
		}
	}
	for _, m := range []string{
		"O.\n[2 players]\n.x",            // water
		"OO\n[2 players]\n.y",            // unknown character
		"OO\n[2 players]\n\n[2 players]", // twice
	} {
		if b, err := Parse("t", m); err == nil || b != nil {
			t.Errorf("%q: expected error and no board, got %v", m, err)
		}
	}
}

func TestBlockedPlacement(t *testing.T) {
	b, err := Parse("t", "OE\n[2 players]\nx")
	if err != nil {
		t.Fatal(err)
	}
	b.Players = 2
	if err := b.CanPlace(hex.FromOffset(0, 0), FoodFacility); !errors.Is(err, ErrBlocked) {
		t.Fatalf("got %v, want ErrBlocked", err)
	}
	// the only empty space is blocked, so food may go on the energy space
	if err := b.CanPlace(hex.FromOffset(1, 0), FoodFacility); err != nil {
		t.Fatalf("fallback: %v", err)
	}
	b.Players = 3
	if err := b.CanPlace(hex.FromOffset(0, 0), FoodFacility); err != nil {
		t.Fatalf("3 players: %v", err)
	}
}

func TestBlockedScoring(t *testing.T) {
	// medium settlement with one locked space, blocked castle next to a food space
	b, err := Parse("t", "SSOCF\n[2 players]\n.x.x")
	if err != nil {
		t.Fatal(err)
	}
	b.Players = 2
	if err := b.PlaceChip(hex.FromOffset(0, 0), Chip{SettlementTile, 2}, 1); err != nil {
		t.Fatal(err)
	}
	if g := b.LastGains(); len(g) != 1 || g[0].Player != 1 || g[0].Points != 8 {
		t.Fatalf("settlement gains %v, want player 1 with 8 (5+3)", g)
	}
	if err := b.PlaceChip(hex.FromOffset(4, 0), Chip{FoodFacility, 0}, 0); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.CastleOwner(hex.FromOffset(3, 0)); ok {
		t.Fatal("blocked castle got an owner")
	}
	if s := b.Score(0); s.Castles != 0 {
		t.Fatalf("castle points %d for a blocked castle", s.Castles)
	}
}
