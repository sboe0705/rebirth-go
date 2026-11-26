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

You are player 1 (blue) and start; computer players (pink, black, white) follow in turn order.
A computer player places its tile on a random allowed space after a short pause.
Placed tiles have a border in the owner's colour; a controlled castle shows its icon and a ring in the owner's colour.
Not modelled: the 32 neutral tiles of the 2-player game and the castle tie-break for the winner (a tie is shared).

Food facilities go on empty or food spaces, energy facilities on empty or energy spaces, settlement tiles on settlement spaces; one tile per space.
A facility may go on a space of the other facility type only if no free empty space and no free space of its own type is left.
The hovered space shows a preview, or a red outline if the tile is not allowed there.
Placed tiles are drawn as smaller hexagons with a white icon: food `#2e6b1f` dark green with leaf, energy `#d9822b` orange with lightning bolt, settlement `#b0302a` red with 1–4 houses (its settlement symbols).
Space icons are drawn in a darker shade (55 %) of the space color.

### Scoring (`internal/board/score.go`)

The top left shows each player's points as if the game ended now:
- **placed** – scored when placing: a facility scores the size of its group (adjacent tiles of the same type and colour),
  a harbor or wind collector scores 1.
  A completed settlement ranks the players by their houses in it: medium 5/3, large 8/5, extra-large 12/8/5;
  a small settlement scores its houses for the player who placed it. On a tie the completing player loses,
  the other ties go by turn order after the completing player. A player alone in a settlement gets all ranks (medium 8, large 13, extra-large 25).
- **castles** – 5 per controlled castle. The first adjacent tile controls a castle; a player takes it over with the most adjacent tiles,
  or when tied for most with the most adjacent facilities among the tied players; otherwise it stays.
- **open settlements** – houses on the player's tiles in incomplete settlements.

The line below explains the most recent placement, e.g.
`Last placement: You +9 = 1 (harbor) + 8 (medium settlement completed)` or
`Last placement: Computer 1 +1 (food facility group of 1), castle taken over (5 at game end); You +0, castle lost`.

Not scored: cathedral mission cards, round tower bonuses.

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
| `s` | settlement space with harbor ("+1") | `#c0845a` brown | house + white plus `#ffffff` |
| `o` / `f` / `e` | empty / food / energy space with wind collector ("+1") | as `O` / `F` / `E` | white plus `#ffffff` |
| `.` / space | no space (water / off-board) | sea background `#2b5d8a` | – |

### Settlements

Adjacent settlement spaces (`S`/`s`) form one settlement, so separate settlements must not touch.
Size = number of spaces: 1 small, 2 medium, 3 large, 4 extra-large (Ireland); more than 4 is an error.
Settlement spaces are drawn with white edges (`#ffffff`).

Map rules (checked when loading and for random maps):
- harbors and wind collectors are only at the coast (space next to water),
- harbors only on settlement spaces, at most one per settlement; wind collectors never on settlement spaces,
- castles, cathedrals and round towers (`C`/`K`/`R`) are never adjacent to each other.

Not yet modelled: castle/cathedral names, 2-player blocked spaces.

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
