package board

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// Cathedrals and Scotland mission cards (SCO p. 4, player aid).
// A player who places a tile next to a cathedral space puts one of their
// cathedrals on it (they stack) and draws the top mission card, at most once
// per cathedral space. Cards are secret and score at game end: 5 points if
// fulfilled, a tie for first is enough; a Cathedral Bonus scores 3.
//
// Our reading where the rules are open: a count of 0 never fulfils a card,
// the coastline "chain" is the largest connected group of own coast tiles,
// a locked cathedral gives nothing, and an empty deck gives no card.
// The "English border" card is not modelled (our maps have no England).

// MissionKind is the type of a mission card.
type MissionKind int

const (
	FoodGroups     MissionKind = iota // most own food facility groups of 3 or more
	EnergyGroups                      // most own energy facility groups of 3 or more
	LargestFood                       // largest own food facility group
	LargestEnergy                     // largest own energy facility group
	PresenceMedium                    // own tile in the most medium settlements
	PresenceLarge                     // own tile in the most large settlements
	OwnSmall                          // all spaces of the most small settlements
	OwnMedium                         // all spaces of the most medium settlements
	LongestCoast                      // largest connected group of own coast tiles
	MostHarbors                       // most own tiles on harbor spaces
	MostCathedrals                    // own cathedral on the most cathedral spaces
	MostCastles                       // most controlled castles
	CastleTiles                       // most own tiles around castle Target
	CathedralTiles                    // most own tiles around cathedral Target
	CathedralBonus                    // no condition
)

// genericMissions are the card types that name no space; each is in the deck once.
var genericMissions = []MissionKind{FoodGroups, EnergyGroups, LargestFood, LargestEnergy,
	PresenceMedium, PresenceLarge, OwnSmall, OwnMedium, LongestCoast, MostHarbors, MostCathedrals, MostCastles}

const (
	MissionPoints        = 5  // per fulfilled mission card
	CathedralBonusPoints = 3  // per Cathedral Bonus card
	DeckSize             = 32 // Scotland mission cards
	cathedralBonusCards  = 3  // our choice; the rules don't give the deck mix
)

// Mission is one mission card; Target is the castle or cathedral space of
// CastleTiles and CathedralTiles.
type Mission struct {
	Kind   MissionKind
	Target hex.Hex
}

var missionNames = map[MissionKind]string{
	FoodGroups: "Most food groups of 3+", EnergyGroups: "Most energy groups of 3+",
	LargestFood: "Largest food facility", LargestEnergy: "Largest energy facility",
	PresenceMedium: "Presence in medium settlements", PresenceLarge: "Presence in large settlements",
	OwnSmall: "Ownership of small settlements", OwnMedium: "Ownership of medium settlements",
	LongestCoast: "Longest coastline", MostHarbors: "Most harbors",
	MostCathedrals: "Most cathedrals", MostCastles: "Most castles", CathedralBonus: "Cathedral bonus",
}

// MissionName describes m, e.g. "Most harbors" or "Tiles around castle C7".
func (b *Board) MissionName(m Mission) string {
	switch m.Kind {
	case CastleTiles:
		return "Tiles around castle " + b.Label(m.Target)
	case CathedralTiles:
		return "Tiles around cathedral " + b.Label(m.Target)
	}
	return missionNames[m.Kind]
}

// Points returns what m is worth when fulfilled.
func (m Mission) Points() int {
	if m.Kind == CathedralBonus {
		return CathedralBonusPoints
	}
	return MissionPoints
}

