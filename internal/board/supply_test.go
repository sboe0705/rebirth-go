package board

import "testing"

func TestChipSet(t *testing.T) {
	n := map[Chip]int{}
	for _, c := range ChipSet() {
		n[c]++
	}
	want := map[Chip]int{
		{FoodFacility, 0}: 12, {EnergyFacility, 0}: 12,
		{SettlementTile, 1}: 3, {SettlementTile, 2}: 3, {SettlementTile, 3}: 3, {SettlementTile, 4}: 3,
	}
	if len(n) != len(want) {
		t.Fatalf("chip kinds = %v", n)
	}
	for c, w := range want {
		if n[c] != w {
			t.Errorf("%v: %d, want %d", c, n[c], w)
		}
	}
}

func TestSupply(t *testing.T) {
	s := NewSupply(1)
	if len(s.Aside) != SetAside || s.Len() != 36-SetAside {
		t.Fatalf("aside %d, supply %d", len(s.Aside), s.Len())
	}
	n := map[Chip]int{}
	for _, c := range s.Aside {
		n[c]++
	}
	for i := 36 - SetAside; i > 0; i-- {
		c, ok := s.Draw()
		if !ok || s.Len() != i-1 {
			t.Fatalf("Draw: ok %v, len %d", ok, s.Len())
		}
		n[c]++
	}
	if _, ok := s.Draw(); ok {
		t.Error("Draw from empty supply")
	}
	if n[Chip{FoodFacility, 0}] != 12 || n[Chip{SettlementTile, 4}] != 3 {
		t.Errorf("not a full set: %v", n)
	}
	a, b := NewSupply(7), NewSupply(7)
	for a.Len() > 0 {
		x, _ := a.Draw()
		if y, _ := b.Draw(); x != y {
			t.Fatal("same seed, different order")
		}
	}
}

func TestChipString(t *testing.T) {
	for c, want := range map[Chip]string{
		{FoodFacility, 0}:   "food facility",
		{SettlementTile, 1}: "settlement tile (1 house)",
		{SettlementTile, 3}: "settlement tile (3 houses)",
	} {
		if got := c.String(); got != want {
			t.Errorf("%v: %q, want %q", c, got, want)
		}
	}
}

func TestChipSetFor4(t *testing.T) {
	n := map[Chip]int{}
	for _, c := range ChipSetFor(4) {
		n[c]++
	}
	if len(ChipSetFor(4)) != 27 || n[Chip{FoodFacility, 0}] != 9 || n[Chip{SettlementTile, 1}] != 3 || n[Chip{SettlementTile, 4}] != 2 {
		t.Errorf("4-player set: %v", n)
	}
	if len(ChipSetFor(3)) != 36 || NewSupplyFor(1, 4).Len() != 27-SetAside {
		t.Error("wrong set sizes")
	}
}
