package board

import (
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func TestLoadExample(t *testing.T) {
	b, err := Load("example")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Tiles) == 0 {
		t.Fatal("no tiles")
	}
}

func TestRadius(t *testing.T) {
	if n := len(Radius(4).Tiles); n != 61 {
		t.Fatalf("got %d tiles, want 61", n)
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("x", "OX"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseKinds(t *testing.T) {
	b, err := Parse("x", "OFESCKR\nfs")
	if err != nil {
		t.Fatal(err)
	}
	want := []Kind{Empty, Food, Energy, Settlement, Castle, Cathedral, RoundTower}
	for col, k := range want {
		if got := b.Tiles[hex.FromOffset(col, 0)]; got.Kind != k || got.Bonus {
			t.Errorf("col %d: got %+v, want %v", col, got, k)
		}
	}
	if got := b.Tiles[hex.FromOffset(1, 1)]; got.Kind != Settlement || !got.Bonus {
		t.Errorf("harbour: got %+v", got)
	}
	// "+1" only on buildable spaces.
	if _, err := Parse("x", "c"); err == nil {
		t.Fatal("expected error for bonus castle")
	}
}
