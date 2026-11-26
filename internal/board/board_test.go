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
	b, err := Parse("x", "OFESCKR\nFs")
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
	// "+1" only on settlement spaces.
	for _, m := range []string{"o", "f", "e", "c", "k", "r"} {
		if _, err := Parse("x", m); err == nil {
			t.Errorf("expected error for %q", m)
		}
	}
}

func TestSettlementAreas(t *testing.T) {
	// One medium settlement and two separate small ones.
	b, err := Parse("x", "SS.S\n..F.S")
	if err != nil {
		t.Fatal(err)
	}
	sizes := map[int]int{}
	for _, a := range b.Settlements {
		sizes[a.Size()]++
	}
	if len(b.Settlements) != 3 || sizes[2] != 1 || sizes[1] != 2 {
		t.Fatalf("got %d settlements, sizes %v", len(b.Settlements), sizes)
	}
	if b.Tiles[hex.FromOffset(0, 0)].Area != b.Tiles[hex.FromOffset(1, 0)].Area {
		t.Fatal("adjacent settlement spaces not in the same settlement")
	}
	if _, err := Parse("x", "SSSSS"); err == nil {
		t.Fatal("expected error for settlement with 5 spaces")
	}
}
