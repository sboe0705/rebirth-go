# AGENTS.md

Hinweise für KI-Agenten und Mitwirkende.

## Projekt
Go-Umsetzung eines Hex-Brettspiels mit Ebitengine (`github.com/hajimehoshi/ebiten/v2`). Ziel: spielbare PC-Version inkl. Computer-Gegner. Aktuell nur Spielfeldanzeige.

## Regeln
- Minimal und funktionierend; kurze Erklärungen. Sprache in Doku/Kommentaren: Deutsch.
- Spiellogik und Rendering strikt trennen: `internal/hex` und `internal/board` importieren nie Ebitengine und kennen keine Pixel. Pixelrechnung nur in `internal/layout`, Zeichnen nur in `internal/render`.
- Koordinaten: axial `(q, r)`, pointy-top, Nachbarrichtungen `(+1,0) (+1,-1) (0,-1) (-1,0) (-1,+1) (0,+1)`. Formeln nach Red Blob Games „Hexagonal Grids“.
- Spielfeld: `map[hex.Hex]*Tile`; fehlender Eintrag = kein Feld.
- Karten als Textdateien in `internal/board/maps/` (Format siehe README).
- Keine Grafiken, Texte oder Kartenlayouts aus kommerziellen Spielen ins Repo übernehmen.

## Prüfen
```sh
go vet ./... && go test ./...
./scripts/build.sh
```
Linux-Build braucht CGO und X11/GL-Header (siehe README); Windows wird mit `CGO_ENABLED=0` cross-kompiliert.
Neue Logik in `hex`, `layout`, `board` mit Tests abdecken (laufen ohne Display).
