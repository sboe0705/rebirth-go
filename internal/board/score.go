package board

import "github.com/sboe0705/rebirth-go/internal/hex"

// Scoring for a single player (no opponents yet), Rebirth rules SCO p. 3–4,
// IRL p. 3–4. Mission cards and round tower bonuses are not scored.

// HousesPerSettlementTile is the number of settlement symbols on a placed
// settlement tile. Tiles with 2–4 houses are not modelled yet.
const HousesPerSettlementTile = 1

// CastlePoints is the final score per controlled castle.
const CastlePoints = 5

// completedSettlementPoints is what a single player gets for completing a
// settlement of the given size: the sum of all ranks (small: houses).
var completedSettlementPoints = map[int]int{2: 5 + 3, 3: 8 + 5, 4: 12 + 8 + 5}

// Score is the current score of the player.
type Score struct {
	Placed          int // points scored while placing tiles
	Castles         int // final scoring: 5 per controlled castle
	OpenSettlements int // final scoring: houses in incomplete settlements
}

// Total returns the score as if the game ended now.
func (s Score) Total() int { return s.Placed + s.Castles + s.OpenSettlements }

// placePoints returns the points for the piece just placed at h:
// facilities score the size of their group, a "+1" space scores 1, and a
// completed settlement scores its value.
func (b *Board) placePoints(h hex.Hex) int {
	t := b.Tiles[h]
	switch t.Piece {
	case FoodFacility, EnergyFacility:
		return b.groupSize(h)
	case SettlementTile:
		p := 0
		if t.Bonus {
			p++
		}
		if a := t.Area; b.complete(a) {
			if a.Size() == 1 {
				p += HousesPerSettlementTile
			} else {
				p += completedSettlementPoints[a.Size()]
			}
		}
		return p
	}
	return 0
}

// groupSize counts the connected pieces of the same type as the one at h.
func (b *Board) groupSize(h hex.Hex) int {
	f := b.Tiles[h].Piece
	seen := map[hex.Hex]bool{h: true}
	for stack := []hex.Hex{h}; len(stack) > 0; {
		c := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, n := range c.Neighbors() {
			if t := b.Tiles[n]; t != nil && t.Piece == f && !seen[n] {
				seen[n] = true
				stack = append(stack, n)
			}
		}
	}
	return len(seen)
}

func (b *Board) complete(a *SettlementArea) bool {
	for _, h := range a.Spaces {
		if b.Tiles[h].Piece == NoPiece {
			return false
		}
	}
	return true
}

// Score returns the placement points plus the final scoring as if the game
// ended now. A castle is controlled once a tile is adjacent to it (only one
// player, so there is no majority to win).
func (b *Board) Score() Score {
	s := Score{Placed: b.points}
	for h, t := range b.Tiles {
		if t.Kind != Castle {
			continue
		}
		for _, n := range h.Neighbors() {
			if nt := b.Tiles[n]; nt != nil && nt.Piece != NoPiece {
				s.Castles += CastlePoints
				break
			}
		}
	}
	for _, a := range b.Settlements {
		if b.complete(a) {
			continue
		}
		for _, h := range a.Spaces {
			if b.Tiles[h].Piece == SettlementTile {
				s.OpenSettlements += HousesPerSettlementTile
			}
		}
	}
	return s
}
