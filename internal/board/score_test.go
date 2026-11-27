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
		var err error
		switch m := moves[i+1].(type) {
		case Piece:
			err = b.Place(h, m)
		case Chip:
			err = b.PlaceChip(h, m, 0)
		}
		if err != nil {
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
		{"houses", placeAll(t, "S.SSS",
			0, Chip{SettlementTile, 3}, 2, Chip{SettlementTile, 4}, 4, Chip{SettlementTile, 2}), Score{Placed: 3, OpenSettlements: 6}},
		{"castle", placeAll(t, "OCO",
			0, FoodFacility, 2, FoodFacility), Score{Placed: 2, Castles: 5}},
		{"no castle", placeAll(t, "OOC"), Score{}},
	}
	for _, c := range cases {
		if got := c.b.Score(0); got != c.want {
			t.Errorf("%s: Score() = %+v, want %+v", c.name, got, c.want)
		}
	}
	if got := (Score{1, 5, 2, 3}).Total(); got != 11 {
		t.Errorf("Total = %d, want 11", got)
	}
}

func TestFailedPlaceScoresNothing(t *testing.T) {
	b := placeAll(t, "O", 0, FoodFacility)
	_ = b.Place(hex.FromOffset(0, 0), FoodFacility)
	if got := b.Score(0).Placed; got != 1 {
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
			[]Gain{{0, 2, "food facility group of 2"}}},
		{"settlement", placeAll(t, "sS", 1, SettlementTile, 0, SettlementTile),
			[]Gain{{0, 1, "harbor"}, {0, 8, "medium settlement completed"}}},
		{"wind collector", placeAll(t, "fO", 1, FoodFacility, 0, FoodFacility),
			[]Gain{{0, 2, "food facility group of 2"}, {0, 1, "wind collector"}}},
		{"open settlement", placeAll(t, "SS", 0, SettlementTile),
			[]Gain{{0, 0, "settlement not yet completed"}}},
		{"castle", placeAll(t, "OCO", 0, EnergyFacility),
			[]Gain{{0, 1, "energy facility group of 1"}, {0, 0, "castle controlled (5 at game end)"}}},
		{"castle already controlled", placeAll(t, "OCO", 0, EnergyFacility, 2, EnergyFacility),
			[]Gain{{0, 1, "energy facility group of 1"}}},
	}
	for _, c := range cases {
		if got := c.b.LastGains(); !slices.Equal(got, c.want) {
			t.Errorf("%s: LastGains() = %v, want %v", c.name, got, c.want)
		}
	}
}

// play parses a one-row map with n players and places chips: col, chip, player, ...
func play(t *testing.T, text string, n int, moves ...any) *Board {
	t.Helper()
	b, err := Parse("t", text)
	if err != nil {
		t.Fatal(err)
	}
	b.Players = n
	for i := 0; i < len(moves); i += 3 {
		c := moves[i+1].(Chip)
		if err := b.PlaceChip(hex.FromOffset(moves[i].(int), 0), c, moves[i+2].(int)); err != nil {
			t.Fatalf("PlaceChip(col %d): %v", moves[i], err)
		}
	}
	return b
}

var food, energy = Chip{FoodFacility, 0}, Chip{EnergyFacility, 0}

func house(n int) Chip { return Chip{SettlementTile, n} }

func TestMultiplayerGroups(t *testing.T) {
	b := play(t, "OOO", 2, 0, food, 0, 1, food, 1, 2, food, 0)
	if got := b.LastGains(); !slices.Equal(got, []Gain{{0, 1, "food facility group of 1"}}) {
		t.Errorf("other colour joined the group: %v", got)
	}
	if b.Score(0).Placed != 2 || b.Score(1).Placed != 1 {
		t.Errorf("scores %v %v", b.Score(0), b.Score(1))
	}
}

