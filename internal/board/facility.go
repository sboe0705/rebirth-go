package board

import (
	"errors"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Facility is a facility tile placed on a space (SCO/IRL p. 4).
type Facility int

const (
	NoFacility Facility = iota
	FoodFacility
	EnergyFacility
)

func (f Facility) String() string {
	switch f {
	case FoodFacility:
		return "food facility"
	case EnergyFacility:
		return "energy facility"
	}
	return "none"
}

// Allows reports whether facility f may be placed on this kind of space:
// food on empty or food spaces, energy on empty or energy spaces.
func (k Kind) Allows(f Facility) bool {
	switch f {
	case FoodFacility:
		return k == Empty || k == Food
	case EnergyFacility:
		return k == Empty || k == Energy
	}
	return false
}

var (
	ErrNoSpace   = errors.New("no space")
	ErrOccupied  = errors.New("space already occupied")
	ErrWrongKind = errors.New("facility not allowed on this space")
)

// CanPlace returns nil if facility f can be placed at h.
func (b *Board) CanPlace(h hex.Hex, f Facility) error {
	t := b.Tiles[h]
	switch {
	case t == nil:
		return ErrNoSpace
	case t.Facility != NoFacility:
		return ErrOccupied
	case !t.Kind.Allows(f):
		return ErrWrongKind
	}
	return nil
}

// Place puts facility f at h if allowed.
func (b *Board) Place(h hex.Hex, f Facility) error {
	if err := b.CanPlace(h, f); err != nil {
		return err
	}
	b.Tiles[h].Facility = f
	return nil
}
