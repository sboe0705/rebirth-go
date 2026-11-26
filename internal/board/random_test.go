package board

import (
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
	for seed := uint64(0); seed < 200; seed++ {
		b, err := Random(seed, o)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if len(b.Tiles) != o.Spaces {
			t.Fatalf("seed %d: %d spaces", seed, len(b.Tiles))
		}
		count := map[Kind]int{}
		bonus := 0
		var specials []hex.Hex
		for h, tl := range b.Tiles {
			count[tl.Kind]++
			if tl.Bonus {
				bonus++
			}
			if tl.Kind > Settlement {
				specials = append(specials, h)
			}
		}
		if count[Castle] != o.Castles || count[Cathedral] != o.Cathedrals || bonus != o.Bonus {
			t.Fatalf("seed %d: counts %v, bonus %d", seed, count, bonus)
		}
		for i, s := range specials {
			if d := minDistance(s, specials[i+1:]); d < 3 {
				t.Fatalf("seed %d: specials %d apart", seed, d)
			}
		}
		if len(b.Settlements) != len(o.Settlements) {
			t.Fatalf("seed %d: %d settlements, want %d (touching?)", seed, len(b.Settlements), len(o.Settlements))
		}
		hasBonus := map[*SettlementArea]bool{}
		for _, tl := range b.Tiles {
			if tl.Bonus {
				if hasBonus[tl.Area] {
					t.Fatalf("seed %d: settlement with two \"+1\" spaces", seed)
				}
				hasBonus[tl.Area] = true
			}
		}
		if !connected(b) {
			t.Fatalf("seed %d: board not connected", seed)
		}
	}
}

func connected(b *Board) bool {
	var start hex.Hex
	for start = range b.Tiles {
		break
	}
	seen := map[hex.Hex]bool{start: true}
	for stack := []hex.Hex{start}; len(stack) > 0; {
		h := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, n := range h.Neighbors() {
			if b.Tiles[n] != nil && !seen[n] {
				seen[n] = true
				stack = append(stack, n)
			}
		}
	}
	return len(seen) == len(b.Tiles)
}
