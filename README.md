# rebirth-go

PC implementation of a hex board game in Go using [Ebitengine](https://ebitengine.org/).
Current state: hex board loaded from a map file; facilities and settlement tiles can be placed (single player, no scoring yet).

## Requirements

- Go (version see `go.mod`)
- Linux: CGO + a C compiler and X11/GL headers. Fedora:
  ```sh
  sudo dnf install gcc pkg-config mesa-libGL-devel mesa-libGLES-devel libXrandr-devel \
    libXcursor-devel libXinerama-devel libXi-devel libXxf86vm-devel alsa-lib-devel
  ```
- Windows: no further dependencies (Ebitengine needs no CGO there).

## Run & build

```sh
go run ./cmd/rebirth                 # new random map on every start
go run ./cmd/rebirth -seed 42        # reproduce a random map (seed is shown in the window title)
go run ./cmd/rebirth -map <name>     # loads internal/board/maps/<name>.txt, e.g. -map example

./scripts/build.sh                   # vet + test + Linux and Windows builds into bin/
./scripts/build-linux.sh
./scripts/build-windows.sh           # cross-compile, also works from Linux
```

## Controls

| Input | Action |
|---|---|
| Left click | place the selected tile on the space |
| `F` / `E` / `S` / right click | select food facility / energy facility / settlement tile, right click cycles |

Food facilities go on empty or food spaces, energy facilities on empty or energy spaces, settlement tiles on settlement spaces; one tile per space.
The hovered space shows a preview, or a red outline if the tile is not allowed there.
Placed tiles are drawn as smaller hexagons with a white icon: food `#2e6b1f` dark green with leaf, energy `#d9822b` orange with lightning bolt, settlement `#b0302a` red with house.
Space icons are drawn in a darker shade (55 %) of the space color.

### Scoring (single player, `internal/board/score.go`)

The top left shows the points as if the game ended now:
- **placed** – scored when placing: a facility scores the size of its group (adjacent tiles of the same type),
  a "+1" space scores 1, a completed settlement scores small = houses, medium 5+3, large 8+5, extra-large 12+8+5.
- **castles** – 5 per castle with at least one adjacent tile (no opponents, so no majority yet).
- **open settlements** – houses on tiles in incomplete settlements.

The line below explains the most recent placement, e.g.
`Last placement: +9 = 1 ("+1" space) + 8 (medium settlement completed)` or
`Last placement: +3 (food facility group of 3); castle controlled (5 at game end)`.

Not scored: cathedral mission cards, round tower bonuses. Each settlement tile counts as 1 house for now.

CI (`.github/workflows/build.yml`) builds for Linux and Windows on every push and uploads the binaries as artifacts.

## Map format

Maps are text files under `internal/board/maps/` and are embedded via `go:embed`.
Each line is a hex row, each character a column ("odd-r": odd rows are shifted right by half a tile).
Lines starting with `#` are comments.

Space kinds follow the Rebirth rules (Scotland and Ireland boards):

| Character | Meaning | Color | Icon |
|---|---|---|---|
| `O` | empty space (no symbol): food or energy facility | `#c8d69a` pale green | – |
| `F` | food space: food facility | `#6fb04a` green | leaf |
| `E` | energy space: energy facility | `#e8c53a` yellow | lightning bolt |
| `S` | settlement space | `#c0845a` brown | house |
| `C` | castle space (no tiles) | `#707078` grey | battlements |
| `K` | cathedral space (no tiles, Scotland) | `#8a5aa8` purple | cross |
| `R` | round tower space (no tiles, Ireland) | `#d8a030` amber | tower |
| `s` | settlement space with "+1" (harbour in Scotland, wind collector in Ireland) | `#c0845a` brown | house + white plus `#ffffff` |
| `.` / space | no space (water / off-board) | sea background `#2b5d8a` | – |

### Settlements

Adjacent settlement spaces (`S`/`s`) form one settlement, so separate settlements must not touch.
Size = number of spaces: 1 small, 2 medium, 3 large, 4 extra-large (Ireland); more than 4 is an error.
Settlement spaces are drawn with white edges (`#ffffff`).

Map rules (checked when loading and for random maps):
- at most one "+1" space (`s`) per settlement,
- castles, cathedrals and round towers (`C`/`K`/`R`) are never adjacent to each other.

Not yet modelled: castle/cathedral names, 2-player blocked spaces.

New space kinds: `internal/board/board.go` (`Kind`, `kindChars`) plus a color in `internal/render/render.go` and an icon in `internal/render/icons.go`.

## Random maps

`board.Random(seed, board.DefaultRandom)` (`internal/board/random.go`) builds a board of 100 spaces:
a compact random shape, 4 castles and 3 cathedrals at least 3 spaces apart,
8 settlements (sizes 1,1,1,2,2,2,3,3) that never touch each other, 3 "+1" spaces in different
settlements (preferably at the coast), and the rest randomly empty, food or energy.
The same seed always gives the same map.

## Structure

```
cmd/rebirth/        entry point, ebiten.Game
internal/hex/       axial coordinates, neighbors, distance, rounding (no pixels)
internal/layout/    hex <-> pixel (pointy-top), corner points
internal/board/     Tile/Board model, map loader, embedded maps
internal/render/    drawing with ebiten/v2/vector
scripts/            build scripts
```