// NewDeck builds the shuffled mission deck for b: each generic card once,
// one card per cathedral space, 3 Cathedral Bonus cards and random distinct
// castle cards until the deck has DeckSize cards (or the castles run out).
func NewDeck(b *Board, seed uint64) []Mission {
	rng := rand.New(rand.NewPCG(seed, 0xc2b2ae3d27d4eb4f))
	var deck, castles []Mission
	for _, k := range genericMissions {
		deck = append(deck, Mission{Kind: k})
	}
	for _, h := range b.sortedHexes() {
		switch b.Tiles[h].Kind {
		case Cathedral:
			deck = append(deck, Mission{CathedralTiles, h})
		case Castle:
			castles = append(castles, Mission{CastleTiles, h})
		}
	}
	for range cathedralBonusCards {
		deck = append(deck, Mission{Kind: CathedralBonus})
	}
	rng.Shuffle(len(castles), func(i, j int) { castles[i], castles[j] = castles[j], castles[i] })
	deck = append(deck, castles[:min(len(castles), max(0, DeckSize-len(deck)))]...)
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// SetDeck sets the mission deck; cards are drawn from the front.
func (b *Board) SetDeck(d []Mission) { b.deck = slices.Clone(d) }

// DeckLen returns the number of cards left in the deck.
func (b *Board) DeckLen() int { return len(b.deck) }

// Missions returns the mission cards held by player p, in drawing order.
func (b *Board) Missions(p int) []Mission { return b.missions[p] }

// CathedralStack returns the players who placed a cathedral on space h, bottom first.
func (b *Board) CathedralStack(h hex.Hex) []int { return b.cathedrals[h] }

// visitCathedral handles player p placing a tile next to cathedral space c:
// the first time p puts a cathedral on c and draws the top mission card.
func (b *Board) visitCathedral(c hex.Hex, p int) []Gain {
	if slices.Contains(b.cathedrals[c], p) {
		return nil
	}
	if b.cathedrals == nil {
		b.cathedrals = map[hex.Hex][]int{}
	}
	b.cathedrals[c] = append(slices.Clone(b.cathedrals[c]), p)
	if len(b.deck) == 0 {
		return []Gain{{p, 0, "cathedral " + b.Label(c) + " placed, no mission card left"}}
	}
	if b.missions == nil {
		b.missions = map[int][]Mission{}
	}
	b.missions[p] = append(slices.Clone(b.missions[p]), b.deck[0])
	b.deck = b.deck[1:]
	return []Gain{{p, 0, "mission card drawn at cathedral " + b.Label(c)}}
}

// Fulfilled reports whether player p fulfils mission m now: p's count is
// above 0 and no other player's is higher (a tie is enough).
func (b *Board) Fulfilled(m Mission, p int) bool {
	if m.Kind == CathedralBonus {
		return true
	}
	mine := b.missionCount(m, p)
	if mine == 0 {
		return false
	}
	for q := range max(b.Players, 1) {
		if q != p && b.missionCount(m, q) > mine {
			return false
		}
	}
	return true
}

// missionCount is player p's value for the condition of mission m.
func (b *Board) missionCount(m Mission, p int) int {
	switch m.Kind {
	case FoodGroups, EnergyGroups, LargestFood, LargestEnergy:
		f := FoodFacility
		if m.Kind == EnergyGroups || m.Kind == LargestEnergy {
			f = EnergyFacility
		}
		n, largest := 0, 0
		for _, s := range b.groupSizes(p, func(h hex.Hex) bool { return b.Tiles[h].Piece == f }) {
			if s >= 3 {
				n++
			}
			largest = max(largest, s)
		}
		if m.Kind == FoodGroups || m.Kind == EnergyGroups {
			return n
		}
		return largest
	case PresenceMedium, PresenceLarge, OwnSmall, OwnMedium:
		size := map[MissionKind]int{PresenceMedium: 2, PresenceLarge: 3, OwnSmall: 1, OwnMedium: 2}[m.Kind]
		own := m.Kind == OwnSmall || m.Kind == OwnMedium
		n := 0
		for _, a := range b.Settlements {
			if a.Size() != size {
				continue
			}
			mine := 0
			for _, h := range a.Spaces {
				if t := b.Tiles[h]; t.Piece != NoPiece && t.Owner == p {
					mine++
				}
			}
			if own && mine == size || !own && mine > 0 {
				n++
			}
		}
		return n
	case LongestCoast:
		return slices.Max(append(b.groupSizes(p, b.Coast), 0))
	case MostHarbors:
		n := 0
		for _, t := range b.Tiles {
			if t.Bonus == Harbor && t.Piece != NoPiece && t.Owner == p {
				n++
			}
		}
		return n
	case MostCathedrals:
		n := 0
		for _, s := range b.cathedrals {
			if slices.Contains(s, p) {
				n++
			}
		}
		return n
	case MostCastles:
		n := 0
		for _, o := range b.castles {
			if o == p {
				n++
			}
		}
		return n
	case CastleTiles, CathedralTiles:
		n := 0
		for _, h := range m.Target.Neighbors() {
			if t := b.Tiles[h]; t != nil && t.Piece != NoPiece && t.Owner == p {
				n++
			}
		}
		return n
	}
	return 0
}

// groupSizes returns the sizes of the connected groups of player p's tiles
// on the spaces where keep is true.
func (b *Board) groupSizes(p int, keep func(hex.Hex) bool) []int {
	mine := func(h hex.Hex) bool {
		t := b.Tiles[h]
		return t != nil && t.Piece != NoPiece && t.Owner == p && keep(h)
	}
	seen := map[hex.Hex]bool{}
	var sizes []int
	for _, start := range b.sortedHexes() {
		if seen[start] || !mine(start) {
			continue
		}
		seen[start] = true
		n := 0
		for stack := []hex.Hex{start}; len(stack) > 0; {
			c := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			n++
			for _, nb := range c.Neighbors() {
				if !seen[nb] && mine(nb) {
					seen[nb] = true
					stack = append(stack, nb)
				}
			}
		}
		sizes = append(sizes, n)
	}
	return sizes
}

// missionScore sums the points of p's fulfilled mission cards.
func (b *Board) missionScore(p int) int {
	s := 0
	for _, m := range b.missions[p] {
		if b.Fulfilled(m, p) {
			s += m.Points()
		}
	}
	return s
}

// labelSpecials numbers castles C1, C2, ... and cathedrals K1, K2, ... in map order.
func (b *Board) labelSpecials() {
	b.labels = map[hex.Hex]string{}
	n := map[Kind]int{}
	prefix := map[Kind]string{Castle: "C", Cathedral: "K"}
	for _, h := range b.sortedHexes() {
		if k := b.Tiles[h].Kind; prefix[k] != "" {
			n[k]++
			b.labels[h] = fmt.Sprintf("%s%d", prefix[k], n[k])
		}
	}
}

// Label returns the label of a castle or cathedral space, e.g. "C7", or "".
func (b *Board) Label(h hex.Hex) string {
	if b.labels == nil {
		b.labelSpecials()
	}
	return b.labels[h]
}
