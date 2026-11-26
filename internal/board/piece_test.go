package board

import (
	"errors"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func TestPlacePiece(t *testing.T) {
	b, err := Parse("t", "OFESC")
	if err != nil {
		t.Fatal(err)
	}
	at := func(col int) hex.Hex { return hex.FromOffset(col, 0) }
	cases := []struct {
		col  int
		f    Piece
		want error
	}{
		{0, FoodFacility, nil},
		{1, EnergyFacility, ErrWrongKind},
		{1, FoodFacility, nil},
		{2, FoodFacility, ErrWrongKind},
		{2, EnergyFacility, nil},
		{3, FoodFacility, ErrWrongKind},
		{3, SettlementTile, nil},
		{0, SettlementTile, ErrOccupied},
		{4, SettlementTile, ErrWrongKind},
		{4, EnergyFacility, ErrWrongKind},
		{5, FoodFacility, ErrNoSpace},
		{0, EnergyFacility, ErrOccupied},
		{0, NoPiece, ErrOccupied},
	}
	for _, c := range cases {
		if got := b.Place(at(c.col), c.f); !errors.Is(got, c.want) {
			t.Errorf("Place(col %d, %v) = %v, want %v", c.col, c.f, got, c.want)
		}
	}
	if b.Tiles[at(0)].Piece != FoodFacility || b.Tiles[at(2)].Piece != EnergyFacility {
		t.Error("facilities not stored")
	}
}

func TestEmptyAllowsBoth(t *testing.T) {
	b := Radius(0)
	if err := b.CanPlace(hex.Hex{}, EnergyFacility); err != nil {
		t.Error(err)
	}
	if err := b.CanPlace(hex.Hex{}, NoPiece); !errors.Is(err, ErrWrongKind) {
		t.Errorf("NoPiece: %v", err)
	}
}
