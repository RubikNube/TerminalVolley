package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terminalvolley/config"
	"terminalvolley/internal/input"
	"terminalvolley/internal/render"
)

func keyFromConfig(field, s string) (byte, error) {
	if len(s) != 1 {
		return 0, fmt.Errorf("%s must be exactly 1 ASCII character, got %q", field, s)
	}
	b := s[0]
	if b > 127 {
		return 0, fmt.Errorf("%s must be ASCII, got %q", field, s)
	}
	return b, nil
}

func mustKeyFromConfig(field, s string) byte {
	b, err := keyFromConfig(field, s)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func resetServe(w int, toLeft bool) (bx, by, vx, vy float64) {
	// Serve from the middle to make the first hit easier.
	bx = float64(w) * 0.5
	by = 4

	// Slow, playable serve speed (cells/sec).
	vx = 6.0
	if toLeft {
		vx = -vx
	}
	vy = 0
	return
}

func main() {
	const (
		w   = 80
		h   = 24
		fps = 200
	)

	controlsCfg, err := config.LoadControls("config/controls.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "load controls:", err)
		return
	}

	// Map config controls (strings) to bytes.
	// Existing config uses single-character strings; keep same assumption.
	quitKey := mustKeyFromConfig("controls.quit", controlsCfg.Quit)
	serveLeftKey := mustKeyFromConfig("controls.serveLeft", controlsCfg.ServeLeft)
	serveRightKey := mustKeyFromConfig("controls.serveRight", controlsCfg.ServeRight)

	gameControls := Controls{
		P1Left:     mustKeyFromConfig("controls.player1.left", controlsCfg.Player1.Left),
		P1Right:    mustKeyFromConfig("controls.player1.right", controlsCfg.Player1.Right),
		P1Jump:     mustKeyFromConfig("controls.player1.jump", controlsCfg.Player1.Jump),
		P2Left:     mustKeyFromConfig("controls.player2.left", controlsCfg.Player2.Left),
		P2Right:    mustKeyFromConfig("controls.player2.right", controlsCfg.Player2.Right),
		P2Jump:     mustKeyFromConfig("controls.player2.jump", controlsCfg.Player2.Jump),
		ServeLeft:  serveLeftKey,
		ServeRight: serveRightKey,
	}
	g := NewGame(w, h, fps, gameControls)

	r := render.NewRenderer(os.Stdout, w, h)

	term, err := input.MakeRaw()
	if err != nil {
		fmt.Fprintln(os.Stderr, "raw mode:", err)
		return
	}
	defer func() { _ = term.Restore() }()

	keys := input.StartKeyReader()

	// Ensure terminal is restored on Ctrl+C.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		_ = term.Restore()
		_ = r.ShowCursor()
		fmt.Fprint(os.Stdout, "\x1b[0m\x1b[2J\x1b[H")
		os.Exit(130)
	}()

	if err := r.HideCursor(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	for range ticker.C {
		// Drain keys available this frame.
		for {
			select {
			case k, ok := <-keys:
				if !ok {
					return
				}

				// Normalize to upper for comparisons (matches previous behavior).
				kU := k
				if kU >= 'a' && kU <= 'z' {
					kU = kU - 32
				}

				if byte(kU) == quitKey {
					_ = r.ShowCursor()
					return
				}

				g.PressKey(byte(kU))

			default:
				goto keysDone
			}
		}
	keysDone:

		g.Step()

		// ---- Render ----
		frame := render.NewFrame(w, h)
		frame.Clear(' ')
		frame.DrawGround()
		frame.DrawNet()

		// Draw players/ball from the (package main) Game state.
		frame.DrawBlob(int(math.Round(g.p1x)), int(math.Round(g.p1y)))
		frame.DrawBlob(int(math.Round(g.p2x)), int(math.Round(g.p2y)))
		frame.DrawBall(int(math.Round(g.bx)), int(math.Round(g.by)))

		// Top row UI.
		p1Score, p2Score := g.Score()
		score := fmt.Sprintf("P1 %d : %d P2", p1Score, p2Score)
		for i := 0; i < len(score) && i < w; i++ {
			frame.Set(i, 0, score[i])
		}
		if g.WaitingServe() {
			msg := fmt.Sprintf("  (%c = serve left, %c = serve right)", serveLeftKey, serveRightKey)
			for i := 0; i < len(msg) && i+len(score) < w; i++ {
				frame.Set(len(score)+i, 0, msg[i])
			}
		}

		if err := r.Draw(frame); err != nil && !errors.Is(err, syscall.EPIPE) {
			fmt.Fprintln(os.Stderr, err)
			break
		}
	}

	_ = r.ShowCursor()
}
