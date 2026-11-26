package board

import (
	"fmt"
	"math/rand/v2"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// RandomOptions controls the random map generator.
type RandomOptions struct {
	Spaces      int // number of land spaces
	Castles     int
	Cathedrals  int
	RoundTowers int
	Settlements []int // size of each settlement (1..MaxSettlementSize)
	Bonus       int   // settlement spaces with "+1", preferably at the coast
}

// DefaultRandom is a medium-sized board similar to the example map.
var DefaultRandom = RandomOptions{
	Spaces:      100,
	Castles:     4,
	Cathedrals:  3,
	Settlements: []int{1, 1, 1, 2, 2, 2, 3, 3},
	Bonus:       3,
}

// Random generates a board from seed; the same seed gives the same board.
// Special spaces are at least 3 apart, separate settlements never touch,
// all other spaces are empty, food or energy.
func Random(seed uint64, o RandomOptions) (*Board, error) {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	b := &Board{Name: fmt.Sprintf("random-%d", seed), Tiles: map[hex.Hex]*Tile{}}

	// Shape: grow a blob from the origin. order keeps iteration deterministic.
	order := []hex.Hex{{}}
	b.Tiles[hex.Hex{}] = &Tile{Kind: Empty}
	for len(order) < o.Spaces {
		h := order[rng.IntN(len(order))].Neighbor(rng.IntN(6))
		// Requiring 2 land neighbors keeps the shape compact.
		if b.Tiles[h] == nil && landNeighbors(b, h) >= min(2, len(order)) {
			b.Tiles[h] = &Tile{Kind: Empty}
			order = append(order, h)
		}
	}

	// Special spaces, spread out.
	var specials []hex.Hex
	placeSpecial := func(k Kind, n int) error {
		for i := 0; i < n; i++ {
			ok := false
			for try := 0; try < 1000 && !ok; try++ {
				h := order[rng.IntN(len(order))]
				if b.Tiles[h].Kind != Empty || minDistance(h, specials) < 3 {
					continue
				}
				b.Tiles[h].Kind = k
				specials = append(specials, h)
				ok = true
			}
			if !ok {
				return fmt.Errorf("no room for %d spaces of kind %d", n, k)
			}
		}
		return nil
	}
	for _, s := range []struct {
		k Kind
		n int
	}{{Castle, o.Castles}, {Cathedral, o.Cathedrals}, {RoundTower, o.RoundTowers}} {
		if err := placeSpecial(s.k, s.n); err != nil {
			return nil, err
		}
	}

	// Settlements: grow each group into free spaces that do not touch another group.
	group := map[hex.Hex]int{}
	free := func(h hex.Hex, id int) bool {
		t := b.Tiles[h]
		if t == nil || t.Kind != Empty {
			return false
		}
		for _, n := range h.Neighbors() {
			if g, ok := group[n]; ok && g != id {
				return false
			}
		}
		return true
	}
	for id, size := range o.Settlements {
		if size < 1 || size > MaxSettlementSize {
			return nil, fmt.Errorf("invalid settlement size %d", size)
		}
		placed := false
		for try := 0; try < 1000 && !placed; try++ {
			start := order[rng.IntN(len(order))]
			if !free(start, id) {
				continue
			}
			spaces := []hex.Hex{start}
			group[start] = id
			for len(spaces) < size {
				var cand []hex.Hex
				for _, s := range spaces {
					for _, n := range s.Neighbors() {
						if _, taken := group[n]; !taken && free(n, id) {
							cand = append(cand, n)
						}
					}
				}
				if len(cand) == 0 {
					break
				}
				n := cand[rng.IntN(len(cand))]
				group[n] = id
				spaces = append(spaces, n)
			}
			if len(spaces) < size {
				for _, s := range spaces {
					delete(group, s)
				}
				continue
			}
			for _, s := range spaces {
				b.Tiles[s].Kind = Settlement
			}
			placed = true
		}
		if !placed {
			return nil, fmt.Errorf("no room for settlement of size %d", size)
		}
	}

	// "+1": coastal settlement spaces first.
	var coast, inland []hex.Hex
	for _, h := range order {
		if b.Tiles[h].Kind != Settlement {
			continue
		}
		if landNeighbors(b, h) < 6 {
			coast = append(coast, h)
		} else {
			inland = append(inland, h)
		}
	}
	rng.Shuffle(len(coast), func(i, j int) { coast[i], coast[j] = coast[j], coast[i] })
	rng.Shuffle(len(inland), func(i, j int) { inland[i], inland[j] = inland[j], inland[i] })
	for i, h := range append(coast, inland...) {
		if i >= o.Bonus {
			break
		}
		b.Tiles[h].Bonus = true
	}

	// Remaining spaces: empty, food or energy.
	for _, h := range order {
		if t := b.Tiles[h]; t.Kind == Empty {
			t.Kind = [...]Kind{Empty, Empty, Food, Food, Energy, Energy, Empty}[rng.IntN(7)]
		}
	}

	if err := b.groupSettlements(); err != nil {
		return nil, err
	}
	return b, nil
}

func landNeighbors(b *Board, h hex.Hex) int {
	n := 0
	for _, x := range h.Neighbors() {
		if b.Tiles[x] != nil {
			n++
		}
	}
	return n
}

func minDistance(h hex.Hex, others []hex.Hex) int {
	m := 1 << 30
	for _, o := range others {
		m = min(m, hex.Distance(h, o))
	}
	return m
}
