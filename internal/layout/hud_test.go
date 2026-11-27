package layout

import (
	"math"
	"testing"

	"github.com/sboe0705/rebirth-go/internal/hex"
)

func TestHUDRects(t *testing.T) {
	w, h := 1024.0, 768.0
	u := NewHUD(w, h).WithMissions(8)
	win := Rect{0, 0, w, h}
	all := append(u.Panels[:], u.Log, u.Missions, u.Board)
	for i, r := range all {
		if r.X < 0 || r.Y < 0 || r.X+r.W > win.W || r.Y+r.H > win.H || r.W <= 0 || r.H <= 0 {
			t.Fatalf("rect %d %v outside window", i, r)
		}
	}
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			// the board area may overlap the left panels and the missions box (Fit avoids them)
			if (i < 2 || i == 5) && j == len(all)-1 {
				continue
			}
			if all[i].Overlaps(all[j]) {
				t.Fatalf("rects %d %v and %d %v overlap", i, all[i], j, all[j])
			}
		}
	}
	if last := u.MissionRow(7); last.Y+last.H > u.Missions.Y+u.Missions.H {
		t.Fatalf("mission row %v outside the box %v", last, u.Missions)
	}
	if u.Missions.X != u.Panels[0].X || u.Missions.Y+u.Missions.H >= u.Panels[0].Y {
		t.Fatalf("missions box %v not above panel 0 %v", u.Missions, u.Panels[0])
	}
	if small := NewHUD(w, h).WithMissions(1); small.Missions.H >= u.Missions.H || small.Missions.Y+small.Missions.H != u.Missions.Y+u.Missions.H {
		t.Fatalf("box for 1 card %v should be a lower part of the box for 8 %v", small.Missions, u.Missions)
	}
	if u.Log.Y <= u.Panels[2].Y || u.Log.Y+u.Log.H >= u.Panels[3].Y {
		t.Fatalf("log %v not between right panels", u.Log)
	}
}

func hexagon(radius int) []hex.Hex {
	var hs []hex.Hex
	for q := -radius; q <= radius; q++ {
		for r := -radius; r <= radius; r++ {
			if s := -q - r; s >= -radius && s <= radius {
				hs = append(hs, hex.Hex{Q: q, R: r})
			}
		}
	}
	return hs
}

func TestFitWithoutObstacles(t *testing.T) {
	hs := hexagon(3) // corners span 7*sqrt(3) ≈ 12.1 sizes wide, 11 high
	area := Rect{10, 20, 1000, 500}
	l := Fit(hs, area, nil)
	if want := 500.0 / 11; math.Abs(l.Size-want) > 1e-9 {
		t.Fatalf("size %v, want %v", l.Size, want)
	}
	if x, y := l.HexToPixel(hex.Hex{}); math.Abs(x-510) > 1e-9 || math.Abs(y-270) > 1e-9 {
		t.Fatalf("center at %v,%v, want 510,270", x, y)
	}
}

func TestFitAvoidsRects(t *testing.T) {
	hs := hexagon(3)
	area := Rect{0, 0, 600, 600}
	free := Fit(hs, area, nil)
	block := Rect{0, 0, 140, 120} // top-left corner, partly hit by the big board
	l := Fit(hs, area, []Rect{block})
	if l.Size >= free.Size {
		t.Fatalf("size %v not smaller than %v", l.Size, free.Size)
	}
	if l.Size < free.Size/2 {
		t.Fatalf("size %v shrank too much (free %v)", l.Size, free.Size)
	}
	hw := l.Size * math.Sqrt(3) / 2
	for _, h := range hs {
		x, y := l.HexToPixel(h)
		if (Rect{x - hw, y - l.Size, 2 * hw, 2 * l.Size}).Overlaps(block) {
			t.Fatalf("hex %v overlaps the blocked rect", h)
		}
	}
}

func TestFitShiftsBeforeShrinking(t *testing.T) {
	hs := hexagon(3) // about 12.1 x 11 sizes: a wide area leaves room to the sides
	area := Rect{0, 0, 1000, 500}
	free := Fit(hs, area, nil)
	block := Rect{0, 200, 300, 100} // left of the centered board, clear of a shifted one
	l := Fit(hs, area, []Rect{block})
	if math.Abs(l.Size-free.Size) > 1e-9 {
		t.Fatalf("size %v, want %v (shifted, not shrunk)", l.Size, free.Size)
	}
	if x, _ := l.HexToPixel(hex.Hex{}); x <= 500 {
		t.Fatalf("center at x %v, want shifted right of 500", x)
	}
}