func TestMultiplayerSettlements(t *testing.T) {
	cases := []struct {
		name string
		b    *Board
		want []Gain
	}{
		{"medium majority", play(t, "SS", 2, 0, house(1), 1, 1, house(2), 0),
			[]Gain{{0, 5, "medium settlement completed, rank 1"}, {1, 3, "medium settlement completed, rank 2"}}},
		{"medium tie: completer gets 3", play(t, "SS", 2, 0, house(2), 1, 1, house(2), 0),
			[]Gain{{1, 5, "medium settlement completed, rank 1"}, {0, 3, "medium settlement completed, rank 2"}}},
		{"large tie: turn order after completer", play(t, "SSS", 3,
			0, house(1), 0, 1, house(1), 1, 2, house(1), 2),
			[]Gain{{0, 8, "large settlement completed, rank 1"}, {1, 5, "large settlement completed, rank 2"}}},
		{"large tie wraps around", play(t, "SSS", 3,
			0, house(1), 2, 1, house(1), 1, 2, house(1), 0),
			[]Gain{{1, 8, "large settlement completed, rank 1"}, {2, 5, "large settlement completed, rank 2"}}},
		{"one player gets all ranks", play(t, "SS", 2, 0, house(1), 1, 1, house(1), 1),
			[]Gain{{1, 8, "medium settlement completed"}}},
	}
	for _, c := range cases {
		if got := c.b.LastGains(); !slices.Equal(got, c.want) {
			t.Errorf("%s: LastGains() = %v, want %v", c.name, got, c.want)
		}
	}
	b := play(t, "SSS", 2, 0, house(3), 1, 1, house(2), 0)
	if b.Score(0).OpenSettlements != 2 || b.Score(1).OpenSettlements != 3 {
		t.Errorf("open settlements %v %v", b.Score(0), b.Score(1))
	}
}

func TestMultiplayerCastles(t *testing.T) {
	// Castle at column 1 of row 0 has the neighbors column 0 and 2 in this row.
	owner := func(b *Board) int {
		p, ok := b.CastleOwner(hex.FromOffset(1, 0))
		if !ok {
			return -1
		}
		return p
	}
	cases := []struct {
		name string
		b    *Board
		want int
	}{
		{"first tile", play(t, "OCO", 2, 0, food, 1), 1},
		{"tie stays", play(t, "OCO", 2, 0, food, 0, 2, energy, 1), 0},
		{"tie, more facilities", play(t, "SCO", 2, 0, house(1), 0, 2, energy, 1), 1},
		{"tie, fewer facilities", play(t, "OCS", 2, 0, food, 0, 2, house(4), 1), 0},
	}
	for _, c := range cases {
		if got := owner(c.b); got != c.want {
			t.Errorf("%s: owner %d, want %d", c.name, got, c.want)
		}
	}
	b := play(t, "SCO", 2, 0, house(1), 0, 2, energy, 1)
	if b.Score(0).Castles != 0 || b.Score(1).Castles != CastlePoints {
		t.Errorf("castle scores %v %v", b.Score(0), b.Score(1))
	}
	want := []Gain{{1, 1, "energy facility group of 1"}, {1, 0, "castle taken over (5 at game end)"}, {0, 0, "castle lost"}}
	if got := b.LastGains(); !slices.Equal(got, want) {
		t.Errorf("LastGains() = %v, want %v", got, want)
	}
}

func TestMoreTilesTakeCastle(t *testing.T) {
	text := "OCO\n.OO" // castle (1,0) has neighbors (0,0), (2,0) and (1,1)
	b := play(t, text, 2, 0, food, 0)
	for _, m := range []struct{ col, row, p int }{{2, 0, 1}, {1, 1, 1}} {
		if err := b.PlaceChip(hex.FromOffset(m.col, m.row), food, m.p); err != nil {
			t.Fatal(err)
		}
	}
	if p, _ := b.CastleOwner(hex.FromOffset(1, 0)); p != 1 {
		t.Errorf("owner %d, want 1", p)
	}
}
