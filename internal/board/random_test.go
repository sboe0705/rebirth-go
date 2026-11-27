package board

import (
	"slices"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func TestRandomDeterministic(t *testing.T) {
	a, err := Random(42, DefaultRandom)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Random(42, DefaultRandom)
	for h, ta := range a.Tiles {
		if tb := b.Tiles[h]; tb == nil || tb.Kind != ta.Kind || tb.Bonus != ta.Bonus {
			t.Fatalf("seed 42 differs at %v", h)
		}
	}
}

func TestRandomValid(t *testing.T) {
	o := DefaultRandom
	islands := map[int]int{} // maps per island count
	for seed := uint64(0); seed < 300; seed++ {
		b, err := Random(seed, o)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if len(b.Tiles) != o.Spaces {
			t.Fatalf("seed %d: %d spaces", seed, len(b.Tiles))
		}
		count := map[Kind]int{}
		bonus := map[Bonus]int{}
		var specials []hex.Hex
		for h, tl := range b.Tiles {
			count[tl.Kind]++
			bonus[tl.Bonus]++
			if tl.Bonus != NoBonus && !b.Coast(h) {
				t.Fatalf("seed %d: %v at %v not at the coast", seed, tl.Bonus, h)
			}
			if (tl.Bonus == Harbor) != (tl.Bonus != NoBonus && tl.Kind == Settlement) {
				t.Fatalf("seed %d: %v on %v space", seed, tl.Bonus, tl.Kind)
			}
			if tl.Kind > Settlement {
				specials = append(specials, h)
			}
		}
		if count[Castle] != o.Castles || count[Cathedral] != o.Cathedrals || count[RoundTower] != o.RoundTowers ||
			count[Food] != o.Food || count[Energy] != o.Energy ||
			bonus[Harbor] != o.Harbors || bonus[WindCollector] != o.WindCollectors {
			t.Fatalf("seed %d: counts %v, bonus %d", seed, count, bonus)
		}
		for i, s := range specials {
			if d := minDistance(s, specials[i+1:]); d < o.SpecialGap {
				t.Fatalf("seed %d: specials %d apart", seed, d)
			}
		}
		if len(b.Settlements) != len(o.Settlements) {
			t.Fatalf("seed %d: %d settlements, want %d (touching?)", seed, len(b.Settlements), len(o.Settlements))
		}
		hasBonus := map[*SettlementArea]bool{}
		for _, tl := range b.Tiles {
			if tl.Bonus == Harbor {
				if hasBonus[tl.Area] {
					t.Fatalf("seed %d: settlement with two harbors", seed)
				}
				hasBonus[tl.Area] = true
			}
		}
		parts := components(b)
		slices.Sort(parts)
		if len(parts) > 1+o.MaxIslands {
			t.Fatalf("seed %d: land parts %v", seed, parts)
		}
		for _, n := range parts[:len(parts)-1] {
			if n < o.IslandMin || n > o.IslandMax {
				t.Fatalf("seed %d: island of %d spaces", seed, n)
			}
		}
		islands[len(parts)-1]++
	}
	for n := 0; n <= o.MaxIslands; n++ {
		if islands[n] == 0 {
			t.Errorf("no map with %d islands", n)
		}
	}
}

// TestDefaultRandomLikeScotland checks that random maps get the Scotland counts.
func TestDefaultRandomLikeScotland(t *testing.T) {
	o := DefaultRandom
	if o.Spaces != 143 || o.Castles != 21 || o.Cathedrals != 8 || o.Harbors != 8 ||
		o.Food != 31 || o.Energy != 31 || o.WindCollectors != 0 || len(o.Settlements) != 19 {
		t.Fatalf("options %+v", o)
	}
}

// components returns the sizes of the connected land parts.
func components(b *Board) []int {
	var sizes []int
	seen := map[hex.Hex]bool{}
	for _, start := range b.sortedHexes() {
		if seen[start] {
			continue
		}
		n := 0
		seen[start] = true
		for stack := []hex.Hex{start}; len(stack) > 0; n++ {
			h := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, x := range h.Neighbors() {
				if b.Tiles[x] != nil && !seen[x] {
					seen[x] = true
					stack = append(stack, x)
				}
			}
		}
		sizes = append(sizes, n)
	}
	return sizes
}
