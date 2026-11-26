package board

import (
	"errors"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func TestPlacePiece(t *testing.T) {
	b, err := Parse("t", "OFESCFE") // F and E at the end keep the fallback rule out
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
		{7, FoodFacility, ErrNoSpace},
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

func TestFacilityFallback(t *testing.T) {
	b, err := Parse("t", "FEOS")
	if err != nil {
		t.Fatal(err)
	}
	at := func(col int) hex.Hex { return hex.FromOffset(col, 0) }
	if err := b.CanPlace(at(1), FoodFacility); !errors.Is(err, ErrWrongKind) {
		t.Errorf("food on energy with free spaces left: %v", err)
	}
	for _, m := range []struct {
		col int
		f   Piece
	}{{0, FoodFacility}, {2, EnergyFacility}} {
		if err := b.Place(at(m.col), m.f); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.CanPlace(at(1), FoodFacility); err != nil {
		t.Errorf("food on energy without food/empty spaces: %v", err)
	}
	if !b.CanPlaceAnywhere(SettlementTile) {
		t.Error("settlement space is free")
	}
	if err := b.Place(at(3), SettlementTile); err != nil {
		t.Fatal(err)
	}
	if b.CanPlaceAnywhere(SettlementTile) {
		t.Error("no settlement space left")
	}
}
