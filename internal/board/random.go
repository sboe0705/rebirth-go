package board

import (
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// RandomOptions controls the random map generator.
type RandomOptions struct {
	Spaces         int // number of land spaces, islands included
	Castles        int
	Cathedrals     int
	RoundTowers    int
	Settlements    []int // size of each settlement (1..MaxSettlementSize)
	Harbors        int   // coastal settlement spaces, each in a different settlement
	WindCollectors int   // coastal empty, food or energy spaces
	Food, Energy   int   // food and energy spaces; all other spaces are empty
	SpecialGap     int   // minimum distance between castles, cathedrals and round towers (>= 2)
	MaxIslands     int   // 0..MaxIslands islands besides the main land
	IslandMin      int   // island size range
	IslandMax      int
}

// DefaultRandom has the space counts of the Scotland map and up to 2 islands of 5-7 spaces.
var DefaultRandom = func() RandomOptions {
	b, err := Load("scotland")
	if err != nil {
		panic(err)
	}
	o := OptionsFrom(b)
	o.MaxIslands, o.IslandMin, o.IslandMax = 2, 5, 7
	return o
}()

// OptionsFrom returns options that reproduce the space counts of b:
// spaces, special spaces, settlement sizes, harbors, wind collectors,
// food and energy spaces. Islands are off; SpecialGap is 2.
func OptionsFrom(b *Board) RandomOptions {
	o := RandomOptions{Spaces: len(b.Tiles), SpecialGap: 2}
	for _, t := range b.Tiles {
		switch t.Kind {
		case Castle:
			o.Castles++
		case Cathedral:
			o.Cathedrals++
		case RoundTower:
			o.RoundTowers++
		case Food:
			o.Food++
		case Energy:
			o.Energy++
		}
		switch t.Bonus {
		case Harbor:
			o.Harbors++
		case WindCollector:
			o.WindCollectors++
		}
	}
	for _, a := range b.Settlements {
		o.Settlements = append(o.Settlements, a.Size())
	}
	// Largest first: they are the hardest to fit.
	slices.SortFunc(o.Settlements, func(a, b int) int { return cmp.Compare(b, a) })
	return o
}

// Random generates a board from seed; the same seed gives the same board.
// Special spaces are at least SpecialGap apart, separate settlements never touch,
// harbors (at most one per settlement) and wind collectors are at the coast.
// Islands never touch the main land or each other.
// A failed attempt is retried with the same random source.
func Random(seed uint64, o RandomOptions) (*Board, error) {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	var err error
	for try := 0; try < 100; try++ {
		var b *Board
		if b, err = randomBoard(rng, o); err == nil {
			b.Name = fmt.Sprintf("random-%d", seed)
			return b, nil
		}
	}
	return nil, err
}

func randomBoard(rng *rand.Rand, o RandomOptions) (*Board, error) {
	b := &Board{Tiles: map[hex.Hex]*Tile{}}

	islands := make([]int, rng.IntN(o.MaxIslands+1))
	main := o.Spaces
	for i := range islands {
		islands[i] = o.IslandMin + rng.IntN(o.IslandMax-o.IslandMin+1)
		main -= islands[i]
	}

	// Shape: grow a blob from the origin. order keeps iteration deterministic.
	order := []hex.Hex{{}}
	b.Tiles[hex.Hex{}] = &Tile{Kind: Empty}
	for len(order) < main {
		h := order[rng.IntN(len(order))].Neighbor(rng.IntN(6))
		// Requiring 2 land neighbors keeps the shape compact.
		if b.Tiles[h] == nil && landNeighbors(b, h) >= min(2, len(order)) {
			b.Tiles[h] = &Tile{Kind: Empty}
			order = append(order, h)
		}
	}
	for _, n := range islands {
		var ok bool
		if order, ok = addIsland(rng, b, order, n); !ok {
			return nil, fmt.Errorf("no room for an island of %d spaces", n)
		}
	}

	// Special spaces, spread out.
	var specials []hex.Hex
	placeSpecial := func(k Kind, n int) error {
		for i := 0; i < n; i++ {
			ok := false
			for try := 0; try < 1000 && !ok; try++ {
				h := order[rng.IntN(len(order))]
				if b.Tiles[h].Kind != Empty || minDistance(h, specials) < o.SpecialGap {
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
	if o.Harbors > len(o.Settlements) {
		return nil, fmt.Errorf("%d harbors but only %d settlements", o.Harbors, len(o.Settlements))
	}
	for id, size := range o.Settlements {
		if size < 1 || size > MaxSettlementSize {
			return nil, fmt.Errorf("invalid settlement size %d", size)
		}
		placed := false
		for try := 0; try < 1000 && !placed; try++ {
			start := order[rng.IntN(len(order))]
			// The first settlements start at the coast so each can get a harbor.
			if !free(start, id) || (id < o.Harbors && !b.Coast(start)) {
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

	b.groupSettlements()

	// Remaining spaces: exactly o.Food food and o.Energy energy spaces, the rest empty.
	var rest []hex.Hex
	for _, h := range order {
		if b.Tiles[h].Kind == Empty {
			rest = append(rest, h)
		}
	}
	if o.Food+o.Energy > len(rest) {
		return nil, fmt.Errorf("no room for %d food and %d energy spaces", o.Food, o.Energy)
	}
	rng.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	for i, h := range rest[:o.Food+o.Energy] {
		b.Tiles[h].Kind = Food
		if i >= o.Food {
			b.Tiles[h].Kind = Energy
		}
	}

	// Harbors (one per settlement) and wind collectors, both at the coast.
	var coast []hex.Hex
	for _, h := range order {
		if b.Coast(h) && b.Tiles[h].Kind.Buildable() {
			coast = append(coast, h)
		}
	}
	rng.Shuffle(len(coast), func(i, j int) { coast[i], coast[j] = coast[j], coast[i] })
	hasHarbor := map[*SettlementArea]bool{}
	harbors, winds := 0, 0
	for _, h := range coast {
		switch t := b.Tiles[h]; {
		case t.Kind == Settlement && harbors < o.Harbors && !hasHarbor[t.Area]:
			t.Bonus = Harbor
			hasHarbor[t.Area] = true
			harbors++
		case t.Kind != Settlement && winds < o.WindCollectors:
			t.Bonus = WindCollector
			winds++
		}
	}
	if harbors < o.Harbors || winds < o.WindCollectors {
		return nil, fmt.Errorf("no room for %d harbors and %d wind collectors at the coast", o.Harbors, o.WindCollectors)
	}

	if err := b.validate(); err != nil {
		return nil, err
	}
	return b, nil
}

// addIsland grows an island of n spaces off the coast that touches no other
// land, and appends its spaces to order.
func addIsland(rng *rand.Rand, b *Board, order []hex.Hex, n int) ([]hex.Hex, bool) {
	island := map[hex.Hex]bool{}
	// open reports whether h is water and touches no land outside the island.
	open := func(h hex.Hex) bool {
		if b.Tiles[h] != nil || island[h] {
			return false
		}
		for _, x := range h.Neighbors() {
			if b.Tiles[x] != nil && !island[x] {
				return false
			}
		}
		return true
	}
	for try := 0; try < 100; try++ {
		// Start 2-3 steps off a coastal space: one or two rows of water in between.
		h, d := order[rng.IntN(len(order))], rng.IntN(6)
		if !b.Coast(h) {
			continue
		}
		for b.Tiles[h] != nil {
			h = h.Neighbor(d)
		}
		h = h.Neighbor(d)
		if rng.IntN(2) == 0 {
			h = h.Neighbor(d)
		}
		if !open(h) {
			continue
		}
		spaces := []hex.Hex{h}
		island[h] = true
		for len(spaces) < n {
			var cand []hex.Hex
			for _, s := range spaces {
				for _, x := range s.Neighbors() {
					// 2 island neighbors keep it compact, like the main land.
					if open(x) && islandNeighbors(island, x) >= min(2, len(spaces)) {
						cand = append(cand, x)
					}
				}
			}
			if len(cand) == 0 {
				break
			}
			x := cand[rng.IntN(len(cand))]
			island[x] = true
			spaces = append(spaces, x)
		}
		if len(spaces) < n {
			clear(island)
			continue
		}
		for _, s := range spaces {
			b.Tiles[s] = &Tile{Kind: Empty}
		}
		return append(order, spaces...), true
	}
	return order, false
}

func islandNeighbors(island map[hex.Hex]bool, h hex.Hex) int {
	n := 0
	for _, x := range h.Neighbors() {
		if island[x] {
			n++
		}
	}
	return n
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
