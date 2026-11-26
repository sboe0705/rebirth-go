package board

import (
	"slices"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func placeAll(t *testing.T, text string, moves ...any) *Board {
	t.Helper()
	b, err := Parse("t", text)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(moves); i += 2 {
		h := hex.FromOffset(moves[i].(int), 0)
		if err := b.Place(h, moves[i+1].(Piece)); err != nil {
			t.Fatalf("Place(col %d): %v", moves[i], err)
		}
	}
	return b
}

func TestScore(t *testing.T) {
	cases := []struct {
		name string
		b    *Board
		want Score
	}{
		{"groups", placeAll(t, "OOOE",
			0, FoodFacility, 1, FoodFacility, 3, EnergyFacility, 2, FoodFacility), Score{Placed: 1 + 2 + 1 + 3}},
		{"types never group", placeAll(t, "OO",
			0, FoodFacility, 1, EnergyFacility), Score{Placed: 2}},
		{"settlements", placeAll(t, "sS.S",
			0, SettlementTile, 1, SettlementTile, 3, SettlementTile), Score{Placed: 1 + 8 + 1}},
		{"large", placeAll(t, "SSS",
			0, SettlementTile, 1, SettlementTile, 2, SettlementTile), Score{Placed: 13}},
		{"open settlement", placeAll(t, "SSS",
			0, SettlementTile, 2, SettlementTile), Score{OpenSettlements: 2}},
		{"castle", placeAll(t, "OCO",
			0, FoodFacility, 2, FoodFacility), Score{Placed: 2, Castles: 5}},
		{"no castle", placeAll(t, "OOC"), Score{}},
	}
	for _, c := range cases {
		if got := c.b.Score(); got != c.want {
			t.Errorf("%s: Score() = %+v, want %+v", c.name, got, c.want)
		}
	}
	if got := (Score{1, 5, 2}).Total(); got != 8 {
		t.Errorf("Total = %d, want 8", got)
	}
}

func TestFailedPlaceScoresNothing(t *testing.T) {
	b := placeAll(t, "O", 0, FoodFacility)
	_ = b.Place(hex.FromOffset(0, 0), FoodFacility)
	if got := b.Score().Placed; got != 1 {
		t.Errorf("Placed = %d, want 1", got)
	}
}

func TestLastGains(t *testing.T) {
	cases := []struct {
		name string
		b    *Board
		want []Gain
	}{
		{"group", placeAll(t, "OO", 0, FoodFacility, 1, FoodFacility),
			[]Gain{{2, "food facility group of 2"}}},
		{"settlement", placeAll(t, "sS", 1, SettlementTile, 0, SettlementTile),
			[]Gain{{1, "harbor"}, {8, "medium settlement completed"}}},
		{"wind collector", placeAll(t, "fO", 1, FoodFacility, 0, FoodFacility),
			[]Gain{{2, "food facility group of 2"}, {1, "wind collector"}}},
		{"open settlement", placeAll(t, "SS", 0, SettlementTile),
			[]Gain{{0, "settlement not yet completed"}}},
		{"castle", placeAll(t, "OCO", 0, EnergyFacility),
			[]Gain{{1, "energy facility group of 1"}, {0, "castle controlled (5 at game end)"}}},
		{"castle already controlled", placeAll(t, "OCO", 0, EnergyFacility, 2, EnergyFacility),
			[]Gain{{1, "energy facility group of 1"}}},
	}
	for _, c := range cases {
		if got := c.b.LastGains(); !slices.Equal(got, c.want) {
			t.Errorf("%s: LastGains() = %v, want %v", c.name, got, c.want)
		}
	}
}
