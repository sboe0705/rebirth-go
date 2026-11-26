# rebirth-go

PC implementation of a hex board game in Go using [Ebitengine](https://ebitengine.org/).
Current state: 1–4 players – you against up to 3 computer players that place their tiles at random. Tiles are drawn at random from shuffled supplies, with scoring.

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
go run ./cmd/rebirth -seed 42        # reproduce a random map, tile order and computer moves (seed is shown in the window title)
go run ./cmd/rebirth -players 3      # you + 2 computer players (default: choose 1-4 on the start screen)
go run ./cmd/rebirth -map <name>     # loads internal/board/maps/<name>.txt, e.g. -map example
go run ./cmd/rebirth -map <file>     # loads a map file from anywhere, e.g. -map ~/maps/mine.txt
go run ./cmd/rebirth -lax -map <m>   # also loads maps that break the map rules (problems are printed as warnings)

./scripts/build.sh                   # vet + test + Linux and Windows builds into bin/
./scripts/build-linux.sh
./scripts/build-windows.sh           # cross-compile, also works from Linux
```

## Controls

| Input | Action |
|---|---|
| 1–4 (start screen) | number of players: you + computer players |
| Left click | on your turn: place your tile (shown top left) on the space |
| Resize / maximize window | the board is refitted to the window |

### Tiles (`internal/board/supply.go`)

Each player has one colour set of 36 tiles: 12 food facilities, 12 energy facilities and 3 settlement tiles each with 1, 2, 3 and 4 houses.
With 4 players each set loses 3 food, 3 energy and one settlement tile each with 2, 3 and 4 houses (27 tiles).
At the start each set is shuffled and 2 tiles are set aside unused; the rest form the player's supply.
Each player holds one tile drawn at random from their supply and draws the next one after placing it.
A tile that fits nowhere on the board when it is the player's turn is discarded (the rules don't cover this).
The game ends when all supplies are empty; the points shown then are the final score.

### Players (`cmd/rebirth/main.go`)

You are player 1 (blue) and start; computer players (red, yellow, green) follow in turn order.
A computer player places its tile on a random allowed space after a short pause.
Placed tiles are filled with the owner's colour with a black border and a white icon (dark on yellow tiles); a controlled castle shows its icon in the owner's colour with a black outline.
Not modelled: the 32 neutral tiles of the 2-player game and the castle tie-break for the winner (a tie is shared).

Each player has a panel in a window corner (you bottom left, then clockwise): name, points, score parts, tiles left in supply,
and the tile in hand (yours always, a computer player's while it is placing). The player to move has a frame in their colour and a ▶ marker.
The right column lists the last 8 moves between the two right panels; messages appear at the bottom.
The board is fitted to the area left of that column without touching the panels (`internal/layout/hud.go`).
Text uses the Go fonts (`golang.org/x/image/font/gofont`, BSD license).

Food facilities go on empty or food spaces, energy facilities on empty or energy spaces, settlement tiles on settlement spaces; one tile per space.
A facility may go on a space of the other facility type only if no free empty space and no free space of its own type is left.
The hovered space shows a preview, or a red outline if the tile is not allowed there.
Placed tiles are drawn as smaller hexagons with an icon: food with leaf, energy with lightning bolt, settlement with 1–4 houses (its settlement symbols). The placement preview shows the tile in its kind colour: food `#2e6b1f` dark green, energy `#d9822b` orange, settlement `#b0302a` red.
Space icons are drawn in a darker shade (55 %) of the space color.

### Scoring (`internal/board/score.go`)

Each panel shows the player's points as if the game ended now, split into:
- **placed** – scored when placing: a facility scores the size of its group (adjacent tiles of the same type and colour),
  a harbor or wind collector scores 1.
  A completed settlement ranks the players by their houses in it: medium 5/3, large 8/5, extra-large 12/8/5;
  a small settlement scores its houses for the player who placed it. On a tie the completing player loses,
  the other ties go by turn order after the completing player. A player alone in a settlement gets all ranks (medium 8, large 13, extra-large 25).
- **castles** – 5 per controlled castle. The first adjacent tile controls a castle; a player takes it over with the most adjacent tiles,
  or when tied for most with the most adjacent facilities among the tied players; otherwise it stays.
- **open settlements** – houses on the player's tiles in incomplete settlements.

The move log explains each placement, e.g. `You +9` with `+1 harbor` and `+8 medium settlement completed`,
or `Computer 1 +1` with `+1 food facility group of 1`, `castle taken over (5 at game end)` and `You: castle lost`.

Not scored: cathedral mission cards, round tower bonuses.

CI (`.github/workflows/build.yml`) builds for Linux and Windows on every push and uploads the binaries as artifacts.

## Map format

Maps are text files under `internal/board/maps/` and are embedded via `go:embed`.
With `-map <file>` (any value containing `/` or ending in `.txt`) a map is read from a file outside the repo instead, e.g. a map you keep for yourself; it is checked against the same rules.
Each line is a hex row, each character a column ("odd-r": odd rows are shifted right by half a tile).
Lines starting with `#` are comments.

Space kinds follow the Rebirth rules (Scotland and Ireland boards):

| Character | Meaning | Color | Icon |
|---|---|---|---|
| `O` | empty space (no symbol): food or energy facility | `#d6e0a8` light green | – |
| `F` | food space: food facility | `#96c878` green | leaf |
| `E` | energy space: energy facility | `#ecd57a` light yellow | lightning bolt |
| `S` | settlement space | `#d4a682` tan | house |
| `C` | castle space (no tiles) | `#a4a4ae` grey | battlements |
| `K` | cathedral space (no tiles, Scotland) | `#b89ad0` lilac | cross |
| `R` | round tower space (no tiles, Ireland) | `#8fc8c0` teal | tower |
| `s` | settlement space with harbor ("+1") | `#d4a682` tan | house + dark plus `#404040` |
| `o` / `f` / `e` | empty / food / energy space with wind collector ("+1") | as `O` / `F` / `E` | dark plus `#404040` |
| `.` / space | no space (water / off-board) | sea background `#2b5d8a` | – |

### Settlements

Adjacent settlement spaces (`S`/`s`) form one settlement, so separate settlements must not touch.
Size = number of spaces: 1 small, 2 medium, 3 large, 4 extra-large (Ireland); more than 4 is an error.
Settlement spaces are drawn with white edges (`#ffffff`).

Map rules (checked when loading and for random maps; all violations are reported with their row and column in the file, counted from 0 without comment lines; `-lax` loads such a map anyway):
- settlements have at most 4 spaces,
- harbors and wind collectors are only at the coast (space next to water),
- harbors only on settlement spaces, at most one per settlement; wind collectors never on settlement spaces,
- castles, cathedrals and round towers (`C`/`K`/`R`) are never adjacent to each other, except a castle next to a cathedral.

### Spaces locked in 2-player games

In a 2-player game some spaces are covered with neutral tiles (SCO p. 2). Mark them in a second grid after a line `[2 players]`, with the same rows and columns as the map above it: `x` = locked, `.` or space = not locked. Any space can be locked, including castles and cathedrals.

```
.OFESOO.
.FCOSSEO
[2 players]
...x....
..x..x..
```

With 2 players a locked space is drawn as a grey tile with a cross and counts as occupied by a neutral tile (the rulebook gives no details, so this is our reading):
- nobody can place there; it doesn't count as a free space for the food/energy fallback rule,
- a settlement with a locked space is completed when its other spaces are occupied and scored by its full size; the neutral tile has no houses,
- the neutral tile belongs to no player, so it counts for no one around a castle and splits facility groups,
- a locked castle can't be controlled and scores nothing; a locked harbor or wind collector can't be used.

With 1, 3 or 4 players the grid is ignored.

Not yet modelled: castle/cathedral names.

New space kinds: `internal/board/board.go` (`Kind`, `kindChars`) plus a color in `internal/render/render.go` and an icon in `internal/render/icons.go`.

## Random maps

`board.Random(seed, board.DefaultRandom)` (`internal/board/random.go`) builds a board of 100 spaces:
a compact random shape, 4 castles and 3 cathedrals at least 3 spaces apart,
8 settlements (sizes 1,1,1,2,2,2,3,3) that never touch each other, the rest randomly empty, food or energy,
3 harbors in different settlements and 3 wind collectors, all at the coast.
The same seed always gives the same map.

## Structure

```
cmd/rebirth/        entry point, ebiten.Game, turns and computer players
internal/hex/       axial coordinates, neighbors, distance, rounding (no pixels)
internal/layout/    hex <-> pixel (pointy-top), corner points
internal/board/     Tile/Board model, map loader, embedded maps
internal/render/    drawing with ebiten/v2/vector
scripts/            build scripts
```
