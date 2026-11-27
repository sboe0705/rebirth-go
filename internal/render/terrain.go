package render

import (
	"bytes"
	"embed"
	"fmt"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sboe0705/rebirth-go/internal/board"
	"github.com/sboe0705/rebirth-go/internal/hex"
	"github.com/sboe0705/rebirth-go/internal/layout"
)

// Terrain images: pointy-top hexagons with a transparent outside, named
// <kind>_<n>.png. Spaces without images keep their plain colour.
//
//go:embed terrain/*.png
var terrainFS embed.FS

var terrainNames = map[board.Kind]string{
	board.Empty:      "empty",
	board.Food:       "food",
	board.Energy:     "energy",
	board.Settlement: "settlement",
	board.Castle:     "castle",
	board.Cathedral:  "cathedral",
}

var (
	terrainOnce sync.Once
	terrainImgs map[board.Kind][]*ebiten.Image
)

func loadTerrain() {
	terrainImgs = map[board.Kind][]*ebiten.Image{}
	for k, name := range terrainNames {
		for i := 1; ; i++ {
			data, err := terrainFS.ReadFile(fmt.Sprintf("terrain/%s_%d.png", name, i))
			if err != nil {
				break
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				panic(fmt.Sprintf("terrain %s_%d: %v", name, i, err))
			}
			terrainImgs[k] = append(terrainImgs[k], ebiten.NewImageFromImage(img))
		}
	}
}

// terrainImage returns the image for space h of kind k, or nil if the kind
// has none. Variants are picked from the coordinates, so a space keeps its look.
func terrainImage(k board.Kind, h hex.Hex) *ebiten.Image {
	terrainOnce.Do(loadTerrain)
	imgs := terrainImgs[k]
	if len(imgs) == 0 {
		return nil
	}
	n := len(imgs)
	return imgs[((h.Q*7+h.R*13)%n+n)%n]
}

// drawTerrain draws img stretched over space h.
func drawTerrain(dst *ebiten.Image, l layout.Layout, h hex.Hex, img *ebiten.Image) {
	b := img.Bounds()
	x, y := l.HexToPixel(h)
	w, ht := sqrt3*l.Size, 2*l.Size
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(w/float64(b.Dx()), ht/float64(b.Dy()))
	op.GeoM.Translate(x-w/2, y-ht/2)
	dst.DrawImage(img, op)
}

const sqrt3 = 1.7320508075688772
