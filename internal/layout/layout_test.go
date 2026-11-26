package layout

import (
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
