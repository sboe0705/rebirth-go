package hex

import "testing"

func TestDistance(t *testing.T) {
	if d := Distance(Hex{0, 0}, Hex{2, -1}); d != 2 {
		t.Fatalf("got %d", d)
	}
	for _, n := range (Hex{3, -2}).Neighbors() {
		if Distance(Hex{3, -2}, n) != 1 {
			t.Fatalf("neighbor %v not at distance 1", n)
		}
	}
}

func TestRound(t *testing.T) {
	if h := Round(0.2, 0.3); h != (Hex{0, 0}) {
		t.Fatalf("got %v", h)
	}
	if h := Round(1.1, -0.9); h != (Hex{1, -1}) {
		t.Fatalf("got %v", h)
	}
}

func TestFromOffset(t *testing.T) {
	// Row 1 (odd) is indented to the right: (0,1) lies to the lower right of (0,0).
	if FromOffset(0, 1) != (Hex{0, 1}) || FromOffset(0, 2) != (Hex{-1, 2}) {
		t.Fatal("offset conversion wrong")
	}
}

func TestToOffset(t *testing.T) {
	for row := -3; row <= 3; row++ {
		for col := -3; col <= 3; col++ {
			if c, r := FromOffset(col, row).ToOffset(); c != col || r != row {
				t.Fatalf("(%d,%d) -> %d,%d", col, row, c, r)
			}
		}
	}
}
