# TerminalVolley

A simple volleyball game playable in the terminal, inspired by **Blobby
Volley**.

## Goals

- Real-time 2-player gameplay in a terminal window
- Simple physics (gravity, jumping, ball bounces)
- Net in the middle, scoring when the ball hits the ground
- Cross-platform terminal rendering (Linux/macOS/Windows via terminal library)

## Controls (current defaults)

- Player 1 (left): `A/D` move, `W` jump
- Player 2 (right): `J/L` move, `I` jump
- Quit: `Q`

## Tech

- Language: Go
- Rendering: custom fixed-size ANSI grid renderer
- Input (Linux): raw mode via `golang.org/x/sys/unix`

## Status

- [x] Basic terminal renderer (grid-based)
- [x] Simple game loop (~30 FPS)
- [x] Player movement + jumping (keyboard)
- [ ] Ball physics + collisions (ground/walls/net/players)
- [ ] Scoring + round reset/serve
- [ ] Simple UI (scoreboard)
- [ ] Cross-platform input (macOS/Windows)

## Development

### Run

```bash
go run ./cmd/terminalvolley/main.go
```

### Build

```bash
go build -o terminalvolley ./cmd/terminalvolley
./terminalvolley
```

## Notes

This project is intentionally lightweight and focuses on fun gameplay rather
than realistic simulation.
