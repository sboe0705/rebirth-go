# rebirth-go

PC implementation of a hex board game in Go using [Ebitengine](https://ebitengine.org/).
Current state: display-only rendering of a hex board loaded from a map file, no interaction.

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
go run ./cmd/rebirth                 # default map "example"
go run ./cmd/rebirth -map <name>     # loads internal/board/maps/<name>.txt

./scripts/build.sh                   # vet + test + Linux and Windows builds into bin/
./scripts/build-linux.sh
./scripts/build-windows.sh           # cross-compile, also works from Linux
```

CI (`.github/workflows/build.yml`) builds for Linux and Windows on every push and uploads the binaries as artifacts.

## Map format

Maps are text files under `internal/board/maps/` and are embedded via `go:embed`.
Each line is a hex row, each character a column ("odd-r": odd rows are shifted right by half a tile).
Lines starting with `#` are comments.

Space kinds follow the Rebirth rules (Scotland and Ireland boards):

| Character | Meaning |
|---|---|
| `O` | empty space (no symbol): food or energy facility |
| `F` | food space: food facility |
| `E` | energy space: energy facility |
| `S` | settlement space |
| `C` | castle space (no tiles) |
| `K` | cathedral space (no tiles, Scotland) |
| `R` | round tower space (no tiles, Ireland) |
| `o` `f` `e` `s` | same as uppercase, plus "+1" (harbour in Scotland, wind collector in Ireland) |
| `.` / space | no space (water / off-board) |

Not yet modelled: settlement frames (which settlement spaces form one small/medium/large settlement), castle/cathedral names, 2-player blocked spaces.

New space kinds: `internal/board/board.go` (`Kind`, `kindChars`) plus a color in `internal/render/render.go`.

## Structure

```
cmd/rebirth/        entry point, ebiten.Game
internal/hex/       axial coordinates, neighbors, distance, rounding (no pixels)
internal/layout/    hex <-> pixel (pointy-top), corner points
internal/board/     Tile/Board model, map loader, embedded maps
internal/render/    drawing with ebiten/v2/vector
scripts/            build scripts
```
