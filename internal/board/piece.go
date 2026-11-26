package board

import (
	"errors"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Piece is a tile a player places on a space: a facility or a settlement tile.
type Piece int

const (
	NoPiece Piece = iota
	FoodFacility
	EnergyFacility
	SettlementTile
)

func (f Piece) String() string {
	switch f {
	case FoodFacility:
		return "food facility"
	case EnergyFacility:
		return "energy facility"
	case SettlementTile:
		return "settlement tile"
	}
	return "none"
}

// Allows reports whether piece f may be placed on this kind of space:
// food on empty or food spaces, energy on empty or energy spaces,
// settlement tiles on settlement spaces.
func (k Kind) Allows(f Piece) bool {
	switch f {
	case FoodFacility:
		return k == Empty || k == Food
	case EnergyFacility:
		return k == Empty || k == Energy
	case SettlementTile:
		return k == Settlement
	}
	return false
}

var (
	ErrNoSpace   = errors.New("no space")
	ErrOccupied  = errors.New("space already occupied")
	ErrWrongKind = errors.New("not allowed on this space")
)

// CanPlace returns nil if piece f can be placed at h.
func (b *Board) CanPlace(h hex.Hex, f Piece) error {
	t := b.Tiles[h]
	switch {
	case t == nil:
		return ErrNoSpace
	case t.Piece != NoPiece:
		return ErrOccupied
	case !t.Kind.Allows(f):
		return ErrWrongKind
	}
	return nil
}

// Place puts piece f at h if allowed and scores it.
func (b *Board) Place(h hex.Hex, f Piece) error {
	if err := b.CanPlace(h, f); err != nil {
		return err
	}
	b.Tiles[h].Piece = f
	b.lastGains = b.placeGains(h)
	for _, g := range b.lastGains {
		b.points += g.Points
	}
	return nil
}
