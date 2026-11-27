package board

import (
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

// mv is one placement in row 0: column, piece and player.
type mv struct {
	col int
	c   Chip
	p   int
}

// playRow parses a one-row map for players and applies the moves.
func playRow(t *testing.T, text string, players int, deck []Mission, moves ...mv) *Board {
	t.Helper()
	b, err := Parse("t", text)
	if err != nil {
		t.Fatal(err)
	}
	b.Players = players
	b.SetDeck(deck)
	for _, m := range moves {
		if err := b.PlaceChip(hex.FromOffset(m.col, 0), m.c, m.p); err != nil {
			t.Fatalf("PlaceChip(col %d): %v", m.col, err)
		}
	}
	return b
}

func at(col int) hex.Hex { return hex.FromOffset(col, 0) }

func TestLabels(t *testing.T) {
	b, err := Parse("t", "CO.K\nO.C.OK")
	if err != nil {
		t.Fatal(err)
	}
	want := map[hex.Hex]string{
		hex.FromOffset(0, 0): "C1", hex.FromOffset(3, 0): "K1",
		hex.FromOffset(2, 1): "C2", hex.FromOffset(5, 1): "K2", hex.FromOffset(1, 0): "",
	}
	for h, w := range want {
		if got := b.Label(h); got != w {
			t.Errorf("Label(%s) = %q, want %q", Pos(h), got, w)
		}
	}
}

func TestNewDeck(t *testing.T) {
	b, err := Load("scotland")
	if err != nil {
		t.Fatal(err)
	}
	d := NewDeck(b, 7)
	if len(d) != DeckSize {
		t.Fatalf("deck has %d cards, want %d", len(d), DeckSize)
	}
	count, castles := map[MissionKind]int{}, map[hex.Hex]bool{}
	for _, m := range d {
		count[m.Kind]++
		if m.Kind == CastleTiles {
			castles[m.Target] = true
		}
	}
	for _, k := range genericMissions {
		if count[k] != 1 {
			t.Errorf("%s: %d cards, want 1", b.MissionName(Mission{Kind: k}), count[k])
		}
	}
	if count[CathedralTiles] != 8 || count[CathedralBonus] != 3 || count[CastleTiles] != 9 || len(castles) != 9 {
		t.Errorf("cathedral %d, bonus %d, castle %d (%d distinct); want 8, 3, 9 (9)",
			count[CathedralTiles], count[CathedralBonus], count[CastleTiles], len(castles))
	}
	if e := NewDeck(b, 7); e[0] != d[0] || e[31] != d[31] {
		t.Error("same seed gives a different deck")
	}
}

func TestCathedralVisit(t *testing.T) {
	deck := []Mission{{Kind: MostHarbors}, {Kind: CathedralBonus}, {Kind: MostCastles}}
	b := playRow(t, "OKOO", 2, deck, mv{0, food, 0}, mv{2, food, 0}, mv{3, food, 1})
	if got := b.Missions(0); len(got) != 1 || got[0].Kind != MostHarbors {
		t.Errorf("player 0 missions = %v, want only the first card", got)
	}
	if len(b.Missions(1)) != 0 || len(b.CathedralStack(at(1))) != 1 {
		t.Errorf("player 1 not next to the cathedral: missions %v, stack %v", b.Missions(1), b.CathedralStack(at(1)))
	}
	c := playRow(t, "OKO", 2, deck, mv{0, food, 0}, mv{2, food, 1})
	if s := c.CathedralStack(at(1)); len(s) != 2 || s[0] != 0 || s[1] != 1 {
		t.Errorf("stack = %v, want [0 1]", s)
	}
	if got := c.Missions(1); len(got) != 1 || got[0].Kind != CathedralBonus {
		t.Errorf("player 1 missions = %v, want the second card", got)
	}
	if c.DeckLen() != 1 {
		t.Errorf("deck has %d cards, want 1", c.DeckLen())
	}

	empty := playRow(t, "OKO", 1, nil, mv{0, food, 0})
	if len(empty.Missions(0)) != 0 || len(empty.CathedralStack(at(1))) != 1 {
		t.Error("empty deck: want a cathedral but no card")
	}

	locked, err := Parse("t", "OKO\n[2 players]\n.x.")
	if err != nil {
		t.Fatal(err)
	}
	locked.Players = 2
	locked.SetDeck(deck)
	if err := locked.PlaceChip(at(0), food, 0); err != nil {
		t.Fatal(err)
	}
	if len(locked.Missions(0)) != 0 || len(locked.CathedralStack(at(1))) != 0 {
		t.Error("locked cathedral gave a card or a cathedral")
	}
}

func TestCloneHidesDeck(t *testing.T) {
	b := playRow(t, "OKO", 1, []Mission{{Kind: CathedralBonus}})
	c := b.Clone()
	if err := c.PlaceChip(at(0), food, 0); err != nil {
		t.Fatal(err)
	}
	if len(c.Missions(0)) != 0 || b.DeckLen() != 1 || len(b.CathedralStack(at(1))) != 0 {
		t.Error("trial placement on a clone drew a card or changed the original")
	}
}

func TestFulfilled(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		m     Mission
		moves []mv
		want  [2]bool // player 0, player 1
	}{
		{"food groups lead", "OOOOO", Mission{Kind: FoodGroups},
			[]mv{{0, food, 0}, {1, food, 0}, {2, food, 0}, {4, food, 1}}, [2]bool{true, false}},
		{"food groups zero", "OOOO", Mission{Kind: FoodGroups},
			[]mv{{0, food, 0}, {1, food, 0}, {3, food, 1}}, [2]bool{false, false}},
		{"energy groups tie", "OOO.OOO", Mission{Kind: EnergyGroups},
			[]mv{{0, energy, 0}, {1, energy, 0}, {2, energy, 0}, {4, energy, 1}, {5, energy, 1}, {6, energy, 1}}, [2]bool{true, true}},
		{"largest food", "OOO.OO", Mission{Kind: LargestFood},
			[]mv{{0, food, 0}, {1, food, 0}, {4, food, 1}, {5, food, 1}, {2, food, 1}}, [2]bool{true, true}},
		{"largest energy", "OOOO", Mission{Kind: LargestEnergy},
			[]mv{{0, energy, 0}, {1, energy, 0}, {3, energy, 1}}, [2]bool{true, false}},
		{"presence medium", "SS.SS", Mission{Kind: PresenceMedium},
			[]mv{{0, house(1), 0}, {3, house(1), 0}, {4, house(1), 1}}, [2]bool{true, false}},
		{"presence large", "SSS", Mission{Kind: PresenceLarge},
			[]mv{{0, house(1), 0}, {2, house(1), 1}}, [2]bool{true, true}},
		{"own small", "S.S.S", Mission{Kind: OwnSmall},
			[]mv{{0, house(1), 0}, {2, house(1), 1}, {4, house(1), 1}}, [2]bool{false, true}},
		{"own medium needs all spaces", "SS", Mission{Kind: OwnMedium},
			[]mv{{0, house(1), 0}, {1, house(1), 1}}, [2]bool{false, false}},
		{"own medium", "SS", Mission{Kind: OwnMedium},
			[]mv{{0, house(1), 0}, {1, house(1), 0}}, [2]bool{true, false}},
		{"coast interrupted by castle", "OOCOOO", Mission{Kind: LongestCoast},
			[]mv{{0, food, 0}, {1, food, 0}, {3, food, 0}, {4, energy, 1}, {5, energy, 1}}, [2]bool{true, true}},
		{"harbors", "s.s", Mission{Kind: MostHarbors},
			[]mv{{0, house(1), 0}, {2, house(1), 0}}, [2]bool{true, false}},
		{"castles", "OCO.OCO", Mission{Kind: MostCastles},
			[]mv{{0, food, 0}, {4, food, 0}, {6, food, 1}}, [2]bool{true, false}},
		{"cathedrals", "OKO.OKO", Mission{Kind: MostCathedrals},
			[]mv{{0, food, 0}, {2, food, 1}, {6, food, 1}}, [2]bool{false, true}},
		{"around castle", "OCO", Mission{CastleTiles, at(1)},
			[]mv{{0, food, 0}, {2, food, 1}}, [2]bool{true, true}},
		{"around cathedral", "OKO", Mission{CathedralTiles, at(1)},
			[]mv{{0, food, 0}}, [2]bool{true, false}},
		{"bonus", "O", Mission{Kind: CathedralBonus}, nil, [2]bool{true, true}},
	}
	for _, c := range cases {
		b := playRow(t, c.text, 2, nil, c.moves...)
		for p := range 2 {
			if got := b.Fulfilled(c.m, p); got != c.want[p] {
				t.Errorf("%s: Fulfilled(player %d) = %v, want %v", c.name, p, got, c.want[p])
			}
		}
	}
}

func TestMissionScore(t *testing.T) {
	deck := []Mission{{Kind: MostHarbors}, {Kind: CathedralBonus}}
	b := playRow(t, "sKOOK", 2, deck, mv{0, house(1), 0}, mv{2, food, 0}, mv{3, food, 1})
	// player 0: harbor card fulfilled (5); player 1: bonus (3)
	if got := b.Score(0).Missions; got != MissionPoints {
		t.Errorf("player 0 missions = %d, want %d", got, MissionPoints)
	}
	if got := b.Score(1); got.Missions != CathedralBonusPoints || got.Total() != got.Placed+CathedralBonusPoints {
		t.Errorf("player 1 score = %+v, want missions %d", got, CathedralBonusPoints)
	}
}
