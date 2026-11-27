// Package ai chooses placements for computer players. It knows nothing
// about rendering.
package ai

import (
	"fmt"
	"math/rand/v2"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Level is the strength of a computer player.
type Level int

const (
	Easy   Level = iota // random allowed space
	Normal              // space with the most points right now, own mission cards included
	Hard                // for now the same as Easy
)

// Levels lists all levels from weakest to strongest.
var Levels = []Level{Easy, Normal, Hard}

func (l Level) String() string {
	switch l {
	case Easy:
		return "easy"
	case Normal:
		return "normal"
	case Hard:
		return "hard"
	}
	return "unknown"
}

// ParseLevel returns the level named s ("easy", "normal" or "hard").
func ParseLevel(s string) (Level, error) {
	for _, l := range Levels {
		if l.String() == s {
			return l, nil
		}
	}
	return 0, fmt.Errorf("unknown level %q (easy, normal or hard)", s)
}

// Choose returns the space where player p places chip c on level l.
// There must be at least one allowed space (board.CanPlaceAnywhere).
func Choose(b *board.Board, c board.Chip, p int, l Level, rng *rand.Rand) hex.Hex {
	free := b.FreeSpaces(c.Piece)
	if l == Normal {
		free = best(b, c, p, free)
	}
	return free[rng.IntN(len(free))]
}

// best returns the spaces that raise p's score (as if the game ended now)
// the most. It looks only at this placement, not at later moves. The score
// counts p's own mission cards; a card drawn by the trial placement is not
// known (board.Clone hides the deck), so it adds nothing.
func best(b *board.Board, c board.Chip, p int, free []hex.Hex) []hex.Hex {
	before := b.Score(p).Total()
	var top []hex.Hex
	most := 0
	for _, h := range free {
		t := b.Clone()
		if err := t.PlaceChip(h, c, p); err != nil {
			continue // FreeSpaces only lists allowed spaces
		}
		switch gain := t.Score(p).Total() - before; {
		case len(top) == 0 || gain > most:
			top, most = []hex.Hex{h}, gain
		case gain == most:
			top = append(top, h)
		}
	}
	return top
}
