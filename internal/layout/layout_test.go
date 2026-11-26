package layout

import (
	"math"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func TestRoundTrip(t *testing.T) {
	l := Layout{Size: 30, OriginX: 100, OriginY: 50}
	for q := -5; q <= 5; q++ {
		for r := -5; r <= 5; r++ {
			h := hex.Hex{Q: q, R: r}
			x, y := l.HexToPixel(h)
			if got := l.PixelToHex(x+3, y-4); got != h {
				t.Fatalf("%v -> %v", h, got)
			}
		}
	}
}

func TestEdgeFacesNeighbor(t *testing.T) {
	l := Layout{Size: 30, OriginX: 100, OriginY: 50}
	h := hex.Hex{Q: 2, R: -1}
	hx, hy := l.HexToPixel(h)
	for dir := 0; dir < 6; dir++ {
		a, b := l.Edge(h, dir, 1)
		nx, ny := l.HexToPixel(h.Neighbor(dir))
		mx, my := (a[0]+b[0])/2, (a[1]+b[1])/2
		if math.Abs(mx-(hx+nx)/2) > 1e-9 || math.Abs(my-(hy+ny)/2) > 1e-9 {
			t.Fatalf("dir %d: edge midpoint (%v,%v) not between centers", dir, mx, my)
		}
	}
}
