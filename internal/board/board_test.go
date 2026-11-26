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
	b, err := Parse("x", "OFESCKR\nFs\n---\n...a\n.b")
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
	b, err := Parse("x", "SSS.S\nSS\n---\naab.c\nbb")
	if err != nil {
		t.Fatal(err)
	}
	sizes := []int{}
	for _, a := range b.Settlements {
		sizes = append(sizes, a.Size())
	}
	if len(sizes) != 3 || sizes[0] != 2 || sizes[1] != 3 || sizes[2] != 1 {
		t.Fatalf("got sizes %v, want [2 3 1]", sizes)
	}
	if b.Tiles[hex.FromOffset(0, 0)].Area != b.Tiles[hex.FromOffset(1, 0)].Area {
		t.Fatal("a spaces not in the same area")
	}
}

func TestSettlementAreaErrors(t *testing.T) {
	for _, m := range []string{
		"SS",                // missing labels
		"SS\n---\naa\n---",  // two sections
		"SF\n---\naa",       // label on a food space
		"S.S\n---\na.a",     // not contiguous
		"SSSSS\n---\naaaaa", // too large
		"S\n---\n*",         // invalid label
	} {
		if _, err := Parse("x", m); err == nil {
			t.Errorf("expected error for %q", m)
		}
	}
	// The map character itself is allowed in the settlement grid.
	if _, err := Parse("x", "FS\n---\nFa"); err != nil {
		t.Fatal(err)
	}
}
