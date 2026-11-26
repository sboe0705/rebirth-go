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
	b, err := Parse("x", "OFESCOKOR\nFsofe")
	if err != nil {
		t.Fatal(err)
	}
	want := []Kind{Empty, Food, Energy, Settlement, Castle, Empty, Cathedral, Empty, RoundTower}
	for col, k := range want {
		if got := b.Tiles[hex.FromOffset(col, 0)]; got.Kind != k || got.Bonus != NoBonus {
			t.Errorf("col %d: got %+v, want %v", col, got, k)
		}
	}
	for col, want := range []Tile{{Kind: Settlement, Bonus: Harbor}, {Kind: Empty, Bonus: WindCollector},
		{Kind: Food, Bonus: WindCollector}, {Kind: Energy, Bonus: WindCollector}} {
		if got := b.Tiles[hex.FromOffset(col+1, 1)]; got.Kind != want.Kind || got.Bonus != want.Bonus {
			t.Errorf("row 1 col %d: got %+v, want %+v", col+1, got, want)
		}
	}
	for _, m := range []string{"c", "k", "r"} {
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

func TestValidate(t *testing.T) {
	for _, m := range []string{"CK", "OR\nC", "ss", "sS\n.s",
		"OOO\nOsO\nOOO", "OOO\nOfO\nOOO"} { // harbor / wind collector inland
		if _, err := Parse("t", m); err == nil {
			t.Errorf("%q: expected error", m)
		}
	}
	for _, m := range []string{"COK", "sOs", "sS", "sSf", "OOO\nOSO\nOOO"} {
		if _, err := Parse("t", m); err != nil {
			t.Errorf("%q: %v", m, err)
		}
	}
}
