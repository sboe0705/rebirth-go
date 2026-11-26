package board

import (
	"fmt"
	"math/rand/v2"
)

// Chip is one tile from a player's supply: a facility or a settlement tile
// with 1–4 settlement symbols (houses).
type Chip struct {
	Piece  Piece
	Houses int // settlement symbols; 0 for facilities
}

func (c Chip) String() string {
	if c.Piece == SettlementTile {
		if c.Houses == 1 {
			return "settlement tile (1 house)"
		}
		return fmt.Sprintf("settlement tile (%d houses)", c.Houses)
	}
	return c.Piece.String()
}

// ChipSet returns the 36 tiles of one colour (SCO p. 1): 12 food facilities,
// 12 energy facilities and 3 settlement tiles each with 1, 2, 3 and 4 houses.
func ChipSet() []Chip {
	var s []Chip
	for i := 0; i < 12; i++ {
		s = append(s, Chip{FoodFacility, 0}, Chip{EnergyFacility, 0})
	}
	for h := 1; h <= 4; h++ {
		for i := 0; i < 3; i++ {
			s = append(s, Chip{SettlementTile, h})
		}
	}
	return s
}

// SetAside is the number of tiles drawn in setup that are not used (SCO p. 2).
const SetAside = 2

// Supply is a player's shuffled, face-down pile of tiles.
type Supply struct {
	chips []Chip
	Aside []Chip // drawn in setup, not used this game
}

// ChipSetFor returns the tiles of one colour used with the given number of
// players: with 4 players 3 food, 3 energy and one settlement tile each with
// 2, 3 and 4 houses are removed (SCO p. 2), leaving 27.
func ChipSetFor(players int) []Chip {
	c := ChipSet()
	if players < 4 {
		return c
	}
	remove := map[Chip]int{{FoodFacility, 0}: 3, {EnergyFacility, 0}: 3,
		{SettlementTile, 2}: 1, {SettlementTile, 3}: 1, {SettlementTile, 4}: 1}
	var s []Chip
	for _, x := range c {
		if remove[x] > 0 {
			remove[x]--
			continue
		}
		s = append(s, x)
	}
	return s
}

// NewSupply shuffles a full chip set and sets SetAside tiles aside.
func NewSupply(seed uint64) *Supply { return NewSupplyFor(seed, 1) }

// NewSupplyFor is NewSupply with the chip set for the given number of players.
func NewSupplyFor(seed uint64, players int) *Supply {
	rng := rand.New(rand.NewPCG(seed, seed^0x5851f42d4c957f2d))
	c := ChipSetFor(players)
	rng.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })
	return &Supply{chips: c[SetAside:], Aside: c[:SetAside]}
}

// Len returns the number of tiles left in the supply.
func (s *Supply) Len() int { return len(s.chips) }

// Draw takes the top tile; ok is false if the supply is empty.
func (s *Supply) Draw() (c Chip, ok bool) {
	if len(s.chips) == 0 {
		return Chip{}, false
	}
	c, s.chips = s.chips[0], s.chips[1:]
	return c, true
}
