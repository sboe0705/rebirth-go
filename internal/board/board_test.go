package board

import (
	"os"
	"path/filepath"
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

func TestLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mine.txt")
	if err := os.WriteFile(path, []byte("# comment\nOFS\n.EC\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if b.Name != "mine" || len(b.Tiles) != 5 {
		t.Fatalf("got name %q with %d tiles, want \"mine\" with 5", b.Name, len(b.Tiles))
	}
	if _, err := LoadFile(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Fatal("expected error for a missing file")
	}
	bad := filepath.Join(t.TempDir(), "bad.txt")
	if err := os.WriteFile(bad, []byte("OX"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFile(bad); err == nil {
		t.Fatal("expected error for an invalid map")
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

func TestParseRuleViolations(t *testing.T) {
	// two adjacent castles (row 0, col 1/2) and an inland wind collector
	b, err := Parse("t", "OCCO\nOOfOO\nOOOO")
	if err == nil {
		t.Fatal("expected error")
	}
	if b == nil || len(b.Tiles) != 13 {
		t.Fatalf("board not returned with the rule error: %v", b)
	}
	want := "t: castle at row 0, col 1 and castle at row 0, col 2 are adjacent\nwind collector at row 1, col 2 is not at the coast"
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err, want)
	}
	if b, _ := Parse("t", "OX"); b != nil {
		t.Fatal("unknown character must not return a board")
	}
}

func TestValidate(t *testing.T) {
	for _, m := range []string{"CC", "KK", "RC", "KR", "OR\nC", "ss", "sS\n.s",
		"OOO\nOsO\nOOO", "OOO\nOfO\nOOO"} { // harbor / wind collector inland
		if _, err := Parse("t", m); err == nil {
			t.Errorf("%q: expected error", m)
		}
	}
	for _, m := range []string{"COK", "CK", "KC", "OK\nC", "sOs", "sS", "sSf", "OOO\nOSO\nOOO"} {
		if _, err := Parse("t", m); err != nil {
			t.Errorf("%q: %v", m, err)
		}
	}
}
