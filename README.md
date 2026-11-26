# rebirth-go

PC-Umsetzung eines Hex-Brettspiels in Go mit [Ebitengine](https://ebitengine.org/).
Aktueller Stand: reine Anzeige eines Hex-Spielfelds aus einer Kartendatei, keine Interaktion.

## Voraussetzungen

- Go (Version siehe `go.mod`)
- Linux: CGO + C-Compiler und X11/GL-Header. Fedora:
  ```sh
  sudo dnf install gcc pkg-config mesa-libGL-devel mesa-libGLES-devel libXrandr-devel \
    libXcursor-devel libXinerama-devel libXi-devel libXxf86vm-devel alsa-lib-devel
  ```
- Windows: keine weiteren Abhängigkeiten (Ebitengine braucht dort kein CGO).

## Starten & Bauen

```sh
go run ./cmd/rebirth                 # Standardkarte "example"
go run ./cmd/rebirth -map <name>     # lädt internal/board/maps/<name>.txt

./scripts/build.sh                   # vet + test + Linux- und Windows-Build nach bin/
./scripts/build-linux.sh
./scripts/build-windows.sh           # Cross-Compile, auch von Linux aus
```

CI (`.github/workflows/build.yml`) baut bei jedem Push für Linux und Windows und lädt die Binaries als Artefakte hoch.

## Kartenformat

Karten liegen als Text unter `internal/board/maps/` und werden per `go:embed` eingebettet.
Jede Zeile ist eine Hex-Reihe, jedes Zeichen eine Spalte („odd-r“: ungerade Zeilen sind um ein halbes Feld nach rechts versetzt).
Zeilen mit `#` am Anfang sind Kommentare.

| Zeichen | Bedeutung |
|---|---|
| `G` | Grasland |
| `F` | Wald |
| `H` | Hochland |
| `W` | See |
| `S` | Siedlung |
| `.` / Leerzeichen | kein Feld |

Neue Geländetypen: `internal/board/board.go` (`Terrain`, `terrainChars`) und Farbe in `internal/render/render.go`.

## Struktur

```
cmd/rebirth/        Einstieg, ebiten.Game
internal/hex/       Axiale Koordinaten, Nachbarn, Distanz, Rounding (ohne Pixel)
internal/layout/    Hex <-> Pixel (pointy-top), Eckpunkte
internal/board/     Tile/Board-Modell, Kartenloader, eingebettete Karten
internal/render/    Zeichnen mit ebiten/v2/vector
scripts/            Build-Skripte
```
