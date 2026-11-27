# AGENTS.md

Notes for AI agents and contributors.

## Project
Go implementation of a hex board game using Ebitengine (`github.com/hajimehoshi/ebiten/v2`). Goal: playable PC version including a computer opponent. Currently 1–4 players (the user plus computer players on level easy, normal or hard, see `internal/ai`): random draws from shuffled tile supplies, placement and scoring (no mission cards).

## Rules
- Minimal and working; short explanations. Language for docs, comments, commit messages and user-facing text: **English**.
- Strictly separate game logic and rendering: `internal/hex`, `internal/board` and `internal/ai` never import Ebitengine and know nothing about pixels. Pixel math only in `internal/layout`, drawing only in `internal/render`.
- Coordinates: axial `(q, r)`, pointy-top, neighbor directions `(+1,0) (+1,-1) (0,-1) (-1,0) (-1,+1) (0,+1)`. Formulas per Red Blob Games "Hexagonal Grids".
- Board: `map[hex.Hex]*Tile`; a missing entry means no tile.
- Maps are text files in `internal/board/maps/` (format see README).
- Do not add graphics, texts or map layouts from commercial games to the repo.

## Checks
```sh
go vet ./... && go test ./...
./scripts/build.sh
```
The Linux build needs CGO and X11/GL headers (see README); Windows is cross-compiled with `CGO_ENABLED=0`.
Cover new logic in `hex`, `layout`, `board`, `ai` with tests (they run without a display).
