package board

import (
	"fmt"
	"slices"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Scoring per player, Rebirth rules SCO p. 3–4, IRL p. 3–4.
// Mission cards and round tower bonuses are not scored.

// CastlePoints is the final score per controlled castle.
const CastlePoints = 5

// settlementRanks are the points for the most, second most, ... houses in a
// completed settlement of the given size (small settlements score houses).
// A single player in a settlement gets the sum of all ranks.
var settlementRanks = map[int][]int{2: {5, 3}, 3: {8, 5}, 4: {12, 8, 5}}

// Score is the current score of one player.
type Score struct {
	Placed          int // points scored while placing tiles
	Castles         int // final scoring: 5 per controlled castle
	OpenSettlements int // final scoring: houses in incomplete settlements
}

// Total returns the score as if the game ended now.
func (s Score) Total() int { return s.Placed + s.Castles + s.OpenSettlements }

// Gain is one part of the points scored by a placement: the player who gets
// them and the reason. Points 0 is a note, e.g. a castle changing hands.
type Gain struct {
	Player int
	Points int
	Reason string
}

var settlementNames = map[int]string{1: "small", 2: "medium", 3: "large", 4: "extra-large"}

// placeGains returns the points for the piece just placed at h:
// facilities score the size of their group, a harbor or wind collector
// scores 1, and a completed settlement scores its ranks. Castle control
// changes are listed with 0 points (castles score at game end).
func (b *Board) placeGains(h hex.Hex) []Gain {
	g := b.pieceGains(h)
	for _, n := range h.Neighbors() {
		if t := b.Tiles[n]; t != nil && t.Kind == Castle && !b.Blocked(n) {
			g = append(g, b.updateCastle(n, b.Tiles[h].Owner)...)
		}
	}
	return g
}

// updateCastle applies the castle rule after player p placed a tile next to
// castle space c (SCO p. 4, player aid): the first tile controls it; later
// p takes it with the most adjacent tiles, or when tied for most with the
// most adjacent facilities among the tied players. Otherwise it stays.
func (b *Board) updateCastle(c hex.Hex, p int) []Gain {
	tiles, facilities := map[int]int{}, map[int]int{}
	for _, n := range c.Neighbors() {
		if t := b.Tiles[n]; t != nil && t.Piece != NoPiece {
			tiles[t.Owner]++
			if t.Piece != SettlementTile {
				facilities[t.Owner]++
			}
		}
	}
	if b.castles == nil {
		b.castles = map[hex.Hex]int{}
	}
	cur, ok := b.castles[c]
	if !ok {
		b.castles[c] = p
		return []Gain{{p, 0, fmt.Sprintf("castle controlled (%d at game end)", CastlePoints)}}
	}
	if cur == p {
		return nil
	}
	most, mostF := 0, 0
	for q, n := range tiles {
		if q != p {
			most = max(most, n)
		}
	}
	for q, n := range tiles {
		if q != p && n == most {
			mostF = max(mostF, facilities[q])
		}
	}
	if tiles[p] > most || tiles[p] == most && facilities[p] > mostF {
		b.castles[c] = p
		return []Gain{
			{p, 0, fmt.Sprintf("castle taken over (%d at game end)", CastlePoints)},
			{cur, 0, "castle lost"},
		}
	}
	return nil
}

// CastleOwner returns the player controlling castle space h, if any.
func (b *Board) CastleOwner(h hex.Hex) (int, bool) {
	p, ok := b.castles[h]
	return p, ok
}

func (b *Board) pieceGains(h hex.Hex) []Gain {
	t := b.Tiles[h]
	switch t.Piece {
	case FoodFacility, EnergyFacility:
		n := b.groupSize(h)
		g := []Gain{{t.Owner, n, fmt.Sprintf("%v group of %d", t.Piece, n)}}
		if t.Bonus != NoBonus {
			g = append(g, Gain{t.Owner, 1, t.Bonus.String()})
		}
		return g
	case SettlementTile:
		var g []Gain
		if t.Bonus != NoBonus {
			g = append(g, Gain{t.Owner, 1, t.Bonus.String()})
		}
		if a := t.Area; b.complete(a) {
			g = append(g, b.settlementGains(a, t.Owner)...)
		}
		if g == nil {
			g = []Gain{{t.Owner, 0, "settlement not yet completed"}}
		}
		return g
	}
	return nil
}

// settlementGains scores settlement a, just completed by player p. Players
// are ranked by their houses in it; on a tie the completing player loses
// and the others win in turn order starting after p (SCO p. 3).
func (b *Board) settlementGains(a *SettlementArea, p int) []Gain {
	name := settlementNames[a.Size()] + " settlement completed"
	houses := map[int]int{}
	for _, h := range a.Spaces {
		if t := b.Tiles[h]; t.Piece != NoPiece { // skip blocked spaces
			houses[t.Owner] += t.Houses
		}
	}
	if a.Size() == 1 {
		return []Gain{{p, houses[p], name}}
	}
	ranks := settlementRanks[a.Size()]
	if len(houses) == 1 {
		sum := 0
		for _, r := range ranks {
			sum += r
		}
		return []Gain{{p, sum, name}}
	}
	n := max(b.Players, 1)
	order := func(q int) int { return ((q-p+n)%n + n - 1) % n } // p last
	var ps []int
	for q := range houses {
		ps = append(ps, q)
	}
	slices.SortFunc(ps, func(x, y int) int {
		if houses[x] != houses[y] {
			return houses[y] - houses[x]
		}
		return order(x) - order(y)
	})
	var g []Gain
	for i, q := range ps {
		if i < len(ranks) {
			g = append(g, Gain{q, ranks[i], fmt.Sprintf("%s, rank %d", name, i+1)})
		}
	}
	return g
}

// LastGains returns how the most recent placement scored.
func (b *Board) LastGains() []Gain { return b.lastGains }

// groupSize counts the connected pieces of the same type and owner as the one at h.
func (b *Board) groupSize(h hex.Hex) int {
	f, o := b.Tiles[h].Piece, b.Tiles[h].Owner
	seen := map[hex.Hex]bool{h: true}
	for stack := []hex.Hex{h}; len(stack) > 0; {
		c := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, n := range c.Neighbors() {
			if t := b.Tiles[n]; t != nil && t.Piece == f && t.Owner == o && !seen[n] {
				seen[n] = true
				stack = append(stack, n)
			}
		}
	}
	return len(seen)
}

// complete reports whether all spaces of a are occupied; a blocked space
// counts as occupied by its neutral tile.
func (b *Board) complete(a *SettlementArea) bool {
	for _, h := range a.Spaces {
		if b.Tiles[h].Piece == NoPiece && !b.Blocked(h) {
			return false
		}
	}
	return true
}

// Score returns player p's placement points plus the final scoring as if
// the game ended now.
func (b *Board) Score(p int) Score {
	s := Score{Placed: b.points[p]}
	for _, o := range b.castles {
		if o == p {
			s.Castles += CastlePoints
		}
	}
	for _, a := range b.Settlements {
		if b.complete(a) {
			continue
		}
		for _, h := range a.Spaces {
			if t := b.Tiles[h]; t.Piece != NoPiece && t.Owner == p {
				s.OpenSettlements += t.Houses
			}
		}
	}
	return s
}
